package pipeline

import (
	"context"
	"sync"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
	"sparkbridge/pkg/pipeline/middleware"
	"sparkbridge/pkg/pipeline/ratelimit"
)

// WorkerPool processes events through middleware and emits encoded messages.
type WorkerPool struct {
	in       chan domain.Event
	out      chan EncodedMessage
	workers  int
	encoder  Encoder
	guard    *ratelimit.TokenBucket
	chain    middleware.Handler
	sinks    []interfaces.OutputSink
	wg       sync.WaitGroup
}

// NewWorkerPool creates a bounded pool.
func NewWorkerPool(workers int, encoder Encoder, guard *ratelimit.TokenBucket, chain middleware.Handler) *WorkerPool {
	if workers < 1 {
		workers = 1
	}
	return &WorkerPool{in: make(chan domain.Event, 64), out: make(chan EncodedMessage, 64), workers: workers, encoder: encoder, guard: guard, chain: chain}
}

// AddSink appends an output sink.
func (p *WorkerPool) AddSink(sink interfaces.OutputSink) { p.sinks = append(p.sinks, sink) }

// In returns the input channel.
func (p *WorkerPool) In() chan<- domain.Event { return p.in }

// Out returns the output channel.
func (p *WorkerPool) Out() <-chan EncodedMessage { return p.out }

// Run starts workers.
func (p *WorkerPool) Run(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case evt, ok := <-p.in:
					if !ok {
						return
					}
					if p.guard != nil && !p.guard.Allow() {
						continue
					}
					if p.chain != nil {
						var err error
						evt, err = p.chain(evt)
						if err != nil {
							continue
						}
					}
					if p.encoder == nil {
						continue
					}
					msg, err := p.encoder(evt)
					if err != nil {
						continue
					}
					select {
					case p.out <- msg:
						for _, sink := range p.sinks {
							if sink != nil {
								_ = sink.Publish(ctx, msg.Topic, msg.Payload)
							}
						}
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
}

// Close waits for all workers.
func (p *WorkerPool) Close() { p.wg.Wait() }
