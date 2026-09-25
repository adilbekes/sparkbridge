//go:build storage_dummy || all_storage

package dummy

import (
	"context"
	"sync"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// DummyStorage is an in-memory bdSeq session counter store.
type DummyStorage struct {
	mu    sync.Mutex
	bdSeq uint64
}

// GetBdSeq returns the current bdSeq value.
func (d *DummyStorage) GetBdSeq() (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.bdSeq, nil
}

// SetBdSeq stores the full bdSeq session counter value.
func (d *DummyStorage) SetBdSeq(val uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.bdSeq = val
	return nil
}

// NewDummyStorage constructs the store.
func NewDummyStorage() (interfaces.StateStore, error) { return &DummyStorage{}, nil }

func init() { registry.RegisterStorage("dummy", NewDummyStorage) }

var _ = context.Background
