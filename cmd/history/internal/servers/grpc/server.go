package grpc

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"whoami-server/cmd/history/internal/services/history"
	pg "whoami-server/cmd/history/internal/services/history/postgresql"
	pb "whoami-server/protogen/golang/history"
)

type Server struct {
	server *grpc.Server
	config *Config
}

func NewServer(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %v", s.config.Server.Port, err)
	}

	connString := s.config.GetPostgresConnectionString()
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse connStr: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	s.server = grpc.NewServer()
	repo := pg.NewRepository(pool)
	service := history.NewService(repo)
	pb.RegisterQuizCompletionHistoryServiceServer(s.server, service)
	reflection.Register(s.server)

	go func() {
		fmt.Printf("Starting gRPC server on port %d\n", s.config.Server.Port)
		if err := s.server.Serve(listener); err != nil {
			fmt.Printf("Failed to serve: %v\n", err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		fmt.Println("Stopping gRPC server...")
		s.server.GracefulStop()
		fmt.Println("gRPC server stopped")
	}
}

func (s *Server) RunAndWait() error {
	if err := s.Start(); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	s.Stop()
	return nil
}
