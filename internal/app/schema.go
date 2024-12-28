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
	outputPath string
}

type SchemaGeneratorOption struct {
	Engine     entity.Engine
	ConfigPath string
	SourcePath string
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

	return &SchemaGenerator{
		engine:     option.Engine,
		config:     config,
		source:     source,
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

	writer := writer.NewOpenAPIWriter(a.config.Tables, a.outputPath)

	tables, err := _parser.Parse(a.source)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	writer.Write(tables)

	return nil
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
