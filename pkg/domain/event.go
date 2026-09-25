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
