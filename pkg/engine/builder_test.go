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
	if metric.Datatype == nil || *metric.Datatype != uint32(spbproto.DataType_Int64) {
		t.Fatalf("unexpected datatype %#v", metric.Datatype)
	}
	if metric.GetLongValue() != 42 {
		t.Fatalf("unexpected long value %d", metric.GetLongValue())
	}
}

func TestPayloadEncodesBytes(t *testing.T) {
	e := NewEngine(&memoryStore{}, noopEncryptor{}, DefaultTypeResolver{})
	_, payload, err := e.BuildPayload(context.Background(), domain.Event{
		GroupID: "group", NodeID: "node", MsgType: domain.MessageTypeNDATA,
		Metrics: []domain.Metric{{Name: "raw", Value: []byte{1, 2, 3}}},
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	var decoded spbproto.Payload
	if err := proto.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Metrics) != 1 || string(decoded.Metrics[0].GetBytesValue()) != string([]byte{1, 2, 3}) {
		t.Fatalf("unexpected bytes metric: %#v", decoded.Metrics)
	}
}

func TestInitializeBDSeqStoresFullSessionCounter(t *testing.T) {
	store := &memoryStore{bdSeq: 255}
	got, err := InitializeBDSeq(context.Background(), store)
	if err != nil {
		t.Fatalf("init bdSeq: %v", err)
	}
	if got != 256 {
		t.Fatalf("expected full session counter 256, got %d", got)
	}
	if store.bdSeq != 256 {
		t.Fatalf("expected persisted counter 256, got %d", store.bdSeq)
	}
}

func TestInitializeBDSeqRejectsCancellationAndOverflow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := InitializeBDSeq(ctx, &memoryStore{}); err == nil {
		t.Fatal("expected cancellation error")
	}
	if _, err := InitializeBDSeq(context.Background(), &memoryStore{bdSeq: ^uint64(0)}); err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestPayloadEmitsModuloBdSeq(t *testing.T) {
	e := NewEngine(&memoryStore{}, noopEncryptor{}, DefaultTypeResolver{})
	e.bdSeq = 513
	_, payload, err := e.BuildPayload(context.Background(), domain.Event{
		GroupID:   "group",
		NodeID:    "node",
		MsgType:   domain.MessageTypeNBIRTH,
		Timestamp: time.Unix(0, 0),
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	var decoded spbproto.Payload
	if err := proto.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	found := false
	for _, metric := range decoded.Metrics {
		if metric.Name != nil && *metric.Name == "bdSeq" {
			found = true
			if metric.GetLongValue() != 1 {
				t.Fatalf("expected wire bdSeq 1, got %d", metric.GetLongValue())
			}
		}
	}
	if !found {
		t.Fatal("expected bdSeq metric in payload")
	}
}

func TestIsRebirthCommand(t *testing.T) {
	evt := domain.Event{MsgType: domain.MessageTypeNCMD, Metrics: []domain.Metric{{Name: "Node Control/Rebirth", Value: true}}}
	if !IsRebirthCommand(evt) {
		t.Fatal("expected rebirth command to be detected")
	}
	if IsRebirthCommand(domain.Event{MsgType: domain.MessageTypeDCMD}) {
		t.Fatal("did not expect rebirth command on device command")
	}
}

func TestCommandRouterRebirthEmitsNBIRTH(t *testing.T) {
	store := &memoryStore{}
	e := NewEngine(store, noopEncryptor{}, DefaultTypeResolver{})
	_, _, err := e.BuildPayload(context.Background(), domain.Event{GroupID: "g", NodeID: "n", MsgType: domain.MessageTypeNBIRTH, Metrics: []domain.Metric{{Name: "active", Value: int64(7)}}})
	if err != nil {
		t.Fatalf("build initial birth: %v", err)
	}
	out := make(chan domain.Event, 1)
	router := NewCommandRouter(e, out)
	if err := router.Handle(context.Background(), domain.Event{GroupID: "g", NodeID: "n", MsgType: domain.MessageTypeNCMD, Metrics: []domain.Metric{{Name: "Node Control/Rebirth", Value: true}}}); err != nil {
		t.Fatalf("handle rebirth: %v", err)
	}
	select {
	case evt := <-out:
		if evt.MsgType != domain.MessageTypeNBIRTH {
			t.Fatalf("expected NBIRTH event, got %s", evt.MsgType)
		}
		if len(evt.Metrics) != 1 || evt.Metrics[0].Name != "active" {
			t.Fatalf("expected active birth metrics, got %#v", evt.Metrics)
		}
		if e.State() != NodeStateBirth {
			t.Fatalf("expected birth state, got %s", e.State())
		}
	case <-time.After(time.Second):
		t.Fatal("expected rebirth event")
	}
}
