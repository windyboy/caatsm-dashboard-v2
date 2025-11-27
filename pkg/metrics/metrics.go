package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// Registry exposes application specific Prometheus collectors.
type Registry struct {
	TelegramsIngested    *prometheus.CounterVec
	SearchLatency        prometheus.Histogram
	HTTPRequestTotal     *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	WebSocketConnections prometheus.Gauge
}

// NewRegistry registers base metrics with the provided Prometheus registry.
func NewRegistry(reg prometheus.Registerer) *Registry {
	r := &Registry{
		TelegramsIngested: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "telegrams_ingested_total",
			Help: "Total number of telegrams processed by the indexer.",
		}, []string{"status"}),
		SearchLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "search_latency_seconds",
			Help:    "Latency of search requests.",
			Buckets: prometheus.DefBuckets,
		}),
		HTTPRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "endpoint", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "endpoint"}),
		WebSocketConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "websocket_connections",
			Help: "Number of active WebSocket connections.",
		}),
	}

	reg.MustRegister(
		r.TelegramsIngested,
		r.SearchLatency,
		r.HTTPRequestTotal,
		r.HTTPRequestDuration,
		r.WebSocketConnections,
	)
	return r
}

// RecordSearchLatency records the latency of a search operation
func (r *Registry) RecordSearchLatency(duration float64) {
	r.SearchLatency.Observe(duration)
}

// RecordHTTPRequest records an HTTP request
func (r *Registry) RecordHTTPRequest(method, endpoint string, status int, duration float64) {
	r.HTTPRequestTotal.WithLabelValues(method, endpoint, strconv.Itoa(status)).Inc()
	r.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// UpdateWebSocketConnections updates the number of active WebSocket connections
func (r *Registry) UpdateWebSocketConnections(count float64) {
	r.WebSocketConnections.Set(count)
}
