package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	quizzesconfig "whoami-server/cmd/quizzes/internal/config"
	"whoami-server/cmd/quizzes/internal/servers/grpc"
	"whoami-server/internal/cache/redis"
	"whoami-server/internal/config"
	"whoami-server/internal/tools"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cfg quizzesconfig.Config
	if err := config.LoanConfig(&cfg); err != nil {
		log.Fatalf("failed to read quizzes service config: %v", err)
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

	if err := tools.MigrateUp("migrations", "quizzes_service_schema_migrations", pool); err != nil {
		log.Fatalf("failed to migrate up: %v", err)
	}

	client, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	s, err := grpc.NewServer(pool, client, cfg.Redis.GetTTL(), cfg.HistoryService.GetAddr())
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

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
