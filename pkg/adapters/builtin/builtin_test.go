//go:build all_inputs && all_sinks && all_storage && all_crypto

package builtin

import (
	"testing"

	"sparkbridge/pkg/adapters/registry"
)

func TestAllCapabilityTagsRegisterAdapters(t *testing.T) {
	for _, name := range []string{"json", "mqtt", "http", "grpc"} {
		if _, err := registry.Input(name); err != nil {
			t.Errorf("input %q: %v", name, err)
		}
	}
	for _, name := range []string{"mqtt", "kafka", "nats", "stdout", "buffered", "multisink"} {
		if _, err := registry.Sink(name); err != nil {
			t.Errorf("sink %q: %v", name, err)
		}
	}
	if _, err := registry.Storage("dummy"); err != nil {
		t.Errorf("storage dummy: %v", err)
	}
	if _, err := registry.Crypto("dummy"); err != nil {
		t.Errorf("crypto dummy: %v", err)
	}
}
