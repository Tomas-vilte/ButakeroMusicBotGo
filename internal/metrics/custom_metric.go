package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// CustomMetric define la interfaz que deben cumplir todas las métricas personalizadas.
type CustomMetric interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
	Inc(labels ...string)
}

type CacheMetrics interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
	IncHits(cacheType string)
	IncMisses(cacheType string)
	SetCacheSize(size float64)
	IncEvictions(cacheType string)
	IncRequests(cacheType string)
	IncSetOperations(cacheType string)
	IncGetOperations(cacheType string)
	IncLatencyGet(cacheType string, duration time.Duration)
	IncLatencySet(cacheType string, duration time.Duration)
}
