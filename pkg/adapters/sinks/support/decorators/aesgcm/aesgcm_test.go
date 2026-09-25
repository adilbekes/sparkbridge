package wrappers

import (
	"context"
	"testing"
)

type noopSink struct{}

func (noopSink) Publish(context.Context, string, []byte) error { return nil }
func (noopSink) Close() error { return nil }

func TestAESGCMWraps(t *testing.T) {
	a := NewAESGCM(make([]byte, 32), noopSink{})
	if err := a.Publish(context.Background(), "topic", []byte("payload")); err != nil {
		t.Fatalf("publish: %v", err)
	}
}
