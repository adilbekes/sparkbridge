//go:build sink_nats || all_sinks

package nats

import (
	"context"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// Sink publishes to NATS subjects.
type Sink struct{}

// New creates a new NATS sink.
func New() *Sink { return &Sink{} }

// Publish accepts a payload for a NATS subject.
func (s *Sink) Publish(ctx context.Context, topic string, payload []byte) error {
	_ = ctx
	_ = topic
	_ = payload
	return nil
}

// Close closes the sink.
func (s *Sink) Close() error { return nil }

func init() { registry.RegisterSink("nats", func() (interfaces.OutputSink, error) { return New(), nil }) }
