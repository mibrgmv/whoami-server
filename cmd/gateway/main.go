package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/gateway/internal"
	"whoami-server/cmd/gateway/internal/servers/http"
	"whoami-server/internal/config"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg internal.Config
	if err := config.LoanConfig(&cfg); err != nil {
		log.Fatalf("failed to read gateway config: %v", err)
	}

	log.Println("loaded keycloak config:", *cfg.Keycloak)

	go func() {
		if err := http.Start(ctx, cfg); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
