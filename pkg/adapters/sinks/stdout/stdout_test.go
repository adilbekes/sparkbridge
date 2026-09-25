//go:build sink_stdout || all_sinks

package stdout

import (
	"context"
	"testing"
)

func TestPublish(t *testing.T) {
	if err := New().Publish(context.Background(), "topic", []byte("abc")); err != nil {
		t.Fatalf("publish: %v", err)
	}
}
