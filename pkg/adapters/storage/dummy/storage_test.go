//go:build storage_dummy || all_storage

package dummy

import (
	"testing"

	"sparkbridge/pkg/adapters/registry"
)

func TestDummyStorageRegisters(t *testing.T) {
	if _, err := registry.Storage("dummy"); err != nil {
		t.Fatalf("expected dummy storage to be registered: %v", err)
	}
}
