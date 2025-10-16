package server

import (
	"log"
	"os"

	"github.com/mibrgmv/whoami-server/shared/grpc/interceptor"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
	"github.com/mibrgmv/whoami-server/user/internal/config"
	usergrpc "github.com/mibrgmv/whoami-server/user/internal/grpc"
	userv1 "github.com/mibrgmv/whoami-server/user/internal/protogen/user/v1"
	"github.com/mibrgmv/whoami-server/user/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(cfg config.Config) *grpc.Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	kc := keycloak.NewClient(&cfg.Keycloak)
	userService := service.NewUserService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			append(
				interceptor.DefaultUnaryInterceptors(logger),
				interceptor.UnaryMetadataInterceptor(),
			)...,
		),
		grpc.ChainStreamInterceptor(
			append(
				interceptor.DefaultStreamInterceptors(logger),
				interceptor.StreamMetadataInterceptor(),
			)...,
		),
	)

	userGrpcServer := usergrpc.NewUserServiceServer(userService)
	userv1.RegisterUserServiceServer(server, userGrpcServer)

	reflection.Register(server)
	return server
}
