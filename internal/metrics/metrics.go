package metrics

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"regexp"
)

type RegistryMetric interface {
	Register(metric CustomMetric)
	RegisterCacheMetrics(cacheMetrics CacheMetrics)
	RegisterStandardMetrics()
	GetRegistry() *prometheus.Registry
}

type PrometheusRegistry struct {
	registry *prometheus.Registry
}

func NewPrometheusRegistry() *PrometheusRegistry {
	return &PrometheusRegistry{
		registry: prometheus.NewRegistry(),
	}
}

func (pr *PrometheusRegistry) RegisterCacheMetrics(cacheMetrics CacheMetrics) {
	pr.registry.MustRegister(cacheMetrics)
}

func (pr *PrometheusRegistry) Register(metric CustomMetric) {
	pr.registry.MustRegister(metric)
}

func (pr *PrometheusRegistry) RegisterStandardMetrics() {
	pr.registry.MustRegister(collectors.NewBuildInfoCollector())
	pr.registry.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
	))
}

func (pr *PrometheusRegistry) GetRegistry() *prometheus.Registry {
	return pr.registry
}

type HTTPMetricsServer interface {
	Start() error
}

type PrometheusHTTPServer struct {
	Address  string
	Registry RegistryMetric
}

func NewPrometheusHTTPServer(address string, registry RegistryMetric) *PrometheusHTTPServer {
	return &PrometheusHTTPServer{
		Address:  address,
		Registry: registry,
	}
}

func (s *PrometheusHTTPServer) Start() error {
	flag.Parse()

	s.Registry.RegisterStandardMetrics()

	http.Handle("/metrics", promhttp.HandlerFor(
		s.Registry.GetRegistry(),
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}))
	fmt.Println("Iniciando server de Prometheus HTTP metric...", s.Address)
	return http.ListenAndServe(s.Address, nil)
}
