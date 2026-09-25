//go:build sink_kafka || all_sinks

package kafka

import (
	"context"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// Sink publishes to Kafka topics.
type Sink struct{}

// New creates a new Kafka sink.
func New() *Sink { return &Sink{} }

// Publish accepts a payload for a Kafka topic.
func (s *Sink) Publish(ctx context.Context, topic string, payload []byte) error {
	_ = ctx
	_ = topic
	_ = payload
	return nil
}

// Close closes the sink.
func (s *Sink) Close() error { return nil }

func init() { registry.RegisterSink("kafka", func() (interfaces.OutputSink, error) { return New(), nil }) }
