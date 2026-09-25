package manager

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/config"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

type testInput struct {
	stopped atomic.Bool
}

func (i *testInput) Start(ctx context.Context, out chan<- domain.Event) error {
	select {
	case out <- domain.Event{GroupID: "group", NodeID: "node", MsgType: domain.MessageTypeNDATA, Metrics: []domain.Metric{{Name: "value", Value: int64(1)}}}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (i *testInput) Stop() error {
	i.stopped.Store(true)
	return nil
}

type testSink struct {
	messages chan string
	closed   atomic.Bool
}

func (s *testSink) Publish(ctx context.Context, topic string, _ []byte) error {
	select {
	case s.messages <- topic:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *testSink) Close() error {
	s.closed.Store(true)
	return nil
}

func TestManagerStartsBridgeAndReusesInstance(t *testing.T) {
	input := &testInput{}
	sink := &testSink{messages: make(chan string, 1)}
	registry.RegisterInput("manager_test_input", func() (interfaces.InputAdapter, error) { return input, nil })
	registry.RegisterSink("manager_test_sink", func() (interfaces.OutputSink, error) { return sink, nil })

	mgr := New()
	cfg := config.Bridge{
		Name:    "bridge",
		Workers: 1,
		Inputs:  []config.InputConfig{{Kind: "manager_test_input"}},
		Sinks:   []config.SinkConfig{{Kind: "manager_test_sink"}},
	}
	first, err := mgr.Get(context.Background(), cfg)
	if err != nil {
		t.Fatalf("get bridge: %v", err)
	}
	second, err := mgr.Get(context.Background(), cfg)
	if err != nil {
		t.Fatalf("get existing bridge: %v", err)
	}
	if first != second {
		t.Fatal("expected multiton instance reuse")
	}

	select {
	case topic := <-sink.messages:
		if topic != "spBv1.0/group/NDATA/node" {
			t.Fatalf("unexpected topic %q", topic)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for sink publish")
	}

	mgr.Close()
	if !input.stopped.Load() || !sink.closed.Load() {
		t.Fatal("expected adapters to close")
	}
}

func TestManagerRejectsMissingAdapter(t *testing.T) {
	mgr := New()
	_, err := mgr.Get(context.Background(), config.Bridge{Name: "bridge", Workers: 1, Inputs: []config.InputConfig{{Kind: "missing"}}})
	if err == nil {
		t.Fatal("expected missing adapter error")
	}
}
