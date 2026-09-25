//go:build sink_mqtt || sink_kafka || sink_nats || all_sinks

package resilience

import (
	"context"
	"errors"
	"testing"

	"sparkbridge/pkg/interfaces"
)

type failingSink struct{}

func (f failingSink) Publish(context.Context, string, []byte) error { return errors.New("fail") }
func (f failingSink) Close() error { return nil }

func TestBreakerBuffersAfterFailures(t *testing.T) {
	b := New(failingSink{}, failingSink{}, 1)
	if err := b.Publish(context.Background(), "topic", []byte("payload")); err == nil {
		t.Fatal("expected failure or buffered route")
	}
}

var _ interfaces.OutputSink = (*Breaker)(nil)
