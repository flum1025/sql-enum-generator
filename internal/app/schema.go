package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flum1025/sql-enum-generator/internal/entity"
	"github.com/flum1025/sql-enum-generator/internal/parser"
	"github.com/flum1025/sql-enum-generator/internal/writer"
)

type SchemaGenerator struct {
	engine     entity.Engine
	config     entity.Config
	source     string
	schema     string
	outputPath string
}

type SchemaGeneratorOption struct {
	Engine     entity.Engine
	ConfigPath string
	SourcePath string
	SchemaPath string
	OutputPath string
}

func NewSchemaGenerator(
	option SchemaGeneratorOption,
) (*SchemaGenerator, error) {
	config, err := entity.NewConfigFromFile(option.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("new config from file: %w", err)
	}

	source, err := loadSources(option.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("load sources: %w", err)
	}

	var schema string

	if config.HasIDType() {
		if option.SchemaPath == "" {
			return nil, fmt.Errorf("schema path is required when id_type is configured")
		}

		schema, err = loadSources(option.SchemaPath)
		if err != nil {
			return nil, fmt.Errorf("load schemas: %w", err)
		}
	}

	return &SchemaGenerator{
		engine:     option.Engine,
		config:     config,
		source:     source,
		schema:     schema,
		outputPath: option.OutputPath,
	}, nil
}

func (a *SchemaGenerator) Run() error {
	_parser := func() parser.Parser {
		if a.engine == entity.EnginePostgres {
			return &parser.PostgresParser{}
		}

		return nil
	}()
	if _parser == nil {
		return fmt.Errorf("unknown engine: %s", a.engine)
	}

	writer := writer.NewOpenAPIWriter(a.config.EnumTables(), a.outputPath)

	tables, err := _parser.Parse(a.source)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	ids, err := a.parseIDs(_parser)
	if err != nil {
		return fmt.Errorf("parse ids: %w", err)
	}

	if err := writer.Write(tables, ids); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (a *SchemaGenerator) parseIDs(
	_parser parser.Parser,
) ([]parser.ID, error) {
	if !a.config.HasIDType() {
		return nil, nil
	}

	definitions, err := _parser.ParseDefinitions(a.schema)
	if err != nil {
		return nil, fmt.Errorf("parse definitions: %w", err)
	}

	ids := make([]parser.ID, 0, len(a.config.Tables))

	for _, def := range a.config.Tables {
		if !def.HasIDType() {
			continue
		}

		id, err := definitions.ToID(def)
		if err != nil {
			return nil, fmt.Errorf("to id: %w", err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func loadSources(
	path string,
) (string, error) {
	files, err := filepath.Glob(path)
	if err != nil {
		return "", fmt.Errorf("glob: %w", err)
	}

	sources := make([]string, 0, len(files))

	for _, file := range files {
		bytes, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}

		sources = append(sources, string(bytes))
	}

	return strings.Join(sources, "\n"), nil
}
