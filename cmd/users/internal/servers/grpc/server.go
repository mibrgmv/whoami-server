package grpc

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
	"whoami-server/cmd/users/internal/services/user"
	usergrpc "whoami-server/cmd/users/internal/services/user/grpc"
	pg "whoami-server/cmd/users/internal/services/user/postgresql"
	redisservice "whoami-server/internal/cache/redis"
	"whoami-server/internal/interceptors"
	userpb "whoami-server/protogen/golang/user"
)

func NewServer(pool *pgxpool.Pool, redisClient *redis.Client, redisTTL time.Duration) *grpc.Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.AuthUnaryInterceptor,
			interceptors.CacheControlUnaryInterceptor,
			interceptors.ETagUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptors.AuthStreamInterceptor,
			interceptors.CacheControlStreamInterceptor,
		),
	)

	redisService := redisservice.NewService(redisClient, redisTTL)
	repo := pg.NewRepository(pool)
	service := user.NewService(repo, redisService)
	server := usergrpc.NewUserService(service)
	userpb.RegisterUserServiceServer(s, server)
	reflection.Register(s)
	return s
}

func Start(pool *pgxpool.Pool, redis *redis.Client, ttl time.Duration, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := NewServer(pool, redis, ttl)
	log.Println("Serving gRPC on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
