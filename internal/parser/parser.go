package parser

type Parser interface {
	Parse(
		source string,
	) ([]Table, error)
}

type Row map[string]string

type Table struct {
	Name string
	Rows []Row
}
