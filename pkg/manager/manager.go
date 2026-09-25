package manager

import (
	"context"
	"sync"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/config"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
	"sparkbridge/pkg/engine"
	"sparkbridge/pkg/pipeline"
	"sparkbridge/pkg/pipeline/middleware"
	"sparkbridge/pkg/pipeline/ratelimit"
)

// SparkBridge represents one named bridge instance.
type SparkBridge struct {
	Name    string
	Pool    *pipeline.WorkerPool
	Sinks   []interfaces.OutputSink
	Engine  *engine.Engine
	Router  *engine.CommandRouter
	ctx     context.Context
	cancel  context.CancelFunc
}

// Manager keeps named bridge instances.
type Manager struct {
	mu       sync.RWMutex
	instances map[string]*SparkBridge
}

// New creates a new bridge manager.
func New() *Manager { return &Manager{instances: map[string]*SparkBridge{}} }

// Get returns or creates a bridge instance.
func (m *Manager) Get(ctx context.Context, cfg config.Bridge) (*SparkBridge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[cfg.Name]; ok {
		return inst, nil
	}
	bridgeCtx, cancel := context.WithCancel(ctx)
	var sinks []interfaces.OutputSink
	if len(cfg.Sinks) > 0 {
		for _, sinkCfg := range cfg.Sinks {
			factory, err := registry.Sink(sinkCfg.Kind)
			if err != nil {
				cancel()
				return nil, err
			}
			sink, err := factory()
			if err != nil {
				cancel()
				return nil, err
			}
			sinks = append(sinks, sink)
		}
	}
	eng := engine.NewEngine(nil, nil, nil)
	pool := pipeline.NewWorkerPool(cfg.Workers, func(evt domain.Event) (pipeline.EncodedMessage, error) {
		if evt.MsgType == domain.MessageTypeNCMD || evt.MsgType == domain.MessageTypeDCMD {
			if engine.IsRebirthCommand(evt) {
				rebirthEvt := domain.Event{GroupID: evt.GroupID, NodeID: evt.NodeID, MsgType: domain.MessageTypeNBIRTH, Timestamp: evt.Timestamp, Metrics: evt.Metrics}
				topic, payload, err := eng.BuildPayload(ctx, rebirthEvt)
				if err != nil {
					return pipeline.EncodedMessage{}, err
				}
				return pipeline.EncodedMessage{Topic: topic, Payload: payload, Timestamp: evt.Timestamp}, nil
			}
		}
		return pipeline.EncodedMessage{Topic: cfg.Name, Payload: []byte(evt.NodeID)}, nil
	}, ratelimit.New(100, 10), middleware.Validation(middleware.Enrichment(func(evt domain.Event) (domain.Event, error) { return evt, nil })))
	router := engine.NewCommandRouter(eng, pool.In(), nil)
	inst := &SparkBridge{Name: cfg.Name, Pool: pool, Sinks: sinks, Engine: eng, Router: router, ctx: bridgeCtx, cancel: cancel}
	m.instances[cfg.Name] = inst
	return inst, nil
}

// Close shuts down all managed bridge instances.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, inst := range m.instances {
		if inst != nil {
			inst.Close()
		}
	}
}

// Close shuts down the instance.
func (b *SparkBridge) Close() {
	if b.cancel != nil {
		b.cancel()
	}
	for _, sink := range b.Sinks {
		_ = sink.Close()
	}
	if b.Pool != nil {
		b.Pool.Close()
	}
}
