//go:build sink_kafka || all_sinks

package kafka

import (
	"context"
	"testing"
)

func TestPublish(t *testing.T) {
	if err := New().Publish(context.Background(), "topic", []byte("abc")); err != nil {
		t.Fatalf("publish: %v", err)
}
}
