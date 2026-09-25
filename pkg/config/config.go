package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config describes daemon configuration.
type Config struct {
	Bridges []Bridge `yaml:"bridges"`
}

// Bridge defines one named bridge instance.
type Bridge struct {
	Name    string         `yaml:"name"`
	Inputs  []InputConfig  `yaml:"inputs"`
	Sinks   []SinkConfig   `yaml:"sinks"`
	Workers int            `yaml:"workers"`
}

// InputConfig describes an input source.
type InputConfig struct {
	Name   string `yaml:"name"`
	Kind   string `yaml:"kind"`
	Config string `yaml:"config"`
}

// SinkConfig describes an output sink.
type SinkConfig struct {
	Name   string `yaml:"name"`
	Kind   string `yaml:"kind"`
	Config string `yaml:"config"`
}

// Load parses a YAML config file.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
