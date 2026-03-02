package server

import (
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	grpc2 "whoami-server/libs/grpc"
	"whoami-server/libs/keycloak"
	"whoami-server/user/internal/config"
	usergrpc "whoami-server/user/internal/grpc"
	"whoami-server/user/internal/service"
	userv1 "whoami-server/user/pkg/protogen/user/v1"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	kc := keycloak.NewClient(&cfg.Keycloak)
	userService := service.NewUserService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			append(
				grpc2.DefaultUnaryInterceptors(logger),
				grpc2.UnaryMetadataInterceptor(),
			)...,
		),
		grpc.ChainStreamInterceptor(
			append(
				grpc2.DefaultStreamInterceptors(logger),
				grpc2.StreamMetadataInterceptor(),
			)...,
		),
	)

	userGrpcServer := usergrpc.NewUserServiceServer(userService)
	userv1.RegisterUserServiceServer(server, userGrpcServer)

	reflection.Register(server)
	return server
}
