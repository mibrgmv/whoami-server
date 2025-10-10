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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token in the format: Bearer <your-token-here>
func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths("internal/config").
		WithConfigName("default").
		WithConfigType("yaml").
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
