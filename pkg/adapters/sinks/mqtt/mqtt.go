//go:build sink_mqtt || all_sinks

package mqtt

import (
	"context"
	"fmt"
	"strings"
	"sync"

	mqttclient "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
	"sparkbridge/pkg/pipeline"
	"sparkbridge/pkg/spbproto"
)

// Config represents the MQTT sink configuration.
type Config struct {
	LWTTopic   string
	LWTPayload []byte
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
	commandOutput chan<- domain.Event
	commandRouter func(context.Context, domain.Event) error
	subscriptions map[string]struct{}
	lastMessage   pipeline.EncodedMessage
}

// New creates a new MQTT sink.
func New(cfg Config) *Sink {
	return &Sink{config: cfg, subscriptions: map[string]struct{}{}}
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
	s.SetCommandRouter(router)
	return s
}

// SetCommandRouter wires command events into a rebirth-capable router.
func (s *Sink) SetCommandRouter(router func(context.Context, domain.Event) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandRouter = router
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
	opts.SetConnectionLostHandler(func(_ mqttclient.Client, _ error) {
		s.mu.Lock()
		clear(s.subscriptions)
		s.mu.Unlock()
	})
	s.client = mqttclient.NewClient(opts)
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Publish accepts payloads and stores them as the latest publish.
func (s *Sink) Publish(ctx context.Context, topic string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	s.lastMessage = pipeline.EncodedMessage{Topic: topic, Payload: payload}
	client := s.client
	s.mu.Unlock()
	if client != nil && client.IsConnected() {
		if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return s.handleLifecycle(ctx, topic)
}

// Subscribe records a topic subscription.
func (s *Sink) Subscribe(ctx context.Context, topic string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	s.subscriptions[topic] = struct{}{}
	client := s.client
	s.mu.Unlock()
	if client != nil && client.IsConnected() {
		if token := client.Subscribe(topic, 0, s.onMessage); token.Wait() && token.Error() != nil {
			s.mu.Lock()
			delete(s.subscriptions, topic)
			s.mu.Unlock()
			return token.Error()
		}
	}
	return nil
}

// Unsubscribe removes a topic subscription.
func (s *Sink) Unsubscribe(ctx context.Context, topic string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.subscriptions, topic)
	client := s.client
	s.mu.Unlock()
	if client != nil && client.IsConnected() {
		if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Close disconnects the MQTT client and clears command subscriptions.
func (s *Sink) Close() error {
	s.mu.Lock()
	client := s.client
	s.client = nil
	clear(s.subscriptions)
	s.mu.Unlock()
	if client != nil && client.IsConnected() {
		client.Disconnect(250)
	}
	return nil
}

func (s *Sink) onMessage(_ mqttclient.Client, msg mqttclient.Message) {
	if evt, err := decodeCommand(string(msg.Topic()), msg.Payload()); err == nil {
		s.mu.Lock()
		out := s.commandOutput
		router := s.commandRouter
		s.mu.Unlock()
		if router != nil {
			_ = router(context.Background(), evt)
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

func (s *Sink) handleLifecycle(ctx context.Context, topic string) error {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		return nil
	}
	if len(parts) >= 5 && parts[2] == "DBIRTH" {
		return s.Subscribe(ctx, fmt.Sprintf("%s/%s/DCMD/%s/%s", parts[0], parts[1], parts[3], parts[4]))
	}
	if len(parts) >= 4 && parts[2] == "NBIRTH" {
		if err := s.Subscribe(ctx, fmt.Sprintf("%s/%s/NCMD/%s", parts[0], parts[1], parts[3])); err != nil {
			return err
		}
		return s.Subscribe(ctx, fmt.Sprintf("%s/%s/DCMD/%s/+", parts[0], parts[1], parts[3]))
	}
	if len(parts) >= 4 && parts[2] == "NDEATH" {
		if err := s.Unsubscribe(ctx, fmt.Sprintf("%s/%s/NCMD/%s", parts[0], parts[1], parts[3])); err != nil {
			return err
		}
		return s.Unsubscribe(ctx, fmt.Sprintf("%s/%s/DCMD/%s/+", parts[0], parts[1], parts[3]))
	}
	if len(parts) >= 5 && parts[2] == "DDEATH" {
		return s.Unsubscribe(ctx, fmt.Sprintf("%s/%s/DCMD/%s/%s", parts[0], parts[1], parts[3], parts[4]))
	}
	return nil
}

func decodeCommand(topic string, payload []byte) (domain.Event, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 || (parts[2] != string(domain.MessageTypeNCMD) && parts[2] != string(domain.MessageTypeDCMD)) {
		return domain.Event{}, fmt.Errorf("invalid Sparkplug command topic %q", topic)
	}
	var wire spbproto.Payload
	if err := proto.Unmarshal(payload, &wire); err != nil {
		return domain.Event{}, err
	}
	evt := domain.Event{GroupID: parts[1], NodeID: parts[3], MsgType: domain.MessageType(parts[2])}
	if len(parts) >= 5 {
		evt.DeviceID = parts[4]
	}
	for _, metric := range wire.Metrics {
		var value any
		switch spbproto.DataType(metric.GetDatatype()) {
		case spbproto.DataType_Boolean:
			value = metric.GetBooleanValue()
		case spbproto.DataType_Int8, spbproto.DataType_Int16, spbproto.DataType_Int32:
			value = int64(int32(metric.GetIntValue()))
		case spbproto.DataType_Int64:
			value = int64(metric.GetLongValue())
		case spbproto.DataType_UInt8, spbproto.DataType_UInt16, spbproto.DataType_UInt32:
			value = uint64(metric.GetIntValue())
		case spbproto.DataType_UInt64:
			value = metric.GetLongValue()
		case spbproto.DataType_Float:
			value = metric.GetFloatValue()
		case spbproto.DataType_Double:
			value = metric.GetDoubleValue()
		case spbproto.DataType_String, spbproto.DataType_Text, spbproto.DataType_UUID:
			value = metric.GetStringValue()
		case spbproto.DataType_Bytes, spbproto.DataType_File:
			value = append([]byte(nil), metric.GetBytesValue()...)
		default:
			return domain.Event{}, fmt.Errorf("unsupported command metric datatype %d", metric.GetDatatype())
		}
		evt.Metrics = append(evt.Metrics, domain.Metric{Name: metric.GetName(), Value: value})
	}
	return evt, nil
}

func init() {
	registry.RegisterSink("mqtt", func() (interfaces.OutputSink, error) {
		return New(Config{LWTTopic: "spBv1.0/bridge/NDEATH/node", LWTPayload: []byte("NDEATH")}), nil
	})
}
