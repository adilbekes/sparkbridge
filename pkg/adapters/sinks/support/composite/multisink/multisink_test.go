//go:build sink_multisink || all_sinks

package multisink

import (
	"context"
	"testing"
)

type stubSink struct{ calls int }

func (s *stubSink) Publish(context.Context, string, []byte) error { s.calls++; return nil }
func (s *stubSink) Close() error { return nil }

func TestCompositePublishesToAll(t *testing.T) {
	a := &stubSink{}
	b := &stubSink{}
	c := New(a, b)
	if err := c.Publish(context.Background(), "topic", []byte("payload")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if a.calls != 1 || b.calls != 1 {
		t.Fatalf("unexpected call counts: %d %d", a.calls, b.calls)
	}
}
