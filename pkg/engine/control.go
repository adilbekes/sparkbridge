package engine

import (
	"context"
	"time"

	"sparkbridge/pkg/domain"
	"sparkbridge/pkg/interfaces"
)

// CommandRouter handles NCMD/DCMD events and can trigger rebirth republishing.
type CommandRouter struct {
	engine *Engine
	out    chan<- domain.Event
	sink   interfaces.OutputSink
}

// NewCommandRouter creates a command router.
func NewCommandRouter(engine *Engine, out chan<- domain.Event, sink interfaces.OutputSink) *CommandRouter {
	return &CommandRouter{engine: engine, out: out, sink: sink}
}

// Handle routes a command event.
func (r *CommandRouter) Handle(ctx context.Context, evt domain.Event) error {
	if IsRebirthCommand(evt) {
		return r.rebirth(ctx, evt)
	}
	select {
	case r.out <- evt:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *CommandRouter) rebirth(ctx context.Context, evt domain.Event) error {
	if r.engine == nil || r.out == nil {
		return nil
	}
	if err := r.engine.InitBDSeq(ctx); err != nil {
		return err
	}
	trigger := domain.Event{GroupID: evt.GroupID, NodeID: evt.NodeID, MsgType: domain.MessageTypeNBIRTH, Timestamp: time.Now().UTC(), Metrics: append([]domain.Metric(nil), evt.Metrics...)}
	topic, payload, err := r.engine.BuildPayload(ctx, trigger)
	if err != nil {
		return err
	}
	if r.sink != nil {
		if err := r.sink.Publish(ctx, topic, payload); err != nil {
			return err
		}
	}
	select {
	case r.out <- trigger:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
