//go:build sink_dummy || all_sinks

package dummy

import (
	"testing"

	"sparkbridge/pkg/adapters/registry"
)

func TestDummySinkRegisters(t *testing.T) {
	if _, err := registry.Sink("dummy"); err != nil {
		t.Fatalf("expected dummy sink to be registered: %v", err)
	}
}
