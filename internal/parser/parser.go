package parser

import (
	"github.com/flum1025/sql-enum-generator/internal/entity"
	"github.com/samber/lo"
)

type Parser interface {
	Parse(
		source string,
	) ([]Table, error)
}

type RowType string

const (
	RowTypeString  RowType = "string"
	RowTypeInteger RowType = "integer"
	RowTypeUnknown RowType = "unknown"
)

type Column struct {
	Name  string
	Value string
	Type  RowType
}

type Row []Column

func (r Row) GetByName(name string) *Column {
	col, ok := lo.Find(r, func(c Column) bool {
		return c.Name == name
	})
	if !ok {
		return nil
	}

	return &col
}

type Rows []Row

func (r Rows) ToEnum(def entity.SchemaTable) Enum {
	valueTypes := lo.Uniq(lo.Map(
		r,
		func(row Row, _ int) RowType {
			return row.GetByName(def.Value).Type
		},
	))
	if len(valueTypes) != 1 {
		panic("multiple value types found")
	}

	keys := lo.Map(
		r,
		func(row Row, _ int) string {
			return row.GetByName(def.Key).Value
		},
	)

	values := lo.Map(
		r,
		func(row Row, _ int) string {
			return row.GetByName(def.Value).Value
		},
	)

	return Enum{
		Name:      def.Name,
		ValueType: lo.FirstOrEmpty(valueTypes),
		Keys:      keys,
		Values:    values,
	}
}

type Enum struct {
	Name      string
	ValueType RowType
	Keys      []string
	Values    []string
}

func (e Enum) IsEmpty() bool {
	return len(e.Values) == 0
}

type Table struct {
	Name string
	Rows Rows
}
