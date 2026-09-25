//go:build input_http || all_inputs

package http

import (
	"context"
	"encoding/json"
	"io"
	"fmt"
	nethttp "net/http"
	"time"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// Adapter handles webhook requests.
type Adapter struct{}

// New creates a new HTTP adapter.
func New() *Adapter { return &Adapter{} }

// Handler returns an HTTP handler that emits domain events.
func (a *Adapter) Handler(outChan chan<- domain.Event) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodPost {
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			return
		}
		event, err := decodeEvent(data)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			return
		}
		event.Timestamp = time.Now().UTC()
		select {
		case outChan <- event:
			w.WriteHeader(nethttp.StatusAccepted)
		case <-r.Context().Done():
			w.WriteHeader(nethttp.StatusRequestTimeout)
		}
	})
}

// Start satisfies the input adapter interface.
func (a *Adapter) Start(ctx context.Context, outChan chan<- domain.Event) error {
	_ = ctx
	_ = outChan
	return nil
}

// Stop stops the adapter.
func (a *Adapter) Stop() error { return nil }

func decodeEvent(data []byte) (domain.Event, error) {
	var raw struct {
		GroupID   string `json:"group_id"`
		NodeID    string `json:"node_id"`
		DeviceID  string `json:"device_id"`
		MsgType   string `json:"msg_type"`
		Timestamp string `json:"timestamp"`
		Metrics   []struct {
			Name      string `json:"name"`
			Value     any    `json:"value"`
			Timestamp string `json:"timestamp"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return domain.Event{}, err
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
		mt := ts
		if m.Timestamp != "" {
			parsed, err := time.Parse(time.RFC3339Nano, m.Timestamp)
			if err != nil {
				return domain.Event{}, fmt.Errorf("metric %s: %w", m.Name, err)
			}
			mt = parsed
		}
		event.Metrics = append(event.Metrics, domain.Metric{Name: m.Name, Value: m.Value, Timestamp: mt})
	}
	return event, nil
}

func init() { registry.RegisterInput("http", func() (interfaces.InputAdapter, error) { return New(), nil }) }
