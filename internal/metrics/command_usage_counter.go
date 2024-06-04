package metrics

import "github.com/prometheus/client_golang/prometheus"

type CommandUsageCounter struct {
	counterVec *prometheus.CounterVec
}

func NewCommandUsageCounter() *CommandUsageCounter {
	return &CommandUsageCounter{
		counterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "command_usage_total",
			Help: "Número total de veces que se utilizan comandos, etiquetados por comando y servidor",
		},
			[]string{"command"},
		),
	}
}

func (c *CommandUsageCounter) Describe(ch chan<- *prometheus.Desc) {
	c.counterVec.Describe(ch)
}

func (c *CommandUsageCounter) Collect(ch chan<- prometheus.Metric) {
	c.counterVec.Collect(ch)
}

func (c *CommandUsageCounter) Inc(labels ...string) {
	c.counterVec.WithLabelValues(labels...).Inc()
}
