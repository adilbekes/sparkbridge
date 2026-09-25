//go:build crypto_dummy || all_crypto

package dummy

import (
	"testing"

	"sparkbridge/pkg/adapters/registry"
)

func TestDummyCryptoRegisters(t *testing.T) {
	if _, err := registry.Crypto("dummy"); err != nil {
		t.Fatalf("expected dummy crypto to be registered: %v", err)
	}
}
