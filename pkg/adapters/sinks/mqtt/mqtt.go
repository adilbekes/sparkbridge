//go:build sink_mqtt || all_sinks

package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"strings"
	"sync"

	mqttclient "github.com/eclipse/paho.mqtt.golang"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
	"sparkbridge/pkg/pipeline"
)

// Config represents the MQTT sink configuration.
type Config struct {
	LWTTopic   string
	LWTPayload  []byte
	ClientID   string
	BrokerURL  string
	GroupID    string
	NodeID     string
	DeviceIDs  []string
}

// Sink is an MQTT client-backed sink.
type Sink struct {
	mu            sync.Mutex
	config        Config
	client        mqttclient.Client
	commandOutput  chan<- domain.Event
	commandRouter  func(context.Context, domain.Event) error
	commandSink   interfaces.OutputSink
	subscriptions map[string]struct{}
	lastMessage   pipeline.EncodedMessage
}

// New creates a new MQTT sink.
func New(cfg Config) *Sink {
	return &Sink{config: cfg, subscriptions: map[string]struct{}{}, commandOutput: make(chan domain.Event, 32)}
}

// WithCommandOutput wires inbound command events into the pipeline.
func (s *Sink) WithCommandOutput(out chan<- domain.Event) *Sink {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandOutput = out
	return s
}

// WithCommandRouter wires command events directly into a rebirth-capable router.
func (s *Sink) WithCommandRouter(router func(context.Context, domain.Event) error) *Sink {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandRouter = router
	return s
}

// WithCommandSink wires the rebirth path back to an output sink.
func (s *Sink) WithCommandSink(sink interfaces.OutputSink) *Sink {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandSink = sink
	return s
}

// Connect initializes the MQTT client and LWT configuration.
func (s *Sink) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil && s.client.IsConnected() {
		return nil
	}
	if s.config.BrokerURL == "" {
		return nil
	}
	opts := mqttclient.NewClientOptions().AddBroker(s.config.BrokerURL)
	if s.config.ClientID != "" {
		opts.SetClientID(s.config.ClientID)
	}
	if s.config.LWTTopic != "" {
		opts.SetWill(s.config.LWTTopic, string(s.config.LWTPayload), 0, false)
	}
	opts.SetDefaultPublishHandler(s.onMessage)
	s.client = mqttclient.NewClient(opts)
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Publish accepts payloads and stores them as the latest publish.
func (s *Sink) Publish(ctx context.Context, topic string, payload []byte) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastMessage = pipeline.EncodedMessage{Topic: topic, Payload: payload}
	if s.client != nil && s.client.IsConnected() {
		if token := s.client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return s.handleLifecycle(topic)
}

// Subscribe records a topic subscription.
func (s *Sink) Subscribe(ctx context.Context, topic string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[topic] = struct{}{}
	if s.client != nil && s.client.IsConnected() {
		if token := s.client.Subscribe(topic, 0, s.onMessage); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Unsubscribe removes a topic subscription.
func (s *Sink) Unsubscribe(ctx context.Context, topic string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscriptions, topic)
	if s.client != nil && s.client.IsConnected() {
		if token := s.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Close closes the sink.
func (s *Sink) Close() error { return nil }

func (s *Sink) onMessage(_ mqttclient.Client, msg mqttclient.Message) {
	if evt, err := decodeCommand(string(msg.Topic()), msg.Payload()); err == nil {
		s.mu.Lock()
		out := s.commandOutput
		router := s.commandRouter
		sink := s.commandSink
		s.mu.Unlock()
		if router != nil {
			_ = router(context.Background(), evt)
			return
		}
		if sink != nil && evt.MsgType == domain.MessageTypeNCMD && isRebirthMetric(evt) {
			trigger := domain.Event{GroupID: evt.GroupID, NodeID: evt.NodeID, MsgType: domain.MessageTypeNBIRTH, Timestamp: time.Now().UTC(), Metrics: evt.Metrics}
			if payload, err := encodeRebirth(trigger); err == nil {
				_ = sink.Publish(context.Background(), triggerTopic(trigger), payload)
			}
			return
		}
		if out != nil {
			select {
			case out <- evt:
			default:
			}
		}
	}
}

func isRebirthMetric(evt domain.Event) bool {
	for _, metric := range evt.Metrics {
		if metric.Name != "Node Control/Rebirth" {
			continue
		}
		if v, ok := metric.Value.(bool); ok && v {
			return true
		}
		if v, ok := metric.Value.(string); ok && v == "true" {
			return true
		}
	}
	return false
}

func triggerTopic(evt domain.Event) string {
	return fmt.Sprintf("spBv1.0/%s/NBIRTH/%s", evt.GroupID, evt.NodeID)
}

func encodeRebirth(evt domain.Event) ([]byte, error) {
	return json.Marshal(map[string]any{
		"group_id": evt.GroupID,
		"node_id":  evt.NodeID,
		"msg_type":  string(evt.MsgType),
		"timestamp": evt.Timestamp.UTC().Format(time.RFC3339Nano),
		"metrics":   evt.Metrics,
	})
}

func (s *Sink) handleLifecycle(topic string) error {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		return nil
	}
	if len(parts) >= 5 && parts[2] == "DBIRTH" {
		return s.Subscribe(context.Background(), strings.Join(parts[:5], "/"))
	}
	if len(parts) >= 4 && parts[2] == "NBIRTH" {
		groupID := parts[1]
		nodeID := parts[3]
		_ = s.Subscribe(context.Background(), fmt.Sprintf("spBv1.0/%s/NCMD/%s", groupID, nodeID))
		_ = s.Subscribe(context.Background(), fmt.Sprintf("spBv1.0/%s/DCMD/%s/+", groupID, nodeID))
	}
	if len(parts) >= 4 && parts[2] == "NDEATH" {
		groupID := parts[1]
		nodeID := parts[3]
		_ = s.Unsubscribe(context.Background(), fmt.Sprintf("spBv1.0/%s/NCMD/%s", groupID, nodeID))
		_ = s.Unsubscribe(context.Background(), fmt.Sprintf("spBv1.0/%s/DCMD/%s/+", groupID, nodeID))
	}
	if len(parts) >= 5 && parts[2] == "DDEATH" {
		groupID := parts[1]
		nodeID := parts[3]
		deviceID := parts[4]
		_ = s.Unsubscribe(context.Background(), fmt.Sprintf("spBv1.0/%s/DCMD/%s/%s", groupID, nodeID, deviceID))
	}
	return nil
}

func decodeCommand(topic string, payload []byte) (domain.Event, error) {
	var raw struct {
		GroupID  string `json:"group_id"`
		NodeID   string `json:"node_id"`
		DeviceID string `json:"device_id"`
		MsgType  string `json:"msg_type"`
		Metrics  []struct {
			Name  string `json:"name"`
			Value any    `json:"value"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return domain.Event{}, err
	}
	evt := domain.Event{GroupID: raw.GroupID, NodeID: raw.NodeID, DeviceID: raw.DeviceID, MsgType: domain.MessageType(raw.MsgType)}
	if evt.GroupID == "" || evt.NodeID == "" {
		parts := strings.Split(topic, "/")
		if len(parts) >= 4 {
			evt.GroupID = parts[1]
			evt.NodeID = parts[3]
		}
		if len(parts) >= 5 {
			evt.DeviceID = parts[4]
		}
	}
	for _, metric := range raw.Metrics {
		evt.Metrics = append(evt.Metrics, domain.Metric{Name: metric.Name, Value: metric.Value})
	}
	return evt, nil
}

func init() { registry.RegisterSink("mqtt", func() (interfaces.OutputSink, error) { return New(Config{LWTTopic: "spBv1.0/bridge/NDEATH/node", LWTPayload: []byte("NDEATH")}), nil }) }
