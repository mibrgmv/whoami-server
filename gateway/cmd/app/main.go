package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appcfg "gordle/gateway/internal/config"
	"gordle/gateway/internal/metrics"
	"gordle/gateway/internal/server"
	"gordle/libs/config"
	"gordle/libs/keycloak"
	"gordle/libs/logging"
	"gordle/libs/storage/redis"
)

func main() {
	logger := logging.NewLogger("gateway")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths(".").
		WithEnvFiles("../../.env").
		Load(&cfg)

	if err != nil {
		logger.Error("failed to read gateway config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisClient, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		logger.Error("failed to create Redis client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("connected to Redis")

	keycloakClient := keycloak.NewClient(&cfg.Keycloak)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collector := metrics.NewMetricsCollector()
	metricsServer := server.NewMetricsServer(&cfg, collector, logger)
	go func() {
		err := metricsServer.Start()
		if err != nil {
			logger.Error("metrics server failed to serve", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	s, err := server.NewHttpServer(ctx, cfg, collector, logger, redisClient.Raw(), keycloakClient)
	if err != nil {
		logger.Error("failed to create HTTP server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		logger.Info("gateway HTTP server starting", slog.String("addr", s.Addr))
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to serve HTTP", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	sig := <-quit
	logger.Info("shutting down", slog.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err := metricsServer.Stop(shutdownCtx); err != nil {
		logger.Error("metrics server forced to shutdown", slog.String("error", err.Error()))
	}

	if err := s.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	logger.Info("gateway server exited gracefully")
}
