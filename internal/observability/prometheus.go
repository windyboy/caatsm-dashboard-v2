package observability

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/windy/caatsm-dashboard/internal/metrics"
)

// MetricsExporter wires Prometheus registry and collectors for HTTP exposure.
type MetricsExporter struct {
	registry   *prometheus.Registry
	collectors *metrics.Registry
}

// NewMetricsExporter constructs a MetricsExporter with the application's collectors.
// Constructs a MetricsExporter with a fresh Prometheus registry and metrics collectors.
func NewMetricsExporter() *MetricsExporter {
	reg := prometheus.NewRegistry()
	collectors := metrics.NewRegistry(reg)
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
func (m *MetricsExporter) Collectors() *metrics.Registry {
	return m.collectors
}

// Registry exposes the underlying Prometheus registry.
func (m *MetricsExporter) Registry() *prometheus.Registry {
	return m.registry
}
