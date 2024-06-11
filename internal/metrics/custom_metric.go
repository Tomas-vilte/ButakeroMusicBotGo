package metrics

import "github.com/prometheus/client_golang/prometheus"

// CustomMetric define la interfaz que deben cumplir todas las métricas personalizadas.
type CustomMetric interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
	Inc(labels ...string)
}

type CacheMetrics interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
	IncHits()
	IncMisses()
	SetCacheSize(size float64)
	IncEvictions()
	IncRequests()
	IncSetOperations()
	IncGetOperations()
}
