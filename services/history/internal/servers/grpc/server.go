package grpc

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/mibrgmv/whoami-server/history/internal/protogen/history"
	"github.com/mibrgmv/whoami-server/history/internal/services/history"
	historygrpc "github.com/mibrgmv/whoami-server/history/internal/services/history/grpc"
	pg "github.com/mibrgmv/whoami-server/history/internal/services/history/postgresql"
	sharedInterceptors "github.com/mibrgmv/whoami-server/shared/grpc"
	"github.com/mibrgmv/whoami-server/shared/grpc/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(pool *pgxpool.Pool) *grpc.Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	interceptorCfg := sharedInterceptors.NewConfig(logger)

	s := grpc.NewServer(
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

	repo := pg.NewRepository(pool)
	service := history.NewService(repo)
	grpcServer := historygrpc.NewService(service)
	pb.RegisterQuizCompletionHistoryServiceServer(s, grpcServer)
	reflection.Register(s)
	return s
}

func Start(pool *pgxpool.Pool, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := NewServer(pool)
	log.Println("Serving gRPC on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
