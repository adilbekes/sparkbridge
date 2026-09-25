//go:build sink_stdout || all_sinks

package stdout

import (
	"context"
	"fmt"
	"os"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// Sink writes payloads to stdout.
type Sink struct{}

// New creates a new stdout sink.
func New() *Sink { return &Sink{} }

// Publish prints the payload to stdout.
func (s *Sink) Publish(ctx context.Context, topic string, payload []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := fmt.Fprintf(os.Stdout, "%s: %x\n", topic, payload)
		return err
	}
}

// Close closes the sink.
func (s *Sink) Close() error { return nil }

func init() { registry.RegisterSink("stdout", func() (interfaces.OutputSink, error) { return New(), nil }) }
