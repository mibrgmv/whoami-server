package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/keycloak"
	"whoami-server/libs/logging"
	"whoami-server/user/internal/config"
	usergrpc "whoami-server/user/internal/grpc"
	"whoami-server/user/internal/service"
	userv1 "whoami-server/user/pkg/protogen/user/v1"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	logger := logging.NewLogger("user-service")
	kc := keycloak.NewClient(&cfg.Keycloak)
	userService := service.NewUserService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
	)

	userGrpcServer := usergrpc.NewUserServiceServer(userService)
	userv1.RegisterUserServiceServer(server, userGrpcServer)

	reflection.Register(server)
	return server
}
