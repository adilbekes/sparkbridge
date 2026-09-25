//go:build sink_buffered || all_sinks

package buffered

import (
	"context"
	"errors"
	"os"
	"testing"
)

type failingSink struct{}

func (f failingSink) Publish(context.Context, string, []byte) error { return errors.New("down") }
func (f failingSink) Close() error { return nil }

func TestProxyBuffersOnFailure(t *testing.T) {
	path := t.TempDir() + "/buffer.log"
	p := New(path, failingSink{})
	if err := p.Publish(context.Background(), "topic", []byte("abc")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("buffer file missing: %v", err)
	}
}
