//go:build sink_buffered || all_sinks

package buffered

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// Proxy buffers payloads on disk when the wrapped sink fails.
type Proxy struct {
	mu   sync.Mutex
	path string
	peer interfaces.OutputSink
}

// New creates a buffered proxy.
func New(path string, peer interfaces.OutputSink) *Proxy { return &Proxy{path: path, peer: peer} }

// Publish forwards to the peer or persists the payload to disk.
func (p *Proxy) Publish(ctx context.Context, topic string, payload []byte) error {
	if p.peer != nil {
		if err := p.peer.Publish(ctx, topic, payload); err == nil {
			return nil
		}
	}
	return p.buffer(topic, payload)
}

// Close closes the wrapped sink.
func (p *Proxy) Close() error {
	if p.peer != nil {
		return p.peer.Close()
	}
	return nil
}

func (p *Proxy) buffer(topic string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.path == "" {
		p.path = filepath.Join(os.TempDir(), "sparkbridge-buffered-sink.log")
	}
	f, err := os.OpenFile(p.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s:%x\n", topic, payload)
	return err
}

func init() { registry.RegisterSink("buffered", func() (interfaces.OutputSink, error) { return New("", nil), nil }) }
