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

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := config.GetDefaultForService("gateway")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	quizzesCfg, err := config.GetDefaultForService("quizzes")
	if err != nil {
		log.Fatalf("failed to get quizzes config: %v", err)
	}

	usersCfg, err := config.GetDefaultForService("users")
	if err != nil {
		log.Fatalf("failed to get users config: %v", err)
	}

	historyCfg, err := config.GetDefaultForService("history")
	if err != nil {
		log.Fatalf("failed to get history config: %v", err)
	}

	grpcAddresses := map[string]string{
		"quizzes": quizzesCfg.Grpc.GetAddr(),
		"users":   usersCfg.Grpc.GetAddr(),
		"history": historyCfg.Grpc.GetAddr(),
	}

	go func() {
		if err := http.Start(ctx, grpcAddresses, cfg.Http); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
