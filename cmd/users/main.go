package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/users/internal/servers/grpc"
	"whoami-server/internal/cache/redis"
	"whoami-server/internal/config"
	rediscfg "whoami-server/internal/config/cache/redis"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := config.GetDefaultForService("users")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	redisCfg, err := rediscfg.LoadDefault()
	if err != nil {
		log.Fatalf("failed to read redis config: %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres.GetConnectionString())
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	client, err := redis.NewClient(ctx, redisCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	go func() {
		if err := grpc.Start(pool, client, redisCfg.GetTTL(), cfg.Grpc.GetAddr()); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
}
