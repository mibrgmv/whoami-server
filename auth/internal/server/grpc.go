package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"whoami-server/auth/internal/config"
	authgrpc "whoami-server/auth/internal/grpc"
	"whoami-server/auth/internal/service"
	authv1 "whoami-server/auth/pkg/protogen/auth/v1"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/keycloak"
	"whoami-server/libs/logging"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	logger := logging.NewLogger("auth-service")
	kc := keycloak.NewClient(&cfg.Keycloak)
	authService := service.NewAuthService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
	)

	authGrpcServer := authgrpc.NewAuthServiceServer(authService)
	authv1.RegisterAuthServiceServer(server, authGrpcServer)

	reflection.Register(server)
	return server
}
