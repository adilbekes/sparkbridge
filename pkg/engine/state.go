package engine

import (
	"context"
	"sync"

	"sparkbridge/pkg/interfaces"
)

// SequenceManager manages the in-memory Sparkplug seq counter.
type SequenceManager struct {
	mu  sync.Mutex
	seq uint8
}

// NewSequenceManager creates a sequence manager with seq initialized to zero.
func NewSequenceManager() *SequenceManager { return &SequenceManager{} }

// NextSeq increments seq and wraps at 255 -> 0.
func (m *SequenceManager) NextSeq() uint8 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	return m.seq
}

// ResetSeq sets seq back to zero.
func (m *SequenceManager) ResetSeq() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq = 0
}

// Current returns the current seq value.
func (m *SequenceManager) Current() uint8 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seq
}

// InitializeBDSeq loads, increments, wraps, and persists bdSeq.
func InitializeBDSeq(ctx context.Context, store interfaces.StateStore) (uint64, error) {
	if store == nil {
		return 0, nil
	}
	current, err := store.GetBdSeq()
	if err != nil {
		return 0, err
	}
	next := (current + 1) % 256
	if err := store.SetBdSeq(next); err != nil {
		return 0, err
	}
	_ = ctx
	return next, nil
}
