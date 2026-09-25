//go:build input_json || all_inputs

package json

import (
	"testing"
	"time"
)

func TestDecodeEvent(t *testing.T) {
	data := []byte(`{"group_id":"g1","node_id":"n1","device_id":"d1","msg_type":"NDATA","timestamp":"2026-09-25T00:00:00Z","metrics":[{"name":"temp","value":42,"timestamp":"2026-09-25T00:01:00Z"}]}`)
	event, err := DecodeEvent(data)
	if err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if event.GroupID != "g1" || event.NodeID != "n1" || event.DeviceID != "d1" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.MsgType != "NDATA" {
		t.Fatalf("unexpected msg type: %s", event.MsgType)
	}
	if event.Timestamp.IsZero() || event.Timestamp.UTC() != time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected timestamp: %v", event.Timestamp)
	}
	if len(event.Metrics) != 1 || event.Metrics[0].Name != "temp" {
		t.Fatalf("unexpected metrics: %#v", event.Metrics)
	}
}

func TestDecodeEventInvalidJSON(t *testing.T) {
	if _, err := DecodeEvent([]byte(`{"group_id":`)); err == nil {
		t.Fatal("expected error")
	}
}
