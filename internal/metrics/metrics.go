package metrics

import "github.com/prometheus/client_golang/prometheus"

// Registry exposes application specific Prometheus collectors.
type Registry struct {
	TelegramsIngested *prometheus.CounterVec
	SearchLatency     prometheus.Histogram
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
	}

	reg.MustRegister(r.TelegramsIngested, r.SearchLatency)
	return r
}
