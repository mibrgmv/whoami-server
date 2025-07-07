package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	usersconfig "whoami-server/cmd/users/internal/config"
	"whoami-server/cmd/users/internal/servers/grpc"
	"whoami-server/internal/cache/redis"
	"whoami-server/internal/config"
	"whoami-server/internal/tools"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg usersconfig.Config
	if err := config.LoanDefault(&cfg); err != nil {
		log.Fatalf("failed to read postgres config: %v", err)
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

	if err := tools.MigrateUp("migrations", pool); err != nil {
		log.Fatalf("failed to migrate up: %v", err)
	}

	client, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	go func() {
		if err := grpc.Start(pool, client, cfg.Redis.GetTTL(), cfg.JWT, cfg.Grpc.GetAddr()); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
}
