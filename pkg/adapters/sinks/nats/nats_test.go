//go:build sink_nats || all_sinks

package nats

import (
	"context"
	"testing"
)

func TestPublish(t *testing.T) {
	if err := New().Publish(context.Background(), "topic", []byte("abc")); err != nil {
		t.Fatalf("publish: %v", err)
}
}
