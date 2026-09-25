//go:build sink_mqtt || all_sinks

package mqtt

import (
	"context"
	"sparkbridge/pkg/domain"
	"testing"
)

func TestNBIRTHTriggersSubscriptions(t *testing.T) {
	sink := New(Config{})
	if err := sink.Publish(context.Background(), "spBv1.0/group/NBIRTH/node", []byte("{}")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if _, ok := sink.subscriptions["spBv1.0/group/NCMD/node"]; !ok {
		t.Fatal("expected node command subscription")
	}
	if _, ok := sink.subscriptions["spBv1.0/group/DCMD/node/+"]; !ok {
		t.Fatal("expected device command wildcard subscription")
	}
}

func TestNDEATHTriggersUnsubscribe(t *testing.T) {
	sink := New(Config{})
	_ = sink.Subscribe(context.Background(), "spBv1.0/group/NCMD/node")
	_ = sink.Subscribe(context.Background(), "spBv1.0/group/DCMD/node/+")
	if err := sink.Publish(context.Background(), "spBv1.0/group/NDEATH/node", []byte("{}")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if _, ok := sink.subscriptions["spBv1.0/group/NCMD/node"]; ok {
		t.Fatal("expected node command subscription to be removed")
	}
	if _, ok := sink.subscriptions["spBv1.0/group/DCMD/node/+"]; ok {
		t.Fatal("expected device command wildcard subscription to be removed")
	}
}

func TestCommandRouterReceivesRebirth(t *testing.T) {
	sink := New(Config{})
	called := false
	sink.WithCommandRouter(func(ctx context.Context, evt domain.Event) error {
		called = evt.MsgType == domain.MessageTypeNBIRTH || ctx != nil
		return nil
	})
	sink.onMessage(nil, fakeMessage{topic: "spBv1.0/group/NCMD/node", payload: []byte(`{"msg_type":"NCMD","metrics":[{"name":"Node Control/Rebirth","value":"true"}]}`)})
	if !called {
		t.Fatal("expected command router to be called")
	}
}

func TestCommandSinkRebirthPublishPath(t *testing.T) {
	sink := New(Config{})
	sink.WithCommandSink(sink)
	evt := domain.Event{GroupID: "group", NodeID: "node", MsgType: domain.MessageTypeNCMD, Metrics: []domain.Metric{{Name: "Node Control/Rebirth", Value: true}}}
	if !isRebirthMetric(evt) {
		t.Fatal("expected rebirth metric")
	}
	if got := triggerTopic(domain.Event{GroupID: "group", NodeID: "node", MsgType: domain.MessageTypeNBIRTH}); got != "spBv1.0/group/NBIRTH/node" {
		t.Fatalf("unexpected topic %q", got)
	}
}

type fakeMessage struct {
	topic   string
	payload []byte
}

func (m fakeMessage) Duplicate() bool   { return false }
func (m fakeMessage) Qos() byte         { return 0 }
func (m fakeMessage) Retained() bool    { return false }
func (m fakeMessage) Topic() string     { return m.topic }
func (m fakeMessage) MessageID() uint16  { return 0 }
func (m fakeMessage) Payload() []byte    { return m.payload }
func (m fakeMessage) Ack()              {}
