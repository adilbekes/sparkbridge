package pipeline

import (
	"context"
	"sync"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// Encoder converts domain events into outbound messages.
type Encoder func(domain.Event) (EncodedMessage, error)

// Pipeline connects inputs to outputs through a worker pool.
type Pipeline struct {
	eventsChan   chan domain.Event
	messagesChan chan EncodedMessage
	workers      int
	encoder      Encoder
	sink         interfaces.OutputSink
	wg           sync.WaitGroup
	once         sync.Once
}

// New creates a pipeline instance.
func New(workers int, encoder Encoder, sink interfaces.OutputSink) *Pipeline {
	if workers < 1 {
		workers = 1
	}
	return &Pipeline{
		eventsChan:   make(chan domain.Event, 32),
		messagesChan: make(chan EncodedMessage, 32),
		workers:      workers,
		encoder:      encoder,
		sink:         sink,
	}
}

// Events returns the inbound events channel.
func (p *Pipeline) Events() chan<- domain.Event { return p.eventsChan }

// Messages returns the outbound messages channel.
func (p *Pipeline) Messages() <-chan EncodedMessage { return p.messagesChan }

// Run starts the worker pool and relay loop.
func (p *Pipeline) Run(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case evt, ok := <-p.eventsChan:
					if !ok {
						return
					}
					if p.encoder == nil {
						continue
					}
					msg, err := p.encoder(evt)
					if err != nil {
						continue
					}
					select {
					case p.messagesChan <- msg:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-p.messagesChan:
				if !ok {
					return
				}
				if p.sink != nil {
					_ = p.sink.Publish(ctx, msg.Topic, msg.Payload)
				}
			}
		}
	}()
}

// Close closes the pipeline channels and waits for shutdown.
func (p *Pipeline) Close() {
	p.once.Do(func() {
		close(p.eventsChan)
	})
	p.wg.Wait()
}
