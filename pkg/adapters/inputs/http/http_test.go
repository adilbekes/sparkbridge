//go:build input_http || all_inputs

package http

import (
	"sparkbridge/pkg/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerRejectsNonPOST(t *testing.T) {
	handler := New().Handler(make(chan domain.Event, 1))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
    t.Fatalf("expected 405, got %d", rec.Code)
}
}

func TestHandlerAcceptsValidJSON(t *testing.T) {
	out := make(chan domain.Event, 1)
	handler := New().Handler(out)
	body := strings.NewReader(`{"group_id":"g1","node_id":"n1","msg_type":"NDATA","metrics":[{"name":"temp","value":42}]}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}
