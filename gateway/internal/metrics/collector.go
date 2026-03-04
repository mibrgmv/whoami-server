package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Collector struct {
	registry *prometheus.Registry

	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPInflightRequests prometheus.Gauge
	AuthFailuresTotal    *prometheus.CounterVec
}

func NewMetricsCollector() *Collector {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	c := &Collector{
		registry: registry,

		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"method", "path"},
		),

		HTTPInflightRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gateway_http_inflight_requests",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		AuthFailuresTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_auth_failures_total",
				Help: "Total authentication failures",
			},
			[]string{"reason"},
		),
	}

	registry.MustRegister(
		c.HTTPRequestsTotal,
		c.HTTPRequestDuration,
		c.HTTPInflightRequests,
		c.AuthFailuresTotal,
	)

	return c
}

func (c *Collector) Handler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{})
}

func (c *Collector) RecordRequest(method, path, status string, duration float64) {
	c.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	c.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func (c *Collector) RecordAuthFailure(reason string) {
	c.AuthFailuresTotal.WithLabelValues(reason).Inc()
}

func (c *Collector) IncInflight() {
	c.HTTPInflightRequests.Inc()
}

func (c *Collector) DecInflight() {
	c.HTTPInflightRequests.Dec()
}
