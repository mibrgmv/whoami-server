package grpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"time"
	authgrpc "whoami-server/cmd/users/internal/services/auth/grpc"
	"whoami-server/cmd/users/internal/services/auth/jwt"
	"whoami-server/cmd/users/internal/services/user"
	usergrpc "whoami-server/cmd/users/internal/services/user/grpc"
	pg "whoami-server/cmd/users/internal/services/user/postgresql"
	redisservice "whoami-server/internal/cache/redis"
	jwtcfg "whoami-server/internal/config/auth/jwt"
	sharedInterceptors "whoami-server/internal/grpc/interceptors"
	"whoami-server/internal/grpc/interceptors/metadata"
	authpb "whoami-server/protogen/golang/auth"
	userpb "whoami-server/protogen/golang/user"
)

func authSkip(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo" &&
		c.FullMethod() != "/auth.AuthorizationService/Login" &&
		c.FullMethod() != "/auth.AuthorizationService/RefreshToken" &&
		c.FullMethod() != "/auth.AuthorizationService/ValidateToken" &&
		c.FullMethod() != "/user.UserService/CreateUser"
}

func NewServer(pool *pgxpool.Pool, redisClient *redis.Client, redisTTL time.Duration, jwtCfg *jwtcfg.Config) *grpc.Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	interceptorCfg := sharedInterceptors.NewConfig(logger)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			append(
				interceptorCfg.BuildUnaryInterceptors(),
				selector.UnaryServerInterceptor(metadata.UnaryInterceptor, selector.MatchFunc(authSkip)),
			)...,
		),
		grpc.ChainStreamInterceptor(
			append(
				interceptorCfg.BuildStreamInterceptors(),
				selector.StreamServerInterceptor(metadata.StreamInterceptor, selector.MatchFunc(authSkip)),
			)...,
		),
	)

	redisService := redisservice.NewService(redisClient, redisTTL)
	repo := pg.NewRepository(pool)
	userService := user.NewService(repo, redisService)

	userGrpcServer := usergrpc.NewUserService(userService)
	userpb.RegisterUserServiceServer(s, userGrpcServer)

	jwtService := jwt.NewService(jwtCfg)
	authGrpcServer := authgrpc.NewService(userService, jwtService)
	authpb.RegisterAuthorizationServiceServer(s, authGrpcServer)

	reflection.Register(s)
	return s
}

func Start(pool *pgxpool.Pool, redis *redis.Client, ttl time.Duration, jwtCfg *jwtcfg.Config, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := NewServer(pool, redis, ttl, jwtCfg)
	log.Println("Serving gRPC on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
