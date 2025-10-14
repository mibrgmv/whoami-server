package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	historygrpc "github.com/mibrgmv/whoami-server/history/internal/grpc"
	historyv1 "github.com/mibrgmv/whoami-server/history/internal/protogen/history/v1"
	"github.com/mibrgmv/whoami-server/history/internal/repository/postgres"
	"github.com/mibrgmv/whoami-server/history/internal/service"
	"github.com/mibrgmv/whoami-server/shared/grpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	grpcServer *grpc.Server
}

func NewGrpcServer(pool *pgxpool.Pool) GrpcServer {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	s := grpc.NewServer(
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

	historyRepo := postgres.NewHistoryRepository(pool)
	historyService := service.NewHistoryService(historyRepo)
	historyGrpc := historygrpc.NewHistoryServiceServer(historyService)
	historyv1.RegisterHistoryServiceServer(s, historyGrpc)

	reflection.Register(s)

	return GrpcServer{
		grpcServer: s,
	}
}

func (s *GrpcServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	log.Println("Serving gRPC on", lis.Addr())
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *GrpcServer) Stop() {
	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
