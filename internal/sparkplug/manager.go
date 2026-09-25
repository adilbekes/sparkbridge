package sparkplug

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	appcfg "sparkplugb-agent/internal/config"
	mqttclient "sparkplugb-agent/internal/mqtt"
	spb "sparkplugb-agent/pkg/pb"

	"github.com/google/uuid"
)

type Manager struct {
	log     *slog.Logger
	clients map[string]*clientRuntime
	subs    []chan *spb.CloudCommand
	mu      sync.RWMutex
}

type clientRuntime struct {
	cfg  appcfg.ClientConfig
	mqtt *mqttclient.Client
	seq  atomic.Uint32
}

func NewManager(ctx context.Context, cfg *appcfg.Config, log *slog.Logger) (*Manager, error) {
	m := &Manager{log: log, clients: make(map[string]*clientRuntime)}
	for _, clientCfg := range cfg.Clients {
		if !clientCfg.Enabled {
			continue
		}
		mqttc, err := mqttclient.New(mqttclient.Config{
			BrokerURL: clientCfg.MQTT.BrokerURL,
			ClientID:  clientCfg.MQTT.ClientID,
			Username:  clientCfg.MQTT.Username,
			Password:  clientCfg.MQTT.Password,
			KeepAlive: clientCfg.MQTT.KeepAliveSec,
			QoS:       clientCfg.MQTT.QoS,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("init client %s: %w", clientCfg.ID, err)
		}
		m.clients[clientCfg.ID] = &clientRuntime{cfg: clientCfg, mqtt: mqttc}
	}
	return m, nil
}

func (m *Manager) Close() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.clients {
		c.mqtt.Close()
	}
}

func (m *Manager) Subscribe() <-chan *spb.CloudCommand {
	ch := make(chan *spb.CloudCommand, 32)
	m.mu.Lock()
	m.subs = append(m.subs, ch)
	m.mu.Unlock()
	return ch
}

func (m *Manager) Publish(ctx context.Context, msg *spb.EdgeMessage) error {
	if msg == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	targets := m.targets(msg.ClientId)
	for _, rt := range targets {
		payload := []byte(fmt.Sprintf(`{"trace_id":"%s","client_id":"%s","device_id":"%s","metric_count":%d,"seq":%d}`,
			uuid.NewString(), rt.cfg.ID, msg.DeviceId, len(msg.Metrics), rt.nextSeq()))
		if err := rt.mqtt.Publish(ctx, rt.topic(), payload); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) targets(id string) []*clientRuntime {
	if id != "" {
		if rt, ok := m.clients[id]; ok {
			return []*clientRuntime{rt}
		}
		return nil
	}
	out := make([]*clientRuntime, 0, len(m.clients))
	for _, rt := range m.clients {
		out = append(out, rt)
	}
	return out
}

func (rt *clientRuntime) nextSeq() uint32 { return rt.seq.Add(1) & 0xff }

func (rt *clientRuntime) topic() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("spBv1.0/%s/NDATA/%s", rt.cfg.Sparkplug.GroupID, rt.cfg.Sparkplug.NodeID)
}

func init() {
	var x [8]byte
	_, _ = rand.Read(x[:])
	_ = binary.LittleEndian.Uint64(x[:])
}
