//go:build input_json || input_mqtt || all_inputs

package e2e

import (
	"context"
	"testing"
	"time"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/pipeline"
	"sparkbridge/pkg/pipeline/middleware"
	"sparkbridge/pkg/pipeline/ratelimit"
)

func TestEndToEnd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan pipeline.EncodedMessage, 1)
	pool := pipeline.NewWorkerPool(1, func(evt domain.Event) (pipeline.EncodedMessage, error) {
		return pipeline.EncodedMessage{Topic: evt.GroupID, Payload: []byte(evt.NodeID), Timestamp: time.Now()}, nil
	}, ratelimit.New(10, time.Millisecond), middleware.Validation(middleware.Enrichment(func(evt domain.Event) (domain.Event, error) { return evt, nil })))
	pool.Run(ctx)
	pool.In() <- domain.Event{GroupID: "g1", NodeID: "n1", MsgType: domain.MessageTypeNDATA}
	select {
	case msg := <-pool.Out():
		out <- msg
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
	cancel()
	pool.Close()
	if len(out) == 0 {
		t.Fatal("expected output")
	}
}
