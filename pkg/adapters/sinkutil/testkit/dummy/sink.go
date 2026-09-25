//go:build sink_dummy

package dummy

import (
	"context"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// DummySink is a no-op sink used for registry verification.
type DummySink struct{}

// Publish accepts and discards payloads.
func (d *DummySink) Publish(_ context.Context, _ string, _ []byte) error { return nil }

// Close closes the sink.
func (d *DummySink) Close() error { return nil }

// NewDummySink constructs the sink.
func NewDummySink() (interfaces.OutputSink, error) { return &DummySink{}, nil }

func init() { registry.RegisterSink("dummy", NewDummySink) }
