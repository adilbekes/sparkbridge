//go:build sink_multisink || all_sinks

package multisink

import (
	"context"
	"sync"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// Composite fans out to multiple sinks.
type Composite struct {
	mu    sync.RWMutex
	sinks []interfaces.OutputSink
}

// New creates a composite sink.
func New(sinks ...interfaces.OutputSink) *Composite { return &Composite{sinks: sinks} }

// Add appends a sink to the fan-out set.
func (c *Composite) Add(sink interfaces.OutputSink) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sinks = append(c.sinks, sink)
}

// Publish forwards the payload to each child sink.
func (c *Composite) Publish(ctx context.Context, topic string, payload []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, sink := range c.sinks {
		if sink == nil {
			continue
		}
		if err := sink.Publish(ctx, topic, payload); err != nil {
			return err
		}
	}
	return nil
}

// Close closes each child sink.
func (c *Composite) Close() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, sink := range c.sinks {
		if sink == nil {
			continue
		}
		if err := sink.Close(); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	registry.RegisterSink("multisink", func() (interfaces.OutputSink, error) { return New(), nil })
}
