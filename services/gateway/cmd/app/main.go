package main

import (
	"context"
	"log"
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

	go func() {
		if err := server.StartHTTP(ctx, cfg); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
