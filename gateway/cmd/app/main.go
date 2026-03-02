package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appcfg "whoami-server/gateway/internal/config"
	"whoami-server/gateway/internal/metrics"
	"whoami-server/gateway/internal/server"
	"whoami-server/libs/config"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths(".").
		WithEnvFiles("../../.env").
		Load(&cfg)

	if err != nil {
		log.Fatalf("failed to read gateway config: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collector := metrics.NewMetricsCollector()
	metricsServer := server.NewMetricsServer(&cfg, collector)
	go func() {
		err := metricsServer.Start()
		if err != nil {
			log.Fatal("metrics server failed to serve: ", err)
		}
	}()

	s, err := server.NewHttpServer(ctx, cfg, collector)
	if err != nil {
		log.Fatal("Failed to create HTTP server:", err)
	}

	go func() {
		log.Printf("Gateway HTTP server starting on %s", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to serve HTTP:", err)
		}
	}()

	sig := <-quit
	log.Printf("Shutting down, received signal: %v", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err := metricsServer.Stop(shutdownCtx); err != nil {
		log.Fatal("Metrics server forced to shutdown:", err)
	}

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Gateway server exited gracefully")
}
