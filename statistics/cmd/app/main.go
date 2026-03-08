package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/libs/config"
	"whoami-server/libs/tools"
	appcfg "whoami-server/statistics/internal/config"
	"whoami-server/statistics/internal/server"
)

func main() {
	ctx := context.Background()

	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths(".").
		WithEnvFiles("../../.env").
		Load(&cfg)

	if err != nil {
		log.Fatalf("failed to read statistics config: %v", err)
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

	if err := tools.MigrateUp("migrations", "statistics_service_schema_migrations", pool); err != nil {
		log.Fatalf("failed to migrate up: %v", err)
	}

	s := server.New(pool, cfg.Kafka)

	go func() {
		if err := s.Start(ctx, cfg.Grpc.GetAddr()); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	s.Stop()
}
