package server

import (
	"context"
	"log/slog"
	"net/http"

	appcfg "whoami-server/gateway/internal/config"
	"whoami-server/gateway/internal/metrics"
)

type MetricsServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewMetricsServer(cfg *appcfg.Config, collector *metrics.Collector, logger *slog.Logger) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", collector.Handler())

	return &MetricsServer{
		server: &http.Server{
			Addr:    cfg.Metrics.GetAddr(),
			Handler: mux,
		},
		logger: logger,
	}
}

func (s *MetricsServer) Start() error {
	s.logger.Info("metrics server starting", slog.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *MetricsServer) Stop(ctx context.Context) error {
	s.logger.Info("shutting down metrics server")
	return s.server.Shutdown(ctx)
}
