package parser

import (
	"strconv"

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

		lists := lo.Map(
			insertStmt.SelectStmt.GetSelectStmt().GetValuesLists(),
			func(list *pg_query.Node, _ int) []string {
				return lo.Map(
					list.GetList().Items,
					func(item *pg_query.Node, _ int) string {
						node := item.GetAConst()
						if node == nil {
							panic("not a const node")
						}

						switch val := node.GetVal().(type) {
						case *pg_query.A_Const_Ival:
							return strconv.FormatInt(int64(val.Ival.Ival), 10)
						case *pg_query.A_Const_Fval:
							return val.Fval.Fval
						case *pg_query.A_Const_Boolval:
							return strconv.FormatBool(val.Boolval.Boolval)
						case *pg_query.A_Const_Sval:
							return val.Sval.Sval
						case *pg_query.A_Const_Bsval:
							return val.Bsval.Bsval
						default:
							panic("unknown const node")
						}
					},
				)
			},
		)

		rows := lo.Map(
			lists,
			func(list []string, _ int) Row {
				return lo.FromEntries(
					lo.Map(
						lo.Zip2(cols, list),
						func(tuple lo.Tuple2[string, string], _ int) lo.Entry[string, string] {
							return lo.Entry[string, string]{
								Key:   tuple.A,
								Value: tuple.B,
							}
						},
					),
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
