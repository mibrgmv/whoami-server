package grpc

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"whoami-server/cmd/whoami/internal/services/question"
	questiongrpc "whoami-server/cmd/whoami/internal/services/question/grpc"
	questionpg "whoami-server/cmd/whoami/internal/services/question/postgresql"
	"whoami-server/cmd/whoami/internal/services/quiz"
	quizgrpc "whoami-server/cmd/whoami/internal/services/quiz/grpc"
	quizpg "whoami-server/cmd/whoami/internal/services/quiz/postgresql"
	"whoami-server/internal/jwt"
	questionpb "whoami-server/protogen/golang/question"
	quizpb "whoami-server/protogen/golang/quiz"
)

func NewServer(pool *pgxpool.Pool, historyServiceAddr string) (*grpc.Server, error) {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(jwt.AuthUnaryInterceptor),
		grpc.StreamInterceptor(jwt.AuthStreamInterceptor),
	)

	quizRepo := quizpg.NewRepository(pool)
	quizService := quiz.NewService(quizRepo)
	quizServer := quizgrpc.NewService(quizService)
	quizpb.RegisterQuizServiceServer(s, quizServer)

	questionRepo := questionpg.NewRepository(pool)
	questionService := question.NewService(questionRepo)
	questionServer, err := questiongrpc.NewService(questionService, historyServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create question service: %w", err)
	}
	questionpb.RegisterQuestionServiceServer(s, questionServer)

	reflection.Register(s)
	return s, nil
}

func Start(pool *pgxpool.Pool, addr string, historyServiceAddr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s, err := NewServer(pool, historyServiceAddr)
	if err != nil {
		return err
	}

	log.Println("Serving gRPC on", addr)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
