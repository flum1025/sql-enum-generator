package entity

import (
	"fmt"
	"os"

	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string        `yaml:"version"`
	Tables  []SchemaTable `yaml:"tables"`
}

func (c Config) HasIDType() bool {
	return lo.SomeBy(c.Tables, SchemaTable.HasIDType)
}

func (c Config) EnumTables() []SchemaTable {
	return lo.Filter(c.Tables, func(table SchemaTable, _ int) bool {
		return table.HasEnum()
	})
}

func (c Config) Validate() error {
	for _, table := range c.Tables {
		if err := table.Validate(); err != nil {
			return fmt.Errorf("table %s: %w", table.Name, err)
		}
	}

	return nil
}

type SchemaTable struct {
	Name   string `yaml:"name"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	IDType string `yaml:"id_type"`
}

func (t SchemaTable) HasEnum() bool {
	return t.Key != "" && t.Value != ""
}

func (t SchemaTable) HasIDType() bool {
	return t.IDType != ""
}

func (t SchemaTable) Validate() error {
	if (t.Key == "") != (t.Value == "") {
		return fmt.Errorf("key and value must be specified together")
	}

	if !t.HasEnum() && !t.HasIDType() {
		return fmt.Errorf("key and value or id_type is required")
	}

	return nil
}

func NewConfigFromFile(
	path string,
) (Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var config Config

	if err := yaml.UnmarshalStrict(bytes, &config); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}
