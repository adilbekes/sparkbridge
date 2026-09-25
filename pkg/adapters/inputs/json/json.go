//go:build input_json || all_inputs

package json

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// Adapter converts JSON streams into domain events.
type Adapter struct {
	file  string
	stdin bool
}

// Config describes the source to read from.
type Config struct {
	File string
	Stdin bool
}

// New creates a new JSON file adapter.
func New(cfg Config) *Adapter {
	return &Adapter{file: cfg.File, stdin: cfg.Stdin}
}

// Start reads JSON data and forwards it to the pipeline.
func (a *Adapter) Start(ctx context.Context, outChan chan<- domain.Event) error {
	var r io.Reader
	switch {
	case a.file == "" || a.file == "stdin" || a.stdin:
		r = os.Stdin
	default:
		f, err := os.Open(a.file)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	event, err := DecodeEvent(data)
	if err != nil {
		return err
	}
	select {
	case outChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop stops the adapter.
func (a *Adapter) Stop() error { return nil }

// DecodeEvent parses a JSON payload into a domain event.
func DecodeEvent(data []byte) (domain.Event, error) {
	var raw struct {
		GroupID    string `json:"group_id"`
		NodeID     string `json:"node_id"`
		DeviceID   string `json:"device_id"`
		MsgType    string `json:"msg_type"`
		Timestamp  string `json:"timestamp"`
		Metrics    []struct {
			Name      string      `json:"name"`
			Value     any         `json:"value"`
			Timestamp string      `json:"timestamp"`
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

func init() { registry.RegisterInput("json", func() (interfaces.InputAdapter, error) { return New(Config{}), nil }) }
