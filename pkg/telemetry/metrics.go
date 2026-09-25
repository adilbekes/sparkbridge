package telemetry

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Metrics stores counters for the daemon.
type Metrics struct {
	mu sync.Mutex
	messagesReceived map[string]float64
	messagesPublished map[string]float64
	pipelineLatency []time.Duration
	activeWorkers int64
	circuitState map[string]string
}

// New creates metrics.
func New() *Metrics {
	return &Metrics{messagesReceived: map[string]float64{}, messagesPublished: map[string]float64{}, circuitState: map[string]string{}}
}

// Handler serves a Prometheus-like plaintext view.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		for k, v := range m.messagesReceived {
			_, _ = w.Write([]byte("sparkbridge_messages_received_total{" + `adapter="` + k + `"}` + " " + fmt.Sprintf("%.0f", v) + "\n"))
		}
		for k, v := range m.messagesPublished {
			_, _ = w.Write([]byte("sparkbridge_messages_published_total{" + `sink="` + k + `"}` + " " + fmt.Sprintf("%.0f", v) + "\n"))
		}
	})
}
