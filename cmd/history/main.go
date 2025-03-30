package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/history/internal/servers/grpc"
)

var (
	port    uint16 = 50053
	connStr string = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=prefer"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	go func() {
		if err := grpc.Start(pool, fmt.Sprintf(":%d", port)); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
}
