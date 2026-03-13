package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gordle/iam/internal/config"
	iamgrpc "gordle/iam/internal/grpc"
	"gordle/iam/internal/service"
	authv1 "gordle/iam/pkg/protogen/auth/v1"
	userv1 "gordle/iam/pkg/protogen/user/v1"
	libsgrpc "gordle/libs/grpc"
	"gordle/libs/keycloak"
	"gordle/libs/logging"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	logger := logging.NewLogger("iam-service")
	kc := keycloak.NewClient(&cfg.Keycloak)

	authService := service.NewAuthService(kc, cfg.GuestSecret)
	userService := service.NewUserService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
	)

	authGrpcServer := iamgrpc.NewAuthServiceServer(authService)
	userGrpcServer := iamgrpc.NewUserServiceServer(userService)

	authv1.RegisterAuthServiceServer(server, authGrpcServer)
	userv1.RegisterUserServiceServer(server, userGrpcServer)

	reflection.Register(server)
	return server
}
