package entity

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string        `yaml:"version"`
	Tables  []SchemaTable `yaml:"tables"`
}

type SchemaTable struct {
	Name  string `yaml:"name"`
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
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

	return config, nil
}
