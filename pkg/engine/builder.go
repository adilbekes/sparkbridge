package engine

import (
	"context"
	"fmt"

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
		return spbproto.DataType_DATA_TYPE_INT32, nil
	case int64:
		return spbproto.DataType_DATA_TYPE_INT64, nil
	case float32:
		return spbproto.DataType_DATA_TYPE_FLOAT, nil
	case float64:
		return spbproto.DataType_DATA_TYPE_DOUBLE, nil
	case string:
		return spbproto.DataType_DATA_TYPE_STRING, nil
	case bool:
		return spbproto.DataType_DATA_TYPE_BOOLEAN, nil
	case []byte:
		return spbproto.DataType_DATA_TYPE_BYTES, nil
	default:
		return spbproto.DataType_DATA_TYPE_UNKNOWN, fmt.Errorf("unsupported payload value %T", val)
	}
}

// Engine ties together sequence, state, encryption, and type resolution.
type Engine struct {
	seq      *SequenceManager
	store    interfaces.StateStore
	encrypt  interfaces.Encryptor
	resolver TypeResolver
	bdSeq    uint64
}

// NewEngine creates a new engine.
func NewEngine(store interfaces.StateStore, encrypt interfaces.Encryptor, resolver TypeResolver) *Engine {
	if resolver == nil {
		resolver = DefaultTypeResolver{}
	}
	return &Engine{seq: NewSequenceManager(), store: store, encrypt: encrypt, resolver: resolver}
}

// InitBDSeq loads and persists bdSeq for the node.
func (e *Engine) InitBDSeq(ctx context.Context) error {
	bdSeq, err := InitializeBDSeq(ctx, e.store)
	if err != nil {
		return err
	}
	e.bdSeq = bdSeq
	return nil
}

// BuildPayload builds topic and payload bytes for the supplied event.
func (e *Engine) BuildPayload(ctx context.Context, event domain.Event) (string, []byte, error) {
	topic := TopicForEvent(event)
	_ = ctx
	metrics := make([]*spbproto.Metric, 0, len(event.Metrics))
	now := uint64(event.Timestamp.UTC().UnixMilli())

	var seq uint64
	switch event.MsgType {
	case domain.MessageTypeNBIRTH:
		e.seq.ResetSeq()
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
		datatype := typ
		switch v := metric.Value.(type) {
		case int32:
			iv := int64(v)
			metrics = append(metrics, &spbproto.Metric{Name: proto.String(name), Datatype: &datatype, IntValue: &iv})
		case int64:
			iv := v
			metrics = append(metrics, &spbproto.Metric{Name: proto.String(name), Datatype: &datatype, IntValue: &iv})
		case float32:
			fv := float64(v)
			metrics = append(metrics, &spbproto.Metric{Name: proto.String(name), Datatype: &datatype, DoubleValue: &fv})
		case float64:
			fv := v
			metrics = append(metrics, &spbproto.Metric{Name: proto.String(name), Datatype: &datatype, DoubleValue: &fv})
		case string:
			sv := v
			metrics = append(metrics, &spbproto.Metric{Name: proto.String(name), Datatype: &datatype, StringValue: &sv})
		case bool:
			bv := v
			metrics = append(metrics, &spbproto.Metric{Name: proto.String(name), Datatype: &datatype, BoolValue: &bv})
		default:
			return "", nil, fmt.Errorf("unsupported metric value %T", metric.Value)
		}
		_ = metric.Timestamp
		_ = metric.Alias
		_ = metric.Properties
	}

	payload := &spbproto.Payload{
		Timestamp: proto.Uint64(now),
		Metrics:   metrics,
		Seq:       proto.Uint64(seq),
		Topic:     proto.String(topic),
		MsgType:   proto.String(string(event.MsgType)),
	}
	if event.MsgType == domain.MessageTypeNBIRTH || event.MsgType == domain.MessageTypeNDEATH {
		bd := interfaces.NormalizeBdSeq(e.bdSeq)
		dt := spbproto.DataType_DATA_TYPE_INT64
		payload.Metrics = append(payload.Metrics, &spbproto.Metric{Name: proto.String("bdSeq"), Datatype: &dt, IntValue: ptrInt64(int64(bd))})
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
	return topic, encoded, nil
}

func ptrInt64(v int64) *int64 { return &v }
