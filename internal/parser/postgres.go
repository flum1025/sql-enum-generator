package parser

import (
	"strconv"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
	"github.com/samber/lo"
)

var _ Parser = &PostgresParser{}

type PostgresParser struct {
}

func (p *PostgresParser) Parse(
	source string,
) ([]Table, error) {
	tree, err := pg_query.Parse(source)
	if err != nil {
		return nil, err
	}

	tables := make([]Table, 0, len(tree.Stmts))

	for _, stmt := range tree.Stmts {
		insertStmt := stmt.Stmt.GetInsertStmt()
		if insertStmt == nil {
			continue
		}

		tableName := insertStmt.Relation.Relname

		cols := lo.Map(
			lo.Filter(
				insertStmt.Cols,
				func(col *pg_query.Node, _ int) bool {
					return col.GetResTarget() != nil
				},
			),
			func(col *pg_query.Node, _ int) string {
				return col.GetResTarget().Name
			},
		)

		type columnValue struct {
			value string
			typ   RowType
		}

		lists := lo.Map(
			insertStmt.SelectStmt.GetSelectStmt().GetValuesLists(),
			func(list *pg_query.Node, _ int) []columnValue {
				return lo.Map(
					list.GetList().Items,
					func(item *pg_query.Node, _ int) columnValue {
						node := item.GetAConst()
						if node == nil {
							panic("not a const node")
						}

						switch val := node.GetVal().(type) {
						case *pg_query.A_Const_Ival:
							return columnValue{
								value: strconv.FormatInt(int64(val.Ival.Ival), 10),
								typ:   RowTypeInteger,
							}
						case *pg_query.A_Const_Fval:
							return columnValue{
								value: val.Fval.Fval,
								typ:   RowTypeString,
							}
						case *pg_query.A_Const_Boolval:
							return columnValue{
								value: strconv.FormatBool(val.Boolval.Boolval),
								typ:   RowTypeString,
							}
						case *pg_query.A_Const_Sval:
							return columnValue{
								value: val.Sval.Sval,
								typ:   RowTypeString,
							}
						case *pg_query.A_Const_Bsval:
							return columnValue{
								value: string(val.Bsval.Bsval),
								typ:   RowTypeString,
							}
						default:
							return columnValue{
								value: "",
								typ:   RowTypeUnknown,
							}
						}
					},
				)
			},
		)

		rows := lo.Map(
			lists,
			func(list []columnValue, _ int) Row {
				return lo.Map(
					lo.Zip2(cols, list),
					func(tuple lo.Tuple2[string, columnValue], _ int) Column {
						return Column{
							Name:  tuple.A,
							Type:  tuple.B.typ,
							Value: tuple.B.value,
						}
					},
				)
			},
		)

		tables = append(tables, Table{
			Name: tableName,
			Rows: rows,
		})
	}

	return tables, nil
}

func (p *PostgresParser) ParseDefinitions(
	source string,
) (TableDefinitions, error) {
	tree, err := pg_query.Parse(stripMetaCommands(source))
	if err != nil {
		return nil, err
	}

	definitions := make(TableDefinitions, 0, len(tree.Stmts))
	alteredPrimaryKeys := make(map[string][]string)

	for _, stmt := range tree.Stmts {
		if alterStmt := stmt.Stmt.GetAlterTableStmt(); alterStmt != nil {
			for _, cmd := range alterStmt.Cmds {
				alterCmd := cmd.GetAlterTableCmd()
				if alterCmd == nil || alterCmd.Subtype != pg_query.AlterTableType_AT_AddConstraint {
					continue
				}

				if isPrimaryKeyConstraint(alterCmd.Def) {
					alteredPrimaryKeys[alterStmt.Relation.Relname] = append(
						alteredPrimaryKeys[alterStmt.Relation.Relname],
						constraintKeys(alterCmd.Def.GetConstraint())...,
					)
				}
			}

			continue
		}

		createStmt := stmt.Stmt.GetCreateStmt()
		if createStmt == nil {
			continue
		}

		columns := make([]ColumnDefinition, 0, len(createStmt.TableElts))
		primaryKeys := make([]string, 0)

		for _, elt := range createStmt.TableElts {
			if columnDef := elt.GetColumnDef(); columnDef != nil {
				columns = append(columns, ColumnDefinition{
					Name:     columnDef.Colname,
					TypeName: typeName(columnDef.TypeName),
				})

				if lo.SomeBy(columnDef.Constraints, isPrimaryKeyConstraint) {
					primaryKeys = append(primaryKeys, columnDef.Colname)
				}

				continue
			}

			if isPrimaryKeyConstraint(elt) {
				primaryKeys = append(primaryKeys, constraintKeys(elt.GetConstraint())...)
			}
		}

		definitions = append(definitions, TableDefinition{
			Name:        createStmt.Relation.Relname,
			Columns:     columns,
			PrimaryKeys: primaryKeys,
		})
	}

	for i, definition := range definitions {
		definitions[i].PrimaryKeys = append(definition.PrimaryKeys, alteredPrimaryKeys[definition.Name]...)
	}

	return definitions, nil
}

func stripMetaCommands(source string) string {
	lines := strings.Split(source, "\n")

	return strings.Join(
		lo.Map(lines, func(line string, _ int) string {
			if strings.HasPrefix(line, `\`) {
				return ""
			}

			return line
		}),
		"\n",
	)
}

func constraintKeys(constraint *pg_query.Constraint) []string {
	return lo.Map(
		constraint.GetKeys(),
		func(key *pg_query.Node, _ int) string {
			return key.GetString_().Sval
		},
	)
}

func isPrimaryKeyConstraint(node *pg_query.Node) bool {
	constraint := node.GetConstraint()

	return constraint != nil && constraint.Contype == pg_query.ConstrType_CONSTR_PRIMARY
}

func typeName(typeName *pg_query.TypeName) string {
	names := lo.Map(
		typeName.GetNames(),
		func(name *pg_query.Node, _ int) string {
			return name.GetString_().Sval
		},
	)

	return lo.LastOrEmpty(names)
}
