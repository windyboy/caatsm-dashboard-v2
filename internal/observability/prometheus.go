package observability

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsRegistry exposes application specific Prometheus collectors.
type MetricsRegistry struct {
	TelegramsIngested *prometheus.CounterVec
	SearchLatency     prometheus.Histogram
}

// NewMetricsRegistry registers base metrics with the provided Prometheus registry.
func NewMetricsRegistry(reg prometheus.Registerer) *MetricsRegistry {
	r := &MetricsRegistry{
		TelegramsIngested: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "telegrams_ingested_total",
			Help: "Total number of telegrams processed by the indexer.",
		}, []string{"status"}),
		SearchLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "search_latency_seconds",
			Help:    "Latency of search requests.",
			Buckets: prometheus.DefBuckets,
		}),
	}

	reg.MustRegister(r.TelegramsIngested, r.SearchLatency)
	return r
}

// MetricsExporter wires Prometheus registry and collectors for HTTP exposure.
type MetricsExporter struct {
	registry   *prometheus.Registry
	collectors *MetricsRegistry
}

// NewMetricsExporter constructs a MetricsExporter with the application's collectors.
// Constructs a MetricsExporter with a fresh Prometheus registry and metrics collectors.
func NewMetricsExporter() *MetricsExporter {
	reg := prometheus.NewRegistry()
	collectors := NewMetricsRegistry(reg)
	return &MetricsExporter{
		registry:   reg,
		collectors: collectors,
	}
}

// Handler returns an Echo handler exposing the metrics endpoint.
func (m *MetricsExporter) Handler() echo.HandlerFunc {
	handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
	return echo.WrapHandler(handler)
}

// Collectors returns references to the registered application metrics.
func (m *MetricsExporter) Collectors() *MetricsRegistry {
	return m.collectors
}

// Registry exposes the underlying Prometheus registry.
func (m *MetricsExporter) Registry() *prometheus.Registry {
	return m.registry
}
