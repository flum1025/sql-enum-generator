package app

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samber/lo"

	"github.com/flum1025/sql-enum-generator/internal/entity"
)

type QueryGenerator struct {
	engine      entity.Engine
	config      entity.Config
	databaseURL string
	outputPath  string
}

type QueryGeneratorOption struct {
	Engine      entity.Engine
	ConfigPath  string
	DatabaseURL string
	OutputPath  string
}

func NewQueryGenerator(
	option QueryGeneratorOption,
) (*QueryGenerator, error) {
	config, err := entity.NewConfigFromFile(option.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("new config from file: %w", err)
	}

	return &QueryGenerator{
		engine:      option.Engine,
		config:      config,
		databaseURL: option.DatabaseURL,
		outputPath:  option.OutputPath,
	}, nil
}

func (a *QueryGenerator) Run() error {
	db, err := sql.Open("pgx", a.databaseURL)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	defer db.Close()

	statements := make([]string, 0, len(a.config.Tables))

	for _, table := range a.config.Tables {
		records, err := a.extract(db, table.Name)
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}

		statement := a.transform(table.Name, records)

		statements = append(statements, statement)
	}

	if err := a.load(statements); err != nil {
		return fmt.Errorf("load: %w", err)
	}

	return nil
}

type records struct {
	Columns []string
	Values  [][]any
}

func (a *QueryGenerator) extract(
	db *sql.DB,
	tableName string,
) (records, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return records{}, fmt.Errorf("query: %w", err)
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return records{}, fmt.Errorf("columns: %w", err)
	}

	values := make([][]any, 0)

	for rows.Next() {
		tmpValues := make([]any, len(columns))
		tmpValuesPtr := lo.Map(tmpValues, func(_ any, i int) any {
			return &tmpValues[i]
		})

		if err := rows.Scan(tmpValuesPtr...); err != nil {
			return records{}, fmt.Errorf("scan: %w", err)
		}

		values = append(values, tmpValues)
	}

	if err = rows.Err(); err != nil {
		return records{}, fmt.Errorf("rows err: %w", err)
	}

	return records{
		Columns: columns,
		Values:  values,
	}, nil
}

func (a *QueryGenerator) transform(
	tableName string,
	records records,
) string {
	relStatement := fmt.Sprintf("(%s)", strings.Join(records.Columns, ", "))

	selectStatements := lo.Map(
		records.Values,
		func(row []any, _ int) string {
			values := lo.Map(row, func(value any, i int) string {
				switch value := value.(type) {
				case string:
					return fmt.Sprintf("'%s'", value)
				case int64:
					return fmt.Sprintf("%d", value)
				case time.Time:
					return fmt.Sprintf("'%s'", value.Format(time.RFC3339))
				case nil:
					return "NULL"
				default:
					panic(fmt.Sprintf("unexpected type: %v", reflect.TypeOf(value)))
				}
			})

			return fmt.Sprintf("(%s)", strings.Join(values, ", "))
		},
	)

	return fmt.Sprintf(
		`INSERT INTO %s
	%s
VALUES
	%s;
		`,
		tableName,
		relStatement,
		strings.Join(selectStatements, ",\n  "),
	)
}

func (a *QueryGenerator) load(
	statements []string,
) error {
	if err := os.WriteFile(
		a.outputPath,
		[]byte(strings.Join(statements, "\n\n")),
		0644,
	); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
