//go:build sink_mqtt || all_sinks

package mqtt

import (
	"context"
	"sync"

	mqttclient "github.com/eclipse/paho.mqtt.golang"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// Config represents the MQTT sink configuration.
type Config struct {
	LWTTopic   string
	LWTPayload []byte
	ClientID   string
	BrokerURL  string
}

// Sink is an MQTT client-backed sink.
type Sink struct {
	mu     sync.Mutex
	config Config
	client mqttclient.Client
}

// New creates a new MQTT sink.
func New(cfg Config) *Sink { return &Sink{config: cfg} }

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
	if s.client != nil && s.client.IsConnected() {
		if token := s.client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Close closes the sink.
func (s *Sink) Close() error { return nil }

func init() { registry.RegisterSink("mqtt", func() (interfaces.OutputSink, error) { return New(Config{LWTTopic: "spBv1.0/bridge/NDEATH/node", LWTPayload: []byte("NDEATH")}), nil }) }

var _ = domain.MessageTypeNDEATH
