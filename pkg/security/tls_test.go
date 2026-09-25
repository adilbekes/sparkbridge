package security

import "testing"

func TestLoadEmptyConfig(t *testing.T) {
	if cfg, err := Load(Config{}); err != nil || cfg == nil {
		t.Fatalf("load: %v", err)
	}
}
