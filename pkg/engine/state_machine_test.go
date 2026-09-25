package engine

import (
	"testing"

	"sparkbridge/pkg/domain"
)

func TestStateMachineTransitions(t *testing.T) {
	machine := NewStateMachine()
	if got := machine.Current(); got != NodeStateOffline {
		t.Fatalf("initial state = %s", got)
	}
	machine.Transition(domain.MessageTypeNBIRTH)
	if got := machine.Current(); got != NodeStateBirth {
		t.Fatalf("birth state = %s", got)
	}
	machine.Transition(domain.MessageTypeNDATA)
	if got := machine.Current(); got != NodeStateOnline {
		t.Fatalf("online state = %s", got)
	}
	machine.Transition(domain.MessageTypeNDEATH)
	if got := machine.Current(); got != NodeStateOffline {
		t.Fatalf("death state = %s", got)
	}
}
