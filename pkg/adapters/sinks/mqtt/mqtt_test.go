//go:build sink_mqtt || all_sinks

package mqtt

import (
	"context"
	"testing"
	"time"
)

func TestSinkPublishAndClose(t *testing.T) {
	sink := New(Config{BrokerURL: "tcp://localhost:1883", ClientID: "test-client", TopicPrefix: "sparkbridge"})
	if err := sink.Publish(context.Background(), "topic", []byte("payload")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestSinkHonorsContextCancellation(t *testing.T) {
	sink := New(Config{BrokerURL: "tcp://localhost:1883", ClientID: "test-client"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sink.Publish(ctx, "topic", []byte("payload")); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestSinkCloseIsIdempotent(t *testing.T) {
	sink := New(Config{BrokerURL: "tcp://localhost:1883", ClientID: "test-client"})
	_ = sink.Publish(context.Background(), "topic", []byte("payload"))
	if err := sink.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
	_ = time.Now()
}
