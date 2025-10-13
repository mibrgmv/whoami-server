package server

import (
	"log"
	"os"

	"github.com/mibrgmv/whoami-server/auth/internal/config"
	authgrpc "github.com/mibrgmv/whoami-server/auth/internal/grpc"
	authv1 "github.com/mibrgmv/whoami-server/auth/internal/protogen/auth/v1"
	"github.com/mibrgmv/whoami-server/auth/internal/service"
	sharedgrpc "github.com/mibrgmv/whoami-server/shared/grpc"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(cfg config.Config) *grpc.Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	interceptorCfg := sharedgrpc.NewConfig(logger)
	kc := keycloak.NewClient(&cfg.Keycloak)
	authService := service.NewAuthService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptorCfg.BuildUnaryInterceptors()...,
		),
		grpc.ChainStreamInterceptor(
			interceptorCfg.BuildStreamInterceptors()...,
		),
	)

	authGrpcServer := authgrpc.NewAuthServiceServer(authService)
	authv1.RegisterAuthServiceServer(server, authGrpcServer)

	reflection.Register(server)
	return server
}
