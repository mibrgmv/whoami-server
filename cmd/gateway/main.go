package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/gateway/internal/servers/http"
	"whoami-server/internal/config"
)

var grpcAddresses = map[string]string{
	"whoami": "localhost:50051",
	"users":  "localhost:50052",
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := config.GetDefault("gateway")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	go func() {
		if err := http.Start(ctx, grpcAddresses, cfg.Http.GetAddr()); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
