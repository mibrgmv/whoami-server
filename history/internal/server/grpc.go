package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	historygrpc "whoami-server/history/internal/grpc"
	"whoami-server/history/internal/repository/postgres"
	"whoami-server/history/internal/service"
	historyv1 "whoami-server/history/pkg/protogen/history/v1"
	libsgrpc "whoami-server/libs/grpc"
)

type GrpcServer struct {
	grpcServer *grpc.Server
}

func NewGrpcServer(pool *pgxpool.Pool) GrpcServer {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			append(
				libsgrpc.DefaultUnaryInterceptors(logger),
				libsgrpc.UnaryMetadataInterceptor(),
			)...,
		),
		grpc.ChainStreamInterceptor(
			append(
				libsgrpc.DefaultStreamInterceptors(logger),
				libsgrpc.StreamMetadataInterceptor(),
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
