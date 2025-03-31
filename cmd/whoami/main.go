package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	grpcserver "whoami-server/cmd/whoami/internal/servers/grpc"
	httpserver "whoami-server/cmd/whoami/internal/servers/http"
	"whoami-server/internal/config"
)

var historyServiceAddr = "localhost:50053"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := config.GetDefault("whoami")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Postgres.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	go func() {
		if err := grpcserver.Start(pool, cfg.Grpc.GetAddr(), historyServiceAddr); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	go func() {
		if err := httpserver.Start(ctx, cfg.Grpc.GetAddr(), cfg.Http.GetAddr()); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
}
