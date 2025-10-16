package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appcfg "github.com/mibrgmv/whoami-server/gateway/internal/config"
	"github.com/mibrgmv/whoami-server/gateway/internal/server"
	"github.com/mibrgmv/whoami-server/shared/config"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths("internal/config").
		WithEnvFiles("../../.env").
		Load(&cfg)

	if err != nil {
		log.Fatalf("failed to read gateway config: %v", err)
	}

	s, err := server.NewHttpServer(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to create HTTP server:", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

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

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Gateway server exited gracefully")
}
