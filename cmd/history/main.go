package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/history/internal/servers/grpc"
	"whoami-server/internal/postgres"
)

type Config struct {
	Grpc     grpc.Config     `mapstructure:"history-grpc"`
	Postgres postgres.Config `mapstructure:"postgres"`
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
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

	go func() {
		if err := grpc.Start(pool, fmt.Sprintf(":%d", cfg.Grpc.Port)); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
}
