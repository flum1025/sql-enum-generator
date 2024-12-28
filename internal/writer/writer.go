package writer

import "github.com/flum1025/sql-enum-generator/internal/parser"

type Writer interface {
	Write(tables []parser.Table) error
}
