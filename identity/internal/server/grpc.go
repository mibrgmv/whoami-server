package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gordle/identity/internal/config"
	identitygrpc "gordle/identity/internal/grpc"
	"gordle/identity/internal/service"
	authv1 "gordle/identity/pkg/protogen/auth/v1"
	userv1 "gordle/identity/pkg/protogen/user/v1"
	libsgrpc "gordle/libs/grpc"
	"gordle/libs/keycloak"
	"gordle/libs/logging"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	logger := logging.NewLogger("identity-service")
	kc := keycloak.NewClient(&cfg.Keycloak)

	authService := service.NewAuthService(kc)
	userService := service.NewUserService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
	)

	authGrpcServer := identitygrpc.NewAuthServiceServer(authService)
	userGrpcServer := identitygrpc.NewUserServiceServer(userService)

	authv1.RegisterAuthServiceServer(server, authGrpcServer)
	userv1.RegisterUserServiceServer(server, userGrpcServer)

	reflection.Register(server)
	return server
}
