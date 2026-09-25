package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := writeConfig(t, `bridges:
  - name: default
    workers: 2
    inputs:
      - name: stdin
        kind: json
        config: stdin
    sinks:
      - name: console
        kind: stdout
        config: ""
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Bridges) != 1 || cfg.Bridges[0].Name != "default" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestCheckedInConfigLoads(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "sparkbridge.yaml")
	if _, err := Load(path); err != nil {
		t.Fatalf("load checked-in config: %v", err)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	path := writeConfig(t, "bridges: []\nunknown: true\n")
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "field unknown") {
		t.Fatalf("expected unknown-field error, got %v", err)
	}
}

func TestValidateRejectsInvalidBridges(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "empty", cfg: Config{}},
		{name: "missing name", cfg: Config{Bridges: []Bridge{{Workers: 1, Inputs: []InputConfig{{Kind: "json"}}, Sinks: []SinkConfig{{Kind: "stdout"}}}}}},
		{name: "zero workers", cfg: Config{Bridges: []Bridge{{Name: "a", Inputs: []InputConfig{{Kind: "json"}}, Sinks: []SinkConfig{{Kind: "stdout"}}}}}},
		{name: "duplicate", cfg: Config{Bridges: []Bridge{{Name: "a", Workers: 1, Inputs: []InputConfig{{Kind: "json"}}, Sinks: []SinkConfig{{Kind: "stdout"}}}, {Name: "a", Workers: 1, Inputs: []InputConfig{{Kind: "json"}}, Sinks: []SinkConfig{{Kind: "stdout"}}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sparkbridge.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
