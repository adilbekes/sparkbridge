package middleware

import (
	"fmt"
	"strings"
	"time"

	"sparkbridge/pkg/domain"
)

// Handler processes an event and optionally passes it onward.
type Handler func(domain.Event) (domain.Event, error)

// Validation validates core event fields.
func Validation(next Handler) Handler {
	return func(evt domain.Event) (domain.Event, error) {
		if strings.TrimSpace(evt.GroupID) == "" || strings.TrimSpace(evt.NodeID) == "" {
			return domain.Event{}, fmt.Errorf("invalid event identity")
		}
		return next(evt)
	}
}

// Enrichment ensures timestamps are present.
func Enrichment(next Handler) Handler {
	return func(evt domain.Event) (domain.Event, error) {
		if evt.Timestamp.IsZero() {
			evt.Timestamp = time.Now().UTC()
		}
		return next(evt)
	}
}
