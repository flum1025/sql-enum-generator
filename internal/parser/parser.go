package parser

import (
	"fmt"
	"strings"

	"github.com/flum1025/sql-enum-generator/internal/entity"
	"github.com/samber/lo"
)

type Parser interface {
	Parse(
		source string,
	) ([]Table, error)
	ParseDefinitions(
		source string,
	) (TableDefinitions, error)
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

type IDType string

const (
	IDTypeUUID  IDType = "uuid"
	IDTypeInt32 IDType = "int32"
	IDTypeInt64 IDType = "int64"
)

type ID struct {
	Name  string
	Table string
	Type  IDType
}

type ColumnDefinition struct {
	Name     string
	TypeName string
}

type TableDefinition struct {
	Name        string
	PrimaryKeys []ColumnDefinition
}

type TableDefinitions []TableDefinition

func (t TableDefinitions) ToID(def entity.SchemaTable) (ID, error) {
	table, ok := lo.Find(t, func(table TableDefinition) bool {
		return table.Name == def.Name
	})
	if !ok {
		return ID{}, fmt.Errorf("table not found: %s", def.Name)
	}

	if len(table.PrimaryKeys) != 1 {
		return ID{}, fmt.Errorf("table %s must have exactly one primary key column, got %d", def.Name, len(table.PrimaryKeys))
	}

	pk := table.PrimaryKeys[0]

	idType, err := toIDType(pk.TypeName)
	if err != nil {
		return ID{}, fmt.Errorf("table %s column %s: %w", def.Name, pk.Name, err)
	}

	return ID{
		Name:  def.IDType,
		Table: def.Name,
		Type:  idType,
	}, nil
}

func toIDType(typeName string) (IDType, error) {
	switch strings.ToLower(typeName) {
	case "uuid":
		return IDTypeUUID, nil
	case "serial", "serial4", "integer", "int", "int4":
		return IDTypeInt32, nil
	case "bigserial", "serial8", "bigint", "int8":
		return IDTypeInt64, nil
	default:
		return "", fmt.Errorf("unsupported primary key type: %s", typeName)
	}
}
