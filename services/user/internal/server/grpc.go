package server

import (
	"log"
	"os"

	sharedgrpc "github.com/mibrgmv/whoami-server/shared/grpc"
	"github.com/mibrgmv/whoami-server/shared/grpc/metadata"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
	"github.com/mibrgmv/whoami-server/user/internal/config"
	usergrpc "github.com/mibrgmv/whoami-server/user/internal/grpc"
	userpb "github.com/mibrgmv/whoami-server/user/internal/protogen/user"
	"github.com/mibrgmv/whoami-server/user/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(cfg config.Config) *grpc.Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	interceptorCfg := sharedgrpc.NewConfig(logger)
	kc := keycloak.NewClient(&cfg.Keycloak)
	userService := service.NewUserService(kc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			append(
				interceptorCfg.BuildUnaryInterceptors(),
				metadata.UnaryInterceptor(),
			)...,
		),
		grpc.ChainStreamInterceptor(
			append(
				interceptorCfg.BuildStreamInterceptors(),
				metadata.StreamInterceptor(),
			)...,
		),
	)

	userGrpcServer := usergrpc.NewUserServiceServer(userService)
	userpb.RegisterUserServiceServer(server, userGrpcServer)

	reflection.Register(server)
	return server
}
