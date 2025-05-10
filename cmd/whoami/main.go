package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/whoami/internal/servers/grpc"
	"whoami-server/internal/cache/redis"
	"whoami-server/internal/config"
	jwtcfg "whoami-server/internal/config/auth/jwt"
	rediscfg "whoami-server/internal/config/cache/redis"
	"whoami-server/internal/tools/jwt"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := config.GetDefaultForService("whoami")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	historyCfg, err := config.GetDefaultForService("history")
	if err != nil {
		log.Fatalf("failed to get history config: %v", err)
	}

	redisCfg, err := rediscfg.LoadDefault()
	if err != nil {
		log.Fatalf("failed to read redis config: %v", err)
	}

	jwtCfg, err := jwtcfg.LoadDefault()
	if err != nil {
		log.Fatalf("failed to read jwt config: %v", err)
	}
	jwt.Init(jwtCfg)

	pool, err := pgxpool.New(context.Background(), cfg.Postgres.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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

	s, err := grpc.NewServer(pool, client, redisCfg.GetTTL(), historyCfg.Grpc.GetAddr())
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
