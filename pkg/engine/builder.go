package engine

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
	"sparkbridge/pkg/spbproto"
)

// TypeResolver resolves native values to Sparkplug payload data types.
type TypeResolver interface {
	InferDataType(val any) (spbproto.DataType, error)
}

// DefaultTypeResolver maps primitive Go values to Sparkplug payload data types.
type DefaultTypeResolver struct{}

// InferDataType resolves primitive values to Sparkplug data types.
func (DefaultTypeResolver) InferDataType(val any) (spbproto.DataType, error) {
	switch val.(type) {
	case int32:
		return spbproto.DataType_Int32, nil
	case int64:
		return spbproto.DataType_Int64, nil
	case float32:
		return spbproto.DataType_Float, nil
	case float64:
		return spbproto.DataType_Double, nil
	case string:
		return spbproto.DataType_String, nil
	case bool:
		return spbproto.DataType_Boolean, nil
	case []byte:
		return spbproto.DataType_Bytes, nil
	default:
		return spbproto.DataType_Unknown, fmt.Errorf("unsupported payload value %T", val)
	}
}

// Engine ties together sequence, state, encryption, and type resolution.
type Engine struct {
	mu       sync.RWMutex
	seq      *SequenceManager
	store    interfaces.StateStore
	encrypt  interfaces.Encryptor
	resolver TypeResolver
	bdSeq    uint64
	birth    *domain.Event
	state    *StateMachine
}

// NewEngine creates a new engine.
func NewEngine(store interfaces.StateStore, encrypt interfaces.Encryptor, resolver TypeResolver) *Engine {
	if resolver == nil {
		resolver = DefaultTypeResolver{}
	}
	return &Engine{seq: NewSequenceManager(), store: store, encrypt: encrypt, resolver: resolver, state: NewStateMachine()}
}

// State returns the current node lifecycle state.
func (e *Engine) State() NodeState { return e.state.Current() }

// Transition applies a node lifecycle transition.
func (e *Engine) Transition(messageType domain.MessageType) { e.state.Transition(messageType) }

// InitBDSeq loads and persists bdSeq for the node.
func (e *Engine) InitBDSeq(ctx context.Context) error {
	bdSeq, err := InitializeBDSeq(ctx, e.store)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.bdSeq = bdSeq
	e.mu.Unlock()
	return nil
}

// BirthEvent returns a snapshot of the last successfully encoded NBIRTH event.
func (e *Engine) BirthEvent() (domain.Event, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.birth == nil {
		return domain.Event{}, false
	}
	return e.birth.Clone(), true
}

// BuildPayload builds topic and payload bytes for the supplied event.
func (e *Engine) BuildPayload(ctx context.Context, event domain.Event) (string, []byte, error) {
	if err := ctx.Err(); err != nil {
		return "", nil, err
	}
	topic := TopicForEvent(event)
	metrics := make([]*spbproto.Payload_Metric, 0, len(event.Metrics))
	now := uint64(event.Timestamp.UTC().UnixMilli())

	var seq uint64
	resetAfterBuild := false
	switch event.MsgType {
	case domain.MessageTypeNBIRTH:
		resetAfterBuild = true
	case domain.MessageTypeNDATA, domain.MessageTypeDDATA:
		seq = uint64(e.seq.NextSeq())
	case domain.MessageTypeNDEATH, domain.MessageTypeDDEATH:
		seq = 0
	}

	for _, metric := range event.Metrics {
		typ, err := e.resolver.InferDataType(metric.Value)
		if err != nil {
			return "", nil, err
		}
		name := metric.Name
		datatype := uint32(typ)
		wireMetric := &spbproto.Payload_Metric{Name: proto.String(name), Datatype: &datatype}
		switch v := metric.Value.(type) {
		case int32:
			wireMetric.Value = &spbproto.Payload_Metric_IntValue{IntValue: uint32(v)}
		case int64:
			wireMetric.Value = &spbproto.Payload_Metric_LongValue{LongValue: uint64(v)}
		case float32:
			wireMetric.Value = &spbproto.Payload_Metric_FloatValue{FloatValue: v}
		case float64:
			wireMetric.Value = &spbproto.Payload_Metric_DoubleValue{DoubleValue: v}
		case string:
			wireMetric.Value = &spbproto.Payload_Metric_StringValue{StringValue: v}
		case bool:
			wireMetric.Value = &spbproto.Payload_Metric_BooleanValue{BooleanValue: v}
		case []byte:
			wireMetric.Value = &spbproto.Payload_Metric_BytesValue{BytesValue: append([]byte(nil), v...)}
		default:
			return "", nil, fmt.Errorf("unsupported metric value %T", metric.Value)
		}
		metrics = append(metrics, wireMetric)
		_ = metric.Timestamp
		_ = metric.Alias
		_ = metric.Properties
	}

	payload := &spbproto.Payload{
		Timestamp: proto.Uint64(now),
		Metrics:   metrics,
		Seq:       proto.Uint64(seq),
	}
	if event.MsgType == domain.MessageTypeNBIRTH || event.MsgType == domain.MessageTypeNDEATH {
		e.mu.RLock()
		bd := interfaces.NormalizeBdSeq(e.bdSeq)
		e.mu.RUnlock()
		dt := uint32(spbproto.DataType_Int64)
		payload.Metrics = append(payload.Metrics, &spbproto.Payload_Metric{Name: proto.String("bdSeq"), Datatype: &dt, Value: &spbproto.Payload_Metric_LongValue{LongValue: bd}})
	}
	encoded, err := proto.Marshal(payload)
	if err != nil {
		return "", nil, err
	}
	if e.encrypt != nil {
		encoded, err = e.encrypt.Encrypt(encoded)
		if err != nil {
			return "", nil, err
		}
	}
	if resetAfterBuild {
		e.seq.ResetSeq()
		birth := event.Clone()
		e.mu.Lock()
		e.birth = &birth
		e.mu.Unlock()
	}
	e.state.Transition(event.MsgType)
	return topic, encoded, nil
}
