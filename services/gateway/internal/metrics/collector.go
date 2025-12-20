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

	BackendRequestsTotal    *prometheus.CounterVec
	BackendRequestDuration  *prometheus.HistogramVec
	BackendInflightRequests *prometheus.GaugeVec
	BackendConnectionErrors *prometheus.CounterVec
}

func NewMetricsCollector() *Collector {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	gc := &Collector{
		registry: registry,

		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_requests_total",
				Help: "Total HTTP requests received by gateway (server-side)",
			},
			[]string{"method", "path", "status"},
		),

		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP request duration in gateway (server-side)",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),

		HTTPInflightRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gateway_http_inflight_requests",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		BackendRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_backend_requests_total",
				Help: "Total requests sent from gateway to backend services (client-side)",
			},
			[]string{"service", "method", "status"},
		),

		BackendRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_backend_request_duration_seconds",
				Help:    "Duration of requests to backend services including network (client-side)",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"service", "method"},
		),

		BackendInflightRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_backend_inflight_requests",
				Help: "Number of outgoing requests to backend services currently in flight",
			},
			[]string{"service"},
		),

		BackendConnectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_backend_connection_errors_total",
				Help: "Total connection/network errors when calling backend services",
			},
			[]string{"service", "error_type"},
		),
	}

	registry.MustRegister(
		gc.HTTPRequestsTotal,
		gc.HTTPRequestDuration,
		gc.HTTPInflightRequests,
		gc.BackendRequestsTotal,
		gc.BackendRequestDuration,
		gc.BackendInflightRequests,
		gc.BackendConnectionErrors,
	)

	return gc
}

func (gc *Collector) Registry() *prometheus.Registry {
	return gc.registry
}

func (gc *Collector) Handler() http.Handler {
	return promhttp.HandlerFor(gc.registry, promhttp.HandlerOpts{})
}

func (gc *Collector) RecordBackendRequest(service, method, status string, duration float64) {
	gc.BackendRequestsTotal.WithLabelValues(service, method, status).Inc()
	gc.BackendRequestDuration.WithLabelValues(service, method).Observe(duration)
}

func (gc *Collector) RecordBackendError(service, errorType string) {
	gc.BackendConnectionErrors.WithLabelValues(service, errorType).Inc()
}

func (gc *Collector) IncBackendInflight(service string) {
	gc.BackendInflightRequests.WithLabelValues(service).Inc()
}

func (gc *Collector) DecBackendInflight(service string) {
	gc.BackendInflightRequests.WithLabelValues(service).Dec()
}
