package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	quizgrpc "github.com/mibrgmv/whoami-server/quiz/internal/grpc"
	questionv1 "github.com/mibrgmv/whoami-server/quiz/internal/protogen/question/v1"
	quizv1 "github.com/mibrgmv/whoami-server/quiz/internal/protogen/quiz/v1"
	"github.com/mibrgmv/whoami-server/quiz/internal/repository/postgres"
	"github.com/mibrgmv/whoami-server/quiz/internal/service"
	"github.com/mibrgmv/whoami-server/shared/grpc/interceptor"
	"github.com/mibrgmv/whoami-server/shared/storage/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	grpcServer     *grpc.Server
	questionServer quizgrpc.QuestionServer
}

func NewGrpcServer(pool *pgxpool.Pool, redisClient *redis.Client, historyServiceAddr string) (*GrpcServer, error) {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(interceptor.DefaultStreamInterceptors(logger)...),
	)

	quizRepo := postgres.NewQuizRepository(pool)
	quizService := service.NewQuizService(quizRepo)
	quizServer := quizgrpc.NewQuizServer(quizService)
	quizv1.RegisterQuizServiceServer(s, quizServer)

	questionRepo := postgres.NewQuestionRepository(pool)
	questionService := service.NewQuestionService(questionRepo, redisClient)
	questionServer, err := quizgrpc.NewQuestionServer(questionService, quizService, historyServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create question service: %w", err)
	}
	questionv1.RegisterQuestionServiceServer(s, questionServer)

	reflection.Register(s)
	return &GrpcServer{
		grpcServer: s,
	}, nil
}

func (s *GrpcServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Println("Serving gRPC on", lis.Addr())
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *GrpcServer) Stop() {
	s.grpcServer.GracefulStop()

	if s.questionServer != nil {
		if err := s.questionServer.Close(); err != nil {
			log.Printf("Error closing history service connection: %v", err)
		}
	}

	log.Println("gRPC server stopped")
}
