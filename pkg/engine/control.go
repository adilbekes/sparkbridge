package engine

import (
	"context"
	"time"

	"sparkbridge/pkg/domain"
)

// CommandRouter handles NCMD/DCMD events and can trigger rebirth republishing.
type CommandRouter struct {
	engine *Engine
	out    chan<- domain.Event
}

// NewCommandRouter creates a command router.
func NewCommandRouter(engine *Engine, out chan<- domain.Event) *CommandRouter {
	return &CommandRouter{engine: engine, out: out}
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
	trigger, ok := r.engine.BirthEvent()
	if !ok {
		trigger = domain.Event{GroupID: evt.GroupID, NodeID: evt.NodeID, MsgType: domain.MessageTypeNBIRTH}
	}
	trigger.MsgType = domain.MessageTypeNBIRTH
	trigger.Timestamp = time.Now().UTC()
	r.engine.Transition(domain.MessageTypeNBIRTH)
	select {
	case r.out <- trigger:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
