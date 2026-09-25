package engine

import (
	"sync"

	"sparkbridge/pkg/domain"
)

// NodeState represents the Sparkplug node lifecycle state.
type NodeState string

const (
	// NodeStateOffline indicates that the node is not publishing a live session.
	NodeStateOffline NodeState = "offline"
	// NodeStateBirth indicates that an NBIRTH publication is pending or in progress.
	NodeStateBirth NodeState = "birth"
	// NodeStateOnline indicates that the node has entered its live data lifecycle.
	NodeStateOnline NodeState = "online"
)

// StateMachine tracks node lifecycle transitions.
type StateMachine struct {
	mu    sync.RWMutex
	state NodeState
}

// NewStateMachine creates an offline node state machine.
func NewStateMachine() *StateMachine { return &StateMachine{state: NodeStateOffline} }

// Transition applies a Sparkplug message lifecycle transition.
func (s *StateMachine) Transition(messageType domain.MessageType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch messageType {
	case domain.MessageTypeNBIRTH:
		s.state = NodeStateBirth
	case domain.MessageTypeNDATA, domain.MessageTypeDBIRTH, domain.MessageTypeDDATA:
		s.state = NodeStateOnline
	case domain.MessageTypeNDEATH:
		s.state = NodeStateOffline
	}
}

// Current returns the current node lifecycle state.
func (s *StateMachine) Current() NodeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
