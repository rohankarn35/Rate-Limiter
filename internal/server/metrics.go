package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics exposes Prometheus counters for limiter behavior.
type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
}

// NewMetrics registers metrics with a fresh registry.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "rate_limiter",
		Name:      "requests_total",
		Help:      "Total requests processed by the rate limiter",
	}, []string{"policy", "result"})
	reg.MustRegister(requests)

	return &Metrics{
		registry: reg,
		requests: requests,
	}
}

// Observe records the outcome.
func (m *Metrics) Observe(policy string, allowed bool) {
	if m == nil {
		return
	}
	status := "limited"
	if allowed {
		status = "allowed"
	}
	if policy == "" {
		policy = "unmatched"
	}
	m.requests.WithLabelValues(policy, status).Inc()
}

// Handler returns an HTTP handler serving the registry.
func (m *Metrics) Handler() http.Handler {
	if m == nil {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
