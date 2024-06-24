package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type (
	// CachePrometheusMetrics es una métrica para contabilizar el número de búsquedas en caché exitosas y fallidas.
	CachePrometheusMetrics struct {
		cacheHits      *prometheus.CounterVec
		cacheMisses    *prometheus.CounterVec
		cacheSize      prometheus.Gauge
		cacheEvictions *prometheus.CounterVec
		cacheRequests  *prometheus.CounterVec
		cacheSetOps    *prometheus.CounterVec
		cacheGetOps    *prometheus.CounterVec
		latencyGet     *prometheus.HistogramVec
		latencySet     *prometheus.HistogramVec
	}
)

// NewCacheMetrics crea una nueva instancia de CacheMetrics.
func NewCacheMetrics() CacheMetrics {
	return &CachePrometheusMetrics{
		cacheHits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Número total de aciertos en caché",
		}, []string{"cache_type"}),
		cacheMisses: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Número total de fallos en caché",
		}, []string{"cache_type"}),
		cacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cache_size",
			Help: "Tamaño actual del caché",
		}),
		cacheEvictions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_evictions_total",
			Help: "Número total de eliminaciones de caché debido a la expiración",
		}, []string{"cache_type"}),
		cacheRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_requests_total",
			Help: "Número total de solicitudes de caché",
		}, []string{"cache_type"}),
		cacheSetOps: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_set_operations_total",
			Help: "Número total de operaciones de establecimiento en caché",
		}, []string{"cache_type"}),
		cacheGetOps: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_get_operations_total",
			Help: "Número total de operaciones de obtención en caché",
		}, []string{"cache_type"}),
		latencyGet: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "custom_cache_get_latency_seconds",
			Help:    "Latencia de las operaciones de obtención en caché",
			Buckets: prometheus.DefBuckets,
		}, []string{"cache_type"}),
		latencySet: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "custom_cache_set_latency_seconds",
			Help:    "Latencia de las operaciones de establecimiento en caché",
			Buckets: prometheus.DefBuckets,
		}, []string{"cache_type"}),
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

func (c *CachePrometheusMetrics) IncHits(cacheType string) {
	c.cacheHits.WithLabelValues(cacheType).Inc()
}

func (c *CachePrometheusMetrics) IncMisses(cacheType string) {
	c.cacheMisses.WithLabelValues(cacheType).Inc()
}

func (c *CachePrometheusMetrics) SetCacheSize(size float64) {
	c.cacheSize.Set(size)
}

func (c *CachePrometheusMetrics) IncEvictions(cacheType string) {
	c.cacheEvictions.WithLabelValues(cacheType).Inc()
}

func (c *CachePrometheusMetrics) IncRequests(cacheType string) {
	c.cacheRequests.WithLabelValues(cacheType).Inc()
}

func (c *CachePrometheusMetrics) IncSetOperations(cacheType string) {
	c.cacheSetOps.WithLabelValues(cacheType).Inc()
}

func (c *CachePrometheusMetrics) IncGetOperations(cacheType string) {
	c.cacheGetOps.WithLabelValues(cacheType).Inc()
}

func (c *CachePrometheusMetrics) IncLatencyGet(cacheType string, duration time.Duration) {
	c.latencyGet.WithLabelValues(cacheType).Observe(duration.Seconds())
}

func (c *CachePrometheusMetrics) IncLatencySet(cacheType string, duration time.Duration) {
	c.latencySet.WithLabelValues(cacheType).Observe(duration.Seconds())
}
