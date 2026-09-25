//go:build input_mqtt || all_inputs

package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// Adapter translates raw MQTT messages into domain events.
type Adapter struct{}

// New creates a new raw MQTT adapter.
func New() *Adapter { return &Adapter{} }

// Decode converts a topic and JSON payload into a domain event.
func (a *Adapter) Decode(topic string, payload []byte) (domain.Event, error) {
	var raw struct {
		GroupID   string `json:"group_id"`
		NodeID    string `json:"node_id"`
		DeviceID  string `json:"device_id"`
		MsgType   string `json:"msg_type"`
		Timestamp string `json:"timestamp"`
		Metrics   []struct {
			Name  string `json:"name"`
			Value any    `json:"value"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return domain.Event{}, fmt.Errorf("topic %s: %w", topic, err)
	}
	ts := time.Now().UTC()
	if raw.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			return domain.Event{}, err
		}
		ts = parsed
	}
	event := domain.Event{GroupID: raw.GroupID, NodeID: raw.NodeID, DeviceID: raw.DeviceID, MsgType: domain.MessageType(raw.MsgType), Timestamp: ts}
	for _, m := range raw.Metrics {
		event.Metrics = append(event.Metrics, domain.Metric{Name: m.Name, Value: m.Value, Timestamp: ts})
	}
	return event, nil
}

// Start satisfies the input adapter interface.
func (a *Adapter) Start(ctx context.Context, outChan chan<- domain.Event) error {
	_ = ctx
	_ = outChan
	return nil
}

// Stop stops the adapter.
func (a *Adapter) Stop() error { return nil }

func init() { registry.RegisterInput("mqtt", func() (interfaces.InputAdapter, error) { return New(), nil }) }
