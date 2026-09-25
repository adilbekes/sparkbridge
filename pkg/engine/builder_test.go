package engine

import (
	"context"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/spbproto"
)

type memoryStore struct{ bdSeq uint64 }

func (m *memoryStore) GetBdSeq() (uint64, error) { return m.bdSeq, nil }
func (m *memoryStore) SetBdSeq(val uint64) error { m.bdSeq = val; return nil }

type noopEncryptor struct{}
func (noopEncryptor) Encrypt(payload []byte) ([]byte, error) { return payload, nil }

func TestSequenceWraps(t *testing.T) {
	seq := NewSequenceManager()
	for i := 0; i < 255; i++ {
		seq.NextSeq()
	}
	if got := seq.NextSeq(); got != 0 {
		t.Fatalf("expected wrap to 0, got %d", got)
	}
}

func TestNBIRTHResetsSeq(t *testing.T) {
	e := NewEngine(&memoryStore{}, noopEncryptor{}, DefaultTypeResolver{})
	e.seq.seq = 42
	_, _, err := e.BuildPayload(context.Background(), domain.Event{GroupID: "g", NodeID: "n", MsgType: domain.MessageTypeNBIRTH, Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if got := e.seq.Current(); got != 0 {
		t.Fatalf("expected seq reset to 0, got %d", got)
	}
}

func TestPayloadEncoding(t *testing.T) {
	e := NewEngine(&memoryStore{}, noopEncryptor{}, DefaultTypeResolver{})
	e.bdSeq = 7
	topic, payload, err := e.BuildPayload(context.Background(), domain.Event{
		GroupID: "group",
		NodeID:  "node",
		MsgType: domain.MessageTypeNDATA,
		Metrics: []domain.Metric{{Name: "temp", Value: int64(42), Timestamp: time.Unix(0, 0)}},
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if topic != "spBv1.0/group/NDATA/node" {
		t.Fatalf("unexpected topic %q", topic)
	}
	if len(payload) == 0 {
		t.Fatalf("expected payload bytes")
	}
	var decoded spbproto.Payload
	if err := proto.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if decoded.Timestamp == nil {
		t.Fatalf("expected timestamp")
	}
	if decoded.Seq == nil || *decoded.Seq != 1 {
		t.Fatalf("expected seq 1, got %#v", decoded.Seq)
	}
	if len(decoded.Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(decoded.Metrics))
	}
	metric := decoded.Metrics[0]
	if metric.Name == nil || *metric.Name != "temp" {
		t.Fatalf("unexpected metric name %#v", metric.Name)
	}
	if metric.Datatype == nil || *metric.Datatype != spbproto.DataType_DATA_TYPE_INT64 {
		t.Fatalf("unexpected datatype %#v", metric.Datatype)
	}
	if metric.IntValue == nil || *metric.IntValue != 42 {
		t.Fatalf("unexpected int value %#v", metric.IntValue)
	}
}
