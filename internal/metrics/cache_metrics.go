package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type (
	// CachePrometheusMetrics es una métrica para contabilizar el número de búsquedas en caché exitosas y fallidas.
	CachePrometheusMetrics struct {
		cacheHits      prometheus.Counter
		cacheMisses    prometheus.Counter
		cacheSize      prometheus.Gauge
		cacheEvictions prometheus.Counter
		cacheRequests  prometheus.Counter
		cacheSetOps    prometheus.Counter
		cacheGetOps    prometheus.Counter
		latencyGet     prometheus.Histogram
		latencySet     prometheus.Histogram
	}
)

// NewCacheMetrics crea una nueva instancia de CacheMetrics.
func NewCacheMetrics() CacheMetrics {
	return &CachePrometheusMetrics{
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Número total de aciertos en caché",
		}),
		cacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Número total de fallos en caché",
		}),
		cacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cache_size",
			Help: "Tamaño actual del caché",
		}),
		cacheEvictions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_evictions_total",
			Help: "Número total de eliminaciones de caché debido a la expiración",
		}),
		cacheRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_requests_total",
			Help: "Número total de solicitudes de caché",
		}),
		cacheSetOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_set_operations_total",
			Help: "Número total de operaciones de establecimiento en caché",
		}),
		cacheGetOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_get_operations_total",
			Help: "Número total de operaciones de obtención en caché",
		}),
		latencyGet: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "custom_cache_get_latency_seconds",
			Help:    "Latencia de las operaciones de obtención en caché",
			Buckets: prometheus.DefBuckets,
		}),
		latencySet: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "custom_cache_set_latency_seconds",
			Help:    "Latencia de las operaciones de establecimiento en caché",
			Buckets: prometheus.DefBuckets,
		}),
	}
}

// Describe implementa el método Describe de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) Describe(ch chan<- *prometheus.Desc) {
	c.cacheHits.Describe(ch)
	c.cacheMisses.Describe(ch)
	c.cacheSize.Describe(ch)
	c.cacheEvictions.Describe(ch)
	c.cacheRequests.Describe(ch)
	c.cacheSetOps.Describe(ch)
	c.cacheGetOps.Describe(ch)
	c.latencyGet.Describe(ch)
	c.latencySet.Describe(ch)
}

// Collect implementa el método Collect de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) Collect(ch chan<- prometheus.Metric) {
	c.cacheHits.Collect(ch)
	c.cacheMisses.Collect(ch)
	c.cacheSize.Collect(ch)
	c.cacheEvictions.Collect(ch)
	c.cacheRequests.Collect(ch)
	c.cacheSetOps.Collect(ch)
	c.cacheGetOps.Collect(ch)
	c.latencyGet.Collect(ch)
	c.latencySet.Collect(ch)
}

// IncHits implementa el método IncHits de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) IncHits() {
	c.cacheHits.Inc()
}

// IncMisses implementa el método IncMisses de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) IncMisses() {
	c.cacheMisses.Inc()
}

// SetCacheSize implementa el método SetCacheSize de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) SetCacheSize(size float64) {
	c.cacheSize.Set(size)
}

// IncEvictions implementa el método IncEvictions de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) IncEvictions() {
	c.cacheEvictions.Inc()
}

// IncRequests implementa el método IncRequests de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) IncRequests() {
	c.cacheRequests.Inc()
}

// IncSetOperations implementa el método IncSetOperations de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) IncSetOperations() {
	c.cacheSetOps.Inc()
}

// IncGetOperations implementa el método IncGetOperations de la interfaz CacheMetrics.
func (c *CachePrometheusMetrics) IncGetOperations() {
	c.cacheGetOps.Inc()
}

func (c *CachePrometheusMetrics) IncLatencyGet(duration time.Duration) {
	c.latencyGet.Observe(duration.Seconds())
}

func (c *CachePrometheusMetrics) IncLatencySet(duration time.Duration) {
	c.latencySet.Observe(duration.Seconds())
}
