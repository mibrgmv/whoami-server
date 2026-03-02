package server

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	appcfg "whoami-server/gateway/internal/config"
	"whoami-server/gateway/internal/metrics"
)

type MetricsServer struct {
	server    *http.Server
	collector *metrics.Collector
}

func NewMetricsServer(cfg *appcfg.Config, collector *metrics.Collector) *MetricsServer {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(
		collector.Registry(),
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          collector.Registry(),
		},
	))

	server := &http.Server{
		Addr:    cfg.Metrics.GetAddr(),
		Handler: mux,
	}

	return &MetricsServer{
		server:    server,
		collector: collector,
	}
}

func (s *MetricsServer) Start() error {
	log.Printf("Gateway metrics server starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *MetricsServer) Stop(ctx context.Context) error {
	log.Println("Shutting down gateway metrics server...")
	return s.server.Shutdown(ctx)
}
