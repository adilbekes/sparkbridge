//go:build sink_mqtt || all_sinks

package mqtt

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/spbproto"
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
		called = evt.MsgType == domain.MessageTypeNCMD && ctx != nil
		return nil
	})
	dt := uint32(spbproto.DataType_Boolean)
	payload, err := proto.Marshal(&spbproto.Payload{Metrics: []*spbproto.Payload_Metric{{Name: proto.String("Node Control/Rebirth"), Datatype: &dt, Value: &spbproto.Payload_Metric_BooleanValue{BooleanValue: true}}}})
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}
	sink.onMessage(nil, fakeMessage{topic: "spBv1.0/group/NCMD/node", payload: payload})
	if !called {
		t.Fatal("expected command router to be called")
	}
}

func TestDBIRTHSubscribesToDeviceCommandTopic(t *testing.T) {
	sink := New(Config{})
	if err := sink.Publish(context.Background(), "spBv1.0/group/DBIRTH/node/device", nil); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if _, ok := sink.subscriptions["spBv1.0/group/DCMD/node/device"]; !ok {
		t.Fatal("expected device command subscription")
	}
}

func TestDecodeDeviceCommand(t *testing.T) {
	dt := uint32(spbproto.DataType_Int64)
	payload, err := proto.Marshal(&spbproto.Payload{Metrics: []*spbproto.Payload_Metric{{Name: proto.String("setpoint"), Datatype: &dt, Value: &spbproto.Payload_Metric_LongValue{LongValue: 42}}}})
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}
	event, err := decodeCommand("spBv1.0/group/DCMD/node/device", payload)
	if err != nil {
		t.Fatalf("decode command: %v", err)
	}
	if event.MsgType != domain.MessageTypeDCMD || event.DeviceID != "device" || len(event.Metrics) != 1 || event.Metrics[0].Value != int64(42) {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestDecodeCommandRejectsInvalidInput(t *testing.T) {
	if _, err := decodeCommand("spBv1.0/group/NDATA/node", nil); err == nil {
		t.Fatal("expected invalid-topic error")
	}
	if _, err := decodeCommand("spBv1.0/group/NCMD/node", []byte("invalid")); err == nil {
		t.Fatal("expected invalid-protobuf error")
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
func (m fakeMessage) MessageID() uint16 { return 0 }
func (m fakeMessage) Payload() []byte   { return m.payload }
func (m fakeMessage) Ack()              {}
