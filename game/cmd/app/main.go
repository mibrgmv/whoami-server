package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"gordle/libs/config"
	"gordle/libs/storage/redis"
	"gordle/libs/tools"

	appcfg "gordle/game/internal/config"
	"gordle/game/internal/seed"
	"gordle/game/internal/server"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths(".").
		WithEnvFiles("../../.env").
		Load(&cfg)

	if err != nil {
		log.Fatalf("failed to read game service config: %v", err)
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

	if err := tools.MigrateUp("migrations", "game_service_schema_migrations", pool); err != nil {
		log.Fatalf("failed to migrate up: %v", err)
	}

	if err := seed.Words(ctx, pool); err != nil {
		log.Fatalf("failed to seed words: %v", err)
	}

	redisClient, err := redis.NewClient(ctx, *cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	log.Println("Connected to Redis successfully")

	s := server.NewGrpcServer(pool, redisClient.Raw(), cfg.Kafka)

	go func() {
		if err := s.Start(cfg.Grpc.GetAddr()); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
	s.Stop()
}
