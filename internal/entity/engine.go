package entity

import "fmt"

type Engine string

const (
	EnginePostgres Engine = "postgres"
)

func (e Engine) String() string {
	return string(e)
}

func NewEngine(s string) (Engine, error) {
	switch s {
	case EnginePostgres.String():
		return EnginePostgres, nil
	default:
		return "", fmt.Errorf("unknown engine: %s", s)
	}
}
