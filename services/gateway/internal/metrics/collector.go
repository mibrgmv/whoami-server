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
	HTTPResponseSize     *prometheus.HistogramVec

	BackendRequestsTotal    *prometheus.CounterVec
	BackendRequestDuration  *prometheus.HistogramVec
	BackendInflightRequests *prometheus.GaugeVec
	BackendConnectionErrors *prometheus.CounterVec

	APIRequestsTotal *prometheus.CounterVec
	APIErrorsTotal   *prometheus.CounterVec
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
				Help: "Total HTTP requests received by gateway",
			},
			[]string{"method", "endpoint", "status"},
		),

		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),

		HTTPInflightRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gateway_http_inflight_requests",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		HTTPResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),

		BackendRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_backend_requests_total",
				Help: "Total requests sent from gateway to backend services",
			},
			[]string{"backend", "method", "status"},
		),

		BackendRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_backend_request_duration_seconds",
				Help:    "Duration of requests to backend services",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"backend", "method"},
		),

		BackendInflightRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_backend_inflight_requests",
				Help: "Number of outgoing requests to backend services currently in flight",
			},
			[]string{"backend"},
		),

		BackendConnectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_backend_connection_errors_total",
				Help: "Total connection/network errors when calling backend services",
			},
			[]string{"backend", "error_type"},
		),

		APIRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_api_requests_total",
				Help: "Total requests by API endpoint",
			},
			[]string{"api", "operation"},
		),

		APIErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_api_errors_total",
				Help: "Total errors by API endpoint",
			},
			[]string{"api", "error_type"},
		),
	}

	registry.MustRegister(
		gc.HTTPRequestsTotal,
		gc.HTTPRequestDuration,
		gc.HTTPInflightRequests,
		gc.HTTPResponseSize,
		gc.BackendRequestsTotal,
		gc.BackendRequestDuration,
		gc.BackendInflightRequests,
		gc.BackendConnectionErrors,
		gc.APIRequestsTotal,
		gc.APIErrorsTotal,
	)

	return gc
}

func (gc *Collector) Registry() *prometheus.Registry {
	return gc.registry
}

func (gc *Collector) Handler() http.Handler {
	return promhttp.HandlerFor(gc.registry, promhttp.HandlerOpts{})
}

func (gc *Collector) RecordHTTPRequest(method, endpoint, status string, duration, responseSize float64) {
	gc.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	gc.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	if responseSize > 0 {
		gc.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(responseSize)
	}
}

func (gc *Collector) RecordBackendRequest(backend, method, status string, duration float64) {
	gc.BackendRequestsTotal.WithLabelValues(backend, method, status).Inc()
	gc.BackendRequestDuration.WithLabelValues(backend, method).Observe(duration)
}

func (gc *Collector) RecordBackendError(backend, errorType string) {
	gc.BackendConnectionErrors.WithLabelValues(backend, errorType).Inc()
}

func (gc *Collector) RecordAPIRequest(api, operation string) {
	gc.APIRequestsTotal.WithLabelValues(api, operation).Inc()
}

func (gc *Collector) RecordAPIError(api, errorType string) {
	gc.APIErrorsTotal.WithLabelValues(api, errorType).Inc()
}

func (gc *Collector) IncHTTPInflight() {
	gc.HTTPInflightRequests.Inc()
}

func (gc *Collector) DecHTTPInflight() {
	gc.HTTPInflightRequests.Dec()
}

func (gc *Collector) IncBackendInflight(backend string) {
	gc.BackendInflightRequests.WithLabelValues(backend).Inc()
}

func (gc *Collector) DecBackendInflight(backend string) {
	gc.BackendInflightRequests.WithLabelValues(backend).Dec()
}
