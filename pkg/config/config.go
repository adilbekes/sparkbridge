package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config describes daemon configuration.
type Config struct {
	Bridges []Bridge `yaml:"bridges"`
}

// Bridge defines one named bridge instance.
type Bridge struct {
	Name    string        `yaml:"name"`
	Inputs  []InputConfig `yaml:"inputs"`
	Sinks   []SinkConfig  `yaml:"sinks"`
	Workers int           `yaml:"workers"`
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
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks the minimum configuration required to start the daemon.
func (c Config) Validate() error {
	if len(c.Bridges) == 0 {
		return fmt.Errorf("at least one bridge is required")
	}
	seen := make(map[string]struct{}, len(c.Bridges))
	for i, bridge := range c.Bridges {
		if strings.TrimSpace(bridge.Name) == "" {
			return fmt.Errorf("bridge %d: name is required", i)
		}
		if _, exists := seen[bridge.Name]; exists {
			return fmt.Errorf("duplicate bridge name %q", bridge.Name)
		}
		seen[bridge.Name] = struct{}{}
		if bridge.Workers < 1 {
			return fmt.Errorf("bridge %q: workers must be at least 1", bridge.Name)
		}
		if len(bridge.Inputs) == 0 {
			return fmt.Errorf("bridge %q: at least one input is required", bridge.Name)
		}
		if len(bridge.Sinks) == 0 {
			return fmt.Errorf("bridge %q: at least one sink is required", bridge.Name)
		}
		for _, input := range bridge.Inputs {
			if strings.TrimSpace(input.Kind) == "" {
				return fmt.Errorf("bridge %q: input kind is required", bridge.Name)
			}
		}
		for _, sink := range bridge.Sinks {
			if strings.TrimSpace(sink.Kind) == "" {
				return fmt.Errorf("bridge %q: sink kind is required", bridge.Name)
			}
		}
	}
	return nil
}
