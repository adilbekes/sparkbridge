package manager

import (
	"context"
	"errors"
	"sync"
	"time"

	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/config"
	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/engine"
	"sparkbridge/pkg/interfaces"
	"sparkbridge/pkg/pipeline"
	"sparkbridge/pkg/pipeline/middleware"
	"sparkbridge/pkg/pipeline/ratelimit"
)

// SparkBridge represents one named bridge instance.
type SparkBridge struct {
	Name   string
	Pool   *pipeline.WorkerPool
	Inputs []interfaces.InputAdapter
	Sinks  []interfaces.OutputSink
	Engine *engine.Engine
	Router *engine.CommandRouter
	Errors chan error
	ctx    context.Context
	cancel context.CancelFunc
}

// Manager keeps named bridge instances.
type Manager struct {
	mu        sync.RWMutex
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
	var inputs []interfaces.InputAdapter
	for _, inputCfg := range cfg.Inputs {
		factory, err := registry.Input(inputCfg.Kind)
		if err != nil {
			cancel()
			closeAdapters(inputs, nil)
			return nil, err
		}
		input, err := factory()
		if err != nil {
			cancel()
			closeAdapters(inputs, nil)
			return nil, err
		}
		inputs = append(inputs, input)
	}
	var sinks []interfaces.OutputSink
	if len(cfg.Sinks) > 0 {
		for _, sinkCfg := range cfg.Sinks {
			factory, err := registry.Sink(sinkCfg.Kind)
			if err != nil {
				cancel()
				closeAdapters(inputs, sinks)
				return nil, err
			}
			sink, err := factory()
			if err != nil {
				cancel()
				closeAdapters(inputs, sinks)
				return nil, err
			}
			sinks = append(sinks, sink)
		}
	}
	eng := engine.NewEngine(nil, nil, nil)
	pool := pipeline.NewWorkerPool(cfg.Workers, func(evt domain.Event) (pipeline.EncodedMessage, error) {
		topic, payload, err := eng.BuildPayload(bridgeCtx, evt)
		return pipeline.EncodedMessage{Topic: topic, Payload: payload, Timestamp: evt.Timestamp}, err
	}, ratelimit.New(100, 10*time.Millisecond), middleware.Validation(middleware.Enrichment(func(evt domain.Event) (domain.Event, error) { return evt, nil })))
	for _, sink := range sinks {
		pool.AddSink(sink)
	}
	router := engine.NewCommandRouter(eng, pool.In())
	for _, sink := range sinks {
		if commandSink, ok := sink.(interfaces.CommandRouterSink); ok {
			commandSink.SetCommandRouter(router.Handle)
		}
	}
	pool.Run(bridgeCtx)
	errorsCh := make(chan error, len(inputs))
	for _, input := range inputs {
		go func(input interfaces.InputAdapter) {
			if err := input.Start(bridgeCtx, pool.In()); err != nil && !errors.Is(err, context.Canceled) {
				select {
				case errorsCh <- err:
				case <-bridgeCtx.Done():
				}
			}
		}(input)
	}
	inst := &SparkBridge{Name: cfg.Name, Pool: pool, Inputs: inputs, Sinks: sinks, Engine: eng, Router: router, Errors: errorsCh, ctx: bridgeCtx, cancel: cancel}
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
	clear(m.instances)
}

func closeAdapters(inputs []interfaces.InputAdapter, sinks []interfaces.OutputSink) {
	for _, input := range inputs {
		_ = input.Stop()
	}
	for _, sink := range sinks {
		_ = sink.Close()
	}
}

// Close shuts down the instance.
func (b *SparkBridge) Close() {
	if b.cancel != nil {
		b.cancel()
	}
	for _, input := range b.Inputs {
		_ = input.Stop()
	}
	for _, sink := range b.Sinks {
		_ = sink.Close()
	}
	if b.Pool != nil {
		b.Pool.Close()
	}
}
