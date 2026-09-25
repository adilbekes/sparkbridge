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


// CommandSink extends OutputSink with MQTT subscription lifecycle hooks.
type CommandSink interface {
	OutputSink
	Subscribe(ctx context.Context, topic string) error
	Unsubscribe(ctx context.Context, topic string) error
}

// Encryptor encrypts payload bytes.
type Encryptor interface {
	Encrypt(payload []byte) ([]byte, error)
}

// StateStore persists the full bdSeq session counter state.
type StateStore interface {
	GetBdSeq() (uint64, error)
	SetBdSeq(val uint64) error
}

// NormalizeBdSeq clamps bdSeq into the Sparkplug-required 0..255 range for wire emission.
func NormalizeBdSeq(val uint64) uint64 { return val % 256 }
