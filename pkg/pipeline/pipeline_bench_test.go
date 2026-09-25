package pipeline

import (
	"context"
	"testing"
	"time"

	"sparkbridge/pkg/domain"
)

func BenchmarkPipeline(b *testing.B) {
	p := New(4, func(evt domain.Event) (EncodedMessage, error) {
		return EncodedMessage{Topic: evt.GroupID, Payload: []byte(evt.NodeID), Timestamp: time.Now()}, nil
	}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Run(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Events() <- domain.Event{GroupID: "g1", NodeID: "n1", MsgType: domain.MessageTypeNDATA, Timestamp: time.Now()}
	}
	cancel()
	p.Close()
}
