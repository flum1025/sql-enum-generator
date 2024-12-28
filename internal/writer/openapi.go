package writer

import (
	"fmt"
	"os"

	"github.com/flum1025/sql-enum-generator/internal/entity"
	"github.com/flum1025/sql-enum-generator/internal/parser"
	"github.com/samber/lo"
	"github.com/swaggest/openapi-go/openapi3"
)

var _ Writer = &OpenAPIWriter{}

type OpenAPIWriter struct {
	targetSchemaTables []entity.SchemaTable
	outputPath         string
}

func NewOpenAPIWriter(
	targetSchemaTables []entity.SchemaTable,
	outputPath string,
) *OpenAPIWriter {
	return &OpenAPIWriter{
		targetSchemaTables: targetSchemaTables,
		outputPath:         outputPath,
	}
}

func (w *OpenAPIWriter) Write(
	tables []parser.Table,
) error {
	nameToTable := lo.SliceToMap(
		tables,
		func(table parser.Table) (string, parser.Table) {
			return table.Name, table
		},
	)

	spec := openapi3.Spec{
		Openapi: "3.0.0",
		Info: openapi3.Info{
			Title: "Code generated by github.com/flum1025/sql-enum-generator, DO NOT EDIT.",
		},
	}

	spec.Components = &openapi3.Components{
		Schemas: &openapi3.ComponentsSchemas{
			MapOfSchemaOrRefValues: lo.SliceToMap(
				w.targetSchemaTables,
				func(table entity.SchemaTable) (string, openapi3.SchemaOrRef) {
					def := nameToTable[table.Name]

					return table.Name, openapi3.SchemaOrRef{
						Schema: &openapi3.Schema{
							Type: lo.ToPtr(openapi3.SchemaTypeString),
							Enum: lo.Map(
								def.Rows,
								func(row parser.Row, _ int) any {
									return fmt.Sprintf("%s", row[table.Key])
								},
							),
						},
					}
				},
			),
		},
	}

	bytes, err := spec.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal bytes: %w", err)
	}

	if err := os.WriteFile(w.outputPath, bytes, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
