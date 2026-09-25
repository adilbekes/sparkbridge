package interfaces

import (
	"context"

	"sparkbridge/pkg/domain"
)

// InputAdapter converts external input into domain events.
type InputAdapter interface {
	Start(ctx context.Context, outChan chan<- domain.Event) error
	Stop() error
}

// OutputSink publishes encoded data to a transport target.
type OutputSink interface {
	Publish(ctx context.Context, topic string, payload []byte) error
	Close() error
}

// Encryptor encrypts payload bytes.
type Encryptor interface {
	Encrypt(payload []byte) ([]byte, error)
}

// StateStore persists bdSeq state.
type StateStore interface {
	GetBdSeq() (uint64, error)
	SetBdSeq(val uint64) error
}
