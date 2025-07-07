package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	gatewayconfig "whoami-server/cmd/gateway/internal/config"
	"whoami-server/cmd/gateway/internal/servers/http"
	"whoami-server/internal/config"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg gatewayconfig.Config
	if err := config.LoanConfig(&cfg); err != nil {
		log.Fatalf("failed to read gateway config: %v", err)
	}

	grpcAddresses := map[string]string{
		"quizzes": cfg.QuizzesService.GetAddr(),
		"users":   cfg.UsersService.GetAddr(),
		"history": cfg.HistoryService.GetAddr(),
	}

	go func() {
		if err := http.Start(ctx, grpcAddresses, cfg.HTTP); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
