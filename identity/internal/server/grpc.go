package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"whoami-server/identity/internal/config"
	identitygrpc "whoami-server/identity/internal/grpc"
	"whoami-server/identity/internal/service"
	authv1 "whoami-server/identity/pkg/protogen/auth/v1"
	userv1 "whoami-server/identity/pkg/protogen/user/v1"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/keycloak"
	"whoami-server/libs/logging"
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
