package main

import (
	"context"
	"github.com/mibrgmv/whoami-server-shared/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	appcfg "whoami-server-gateway/internal/config"
	"whoami-server-gateway/internal/servers/http"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg appcfg.Config
	if err := config.LoanConfig(&cfg); err != nil {
		log.Fatalf("failed to read gateway config: %v", err)
	}

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
