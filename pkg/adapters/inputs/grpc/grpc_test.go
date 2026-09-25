//go:build input_grpc || all_inputs

package grpc

import (
	"testing"

	"sparkbridge/pkg/adapters/registry"
)

func TestGRPCInputRegisters(t *testing.T) {
	if _, err := registry.Input("grpc"); err != nil {
		t.Fatalf("expected grpc adapter to be registered: %v", err)
	}
}
