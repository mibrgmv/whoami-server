package grpc

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"whoami-server/cmd/history/internal/services/history"
	pg "whoami-server/cmd/history/internal/services/history/postgresql"
	pb "whoami-server/protogen/golang/history"
)

func NewServer(pool *pgxpool.Pool) *grpc.Server {
	s := grpc.NewServer()
	repo := pg.NewRepository(pool)
	service := history.NewService(repo)
	pb.RegisterQuizCompletionHistoryServiceServer(s, service)
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
