package middleware

import (
	"testing"
	"time"

	"sparkbridge/pkg/domain"
)

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		event   domain.Event
		wantErr bool
	}{
		{name: "valid", event: domain.Event{GroupID: "group", NodeID: "node"}},
		{name: "missing group", event: domain.Event{NodeID: "node"}, wantErr: true},
		{name: "missing node", event: domain.Event{GroupID: "group"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Validation(nil)(tt.event)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnrichmentAddsTimestamp(t *testing.T) {
	event, err := Enrichment(nil)(domain.Event{})
	if err != nil {
		t.Fatalf("enrich: %v", err)
	}
	if event.Timestamp.IsZero() || time.Since(event.Timestamp) > time.Second {
		t.Fatalf("unexpected timestamp %v", event.Timestamp)
	}
}
