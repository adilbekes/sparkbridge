//go:build input_grpc || all_inputs

package grpc

import (
	"context"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// GRPCInput is a minimal sample input adapter.
type GRPCInput struct{}

// Start emits a single test event.
func (g *GRPCInput) Start(_ context.Context, outChan chan<- domain.Event) error {
	if outChan != nil {
		outChan <- domain.Event{MsgType: domain.MessageTypeNDATA}
	}
	return nil
}

// Stop shuts down the adapter.
func (g *GRPCInput) Stop() error { return nil }

// NewGRPCInput constructs the input adapter.
func NewGRPCInput() (interfaces.InputAdapter, error) { return &GRPCInput{}, nil }

func init() {
	registry.RegisterInput("grpc", NewGRPCInput)
}
