package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Process struct {
	total    *prometheus.CounterVec
	duration *prometheus.HistogramVec
	current  *prometheus.GaugeVec
}

func NewProcess() *Process {
	m := &Process{}

	m.total = prometheus.NewCounterVec(prometheus.CounterOpts{ // счётчик, который увеличивается, но
		// не уменьшается в течение жизни процесса.
		Name: "entity_processed_total",
		Help: "Total number of processed entities",
	}, []string{"name", "status"})
	prometheus.MustRegister(m.total) // Регистрирует метрику в глобальном реестре Prometheus

	m.duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{ //используется для наблюдения за распределением длительности.
		Name:    "entity_processed_duration",
		Help:    "Duration of processing entities (in seconds)",
		Buckets: buckets,
	}, []string{"name"})
	prometheus.MustRegister(m.duration) // Регистрирует метрику в глобальном реестре Prometheus

	m.current = prometheus.NewGaugeVec(prometheus.GaugeOpts{ //может увеличиваться и уменьшаться.
		Name: "entity_processing_current",
		Help: "Current number of processing entities",
	}, []string{"name"})
	prometheus.MustRegister(m.current) // Регистрирует метрику в глобальном реестре Prometheus

	return m
}

func (p *Process) Total(name string, status Status) {
	p.total.WithLabelValues(name, status.String()).Inc()
}

func (p *Process) TotalAdd(name string, status Status, counter int) {
	p.total.WithLabelValues(name, status.String()).Add(float64(counter))
}

func (p *Process) Duration(name string, startTime time.Time) {
	p.duration.WithLabelValues(name).Observe(time.Since(startTime).Seconds())
}

func (p *Process) Current(name string, value float64) {
	p.current.WithLabelValues(name).Set(value)
}
