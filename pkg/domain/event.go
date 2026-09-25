package domain

import "time"

// MessageType identifies the Sparkplug message category.
type MessageType string

const (
	// MessageTypeNBIRTH marks a node birth payload.
	MessageTypeNBIRTH MessageType = "NBIRTH"
	// MessageTypeNDATA marks a node data payload.
	MessageTypeNDATA MessageType = "NDATA"
	// MessageTypeDBIRTH marks a device birth payload.
	MessageTypeDBIRTH MessageType = "DBIRTH"
	// MessageTypeDDATA marks a device data payload.
	MessageTypeDDATA MessageType = "DDATA"
	// MessageTypeDDEATH marks a device death payload.
	MessageTypeDDEATH MessageType = "DDEATH"
	// MessageTypeNCMD marks a node command payload.
	MessageTypeNCMD MessageType = "NCMD"
	// MessageTypeDCMD marks a device command payload.
	MessageTypeDCMD MessageType = "DCMD"
	// MessageTypeNDEATH marks a node death payload.
	MessageTypeNDEATH MessageType = "NDEATH"
)

// Metric is the canonical domain representation of a Sparkplug metric.
type Metric struct {
	Name       string
	Value      any
	DataType   string
	Timestamp  time.Time
	Alias      *uint64
	Properties map[string]any
}

// Event is the canonical domain event flowing through the pipeline.
type Event struct {
	GroupID    string
	NodeID     string
	DeviceID   string
	MsgType    MessageType
	Timestamp  time.Time
	Metrics    []Metric
	Properties map[string]any
}

// Clone returns an independent copy of the metric.
func (m Metric) Clone() Metric {
	clone := m
	if m.Alias != nil {
		alias := *m.Alias
		clone.Alias = &alias
	}
	clone.Properties = cloneMap(m.Properties)
	if value, ok := m.Value.([]byte); ok {
		clone.Value = append([]byte(nil), value...)
	}
	return clone
}

// Clone returns an independent copy of the event and its metrics.
func (e Event) Clone() Event {
	clone := e
	clone.Properties = cloneMap(e.Properties)
	clone.Metrics = make([]Metric, len(e.Metrics))
	for i := range e.Metrics {
		clone.Metrics[i] = e.Metrics[i].Clone()
	}
	return clone
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	clone := make(map[string]any, len(source))
	for key, value := range source {
		switch typed := value.(type) {
		case []byte:
			clone[key] = append([]byte(nil), typed...)
		case map[string]any:
			clone[key] = cloneMap(typed)
		default:
			clone[key] = value
		}
	}
	return clone
}
