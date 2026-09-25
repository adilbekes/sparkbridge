package pipeline

import (
	"context"
	"sync"
	"testing"
	"time"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

type mockSink struct {
	mu       sync.Mutex
	received []EncodedMessage
}

func (m *mockSink) Publish(_ context.Context, topic string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.received = append(m.received, EncodedMessage{Topic: topic, Payload: payload, Timestamp: time.Now()})
	return nil
}

func (m *mockSink) Close() error { return nil }

type mockInput struct{}

func (m *mockInput) Start(ctx context.Context, outChan chan<- domain.Event) error {
	select {
	case outChan <- domain.Event{GroupID: "g1", NodeID: "n1", MsgType: domain.MessageTypeNDATA, Timestamp: time.Now()}:
	case <-ctx.Done():
	}
	return nil
}

func (m *mockInput) Stop() error { return nil }

var _ interfaces.InputAdapter = (*mockInput)(nil)
var _ interfaces.OutputSink = (*mockSink)(nil)

func TestPipelineProcessesEvents(t *testing.T) {
	sink := &mockSink{}
	p := New(2, func(evt domain.Event) (EncodedMessage, error) {
		return EncodedMessage{Topic: "spBv1.0/g1/NDATA/n1", Payload: []byte(evt.NodeID), Timestamp: time.Now()}, nil
	}, sink)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Run(ctx)

	input := &mockInput{}
	if err := input.Start(ctx, p.Events()); err != nil {
		t.Fatalf("start input: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	p.Close()

	sink.mu.Lock()
	defer sink.mu.Unlock()
	if len(sink.received) == 0 {
		t.Fatalf("expected at least one encoded message")
	}
}

func TestPipelineHandlesGracefulClose(t *testing.T) {
	sink := &mockSink{}
	p := New(1, func(evt domain.Event) (EncodedMessage, error) {
		return EncodedMessage{Topic: "spBv1.0/g1/NDATA/n1", Payload: []byte(evt.NodeID), Timestamp: time.Now()}, nil
	}, sink)

	ctx, cancel := context.WithCancel(context.Background())
	p.Run(ctx)
	cancel()
	p.Close()

	sink.mu.Lock()
	defer sink.mu.Unlock()
	_ = len(sink.received)
}
