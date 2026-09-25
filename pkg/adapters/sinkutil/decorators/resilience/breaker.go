package resilience

import (
	"context"
	"sync"
	"time"

	"sparkbridge/pkg/interfaces"
)

// State represents the circuit breaker state.
type State string

const (
	Closed   State = "closed"
	Open     State = "open"
	HalfOpen State = "half-open"
)

// Breaker wraps a sink with circuit breaking.
type Breaker struct {
	mu        sync.Mutex
	state     State
	failures  int
	threshold int
	openedAt  time.Time
	peer      interfaces.OutputSink
	buffered  interfaces.OutputSink
}

// New creates a breaker.
func New(peer interfaces.OutputSink, bufferedSink interfaces.OutputSink, threshold int) *Breaker {
	if threshold < 1 {
		threshold = 3
	}
	return &Breaker{state: Closed, threshold: threshold, peer: peer, buffered: bufferedSink}
}

// Publish forwards or buffers depending on state.
func (b *Breaker) Publish(ctx context.Context, topic string, payload []byte) error {
	b.mu.Lock()
	state := b.state
	if state == Open && time.Since(b.openedAt) > time.Second {
		b.state = HalfOpen
		state = HalfOpen
	}
	b.mu.Unlock()

	if state == Open {
		if b.buffered != nil {
			return b.buffered.Publish(ctx, topic, payload)
		}
		return nil
	}

	if b.peer == nil {
		return nil
	}
	if err := b.peer.Publish(ctx, topic, payload); err != nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.failures++
		if b.failures >= b.threshold {
			b.state = Open
			b.openedAt = time.Now()
		}
		if b.buffered != nil {
			return b.buffered.Publish(ctx, topic, payload)
		}
		return err
	}

	b.mu.Lock()
	b.failures = 0
	b.state = Closed
	b.mu.Unlock()
	return nil
}

// Close closes wrapped sinks.
func (b *Breaker) Close() error {
	if b.peer != nil {
		_ = b.peer.Close()
	}
	if b.buffered != nil {
		return b.buffered.Close()
	}
	return nil
}
