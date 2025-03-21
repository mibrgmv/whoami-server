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
	"whoami-server/cmd/whoami/internal/services/question"
	grpcQuestion "whoami-server/cmd/whoami/internal/services/question/grpc"
	pgquestion "whoami-server/cmd/whoami/internal/services/question/postgresql"
	"whoami-server/cmd/whoami/internal/services/quiz"
	pgquiz "whoami-server/cmd/whoami/internal/services/quiz/postgresql"
	pbquestion "whoami-server/protogen/golang/question"
	pbquiz "whoami-server/protogen/golang/quiz"
)

type Server struct {
	server *grpc.Server
	config *Config
	pool   *pgxpool.Pool
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

	s.pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	if err := s.pool.Ping(context.Background()); err != nil {
		s.pool.Close()
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	s.server = grpc.NewServer()
	quizRepo := pgquiz.NewRepository(s.pool)
	questionRepo := pgquestion.NewRepository(s.pool)
	quizService := quiz.NewService(quizRepo)
	questionService := question.NewService(questionRepo)
	questionServiceGrpc, err := grpcQuestion.NewService(questionService, "localhost:50052")
	if err != nil {
		log.Fatalf("Failed to create question grpc service: %v", err)
	}

	pbquiz.RegisterQuizServiceServer(s.server, quizService)
	pbquestion.RegisterQuestionServiceServer(s.server, questionServiceGrpc)
	reflection.Register(s.server)

	fmt.Printf("Starting gRPC server on port %d\n", s.config.Server.Port)
	if err := s.server.Serve(listener); err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		fmt.Println("Stopping gRPC server...")
		s.server.GracefulStop()
		fmt.Println("gRPC server stopped")
	}

	if s.pool != nil {
		fmt.Println("Closing database connection pool...")
		s.pool.Close()
		fmt.Println("Database connection pool closed")
	}
	// todo need to close grpc conns
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
