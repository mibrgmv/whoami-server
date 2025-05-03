package grpc

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
	"whoami-server/cmd/whoami/internal/services/question"
	questiongrpc "whoami-server/cmd/whoami/internal/services/question/grpc"
	questionpg "whoami-server/cmd/whoami/internal/services/question/postgresql"
	"whoami-server/cmd/whoami/internal/services/quiz"
	quizgrpc "whoami-server/cmd/whoami/internal/services/quiz/grpc"
	quizpg "whoami-server/cmd/whoami/internal/services/quiz/postgresql"
	redisservice "whoami-server/internal/cache/redis"
	"whoami-server/internal/interceptors"
	questionpb "whoami-server/protogen/golang/question"
	quizpb "whoami-server/protogen/golang/quiz"
)

func NewServer(pool *pgxpool.Pool, redisClient *redis.Client, redisTTL time.Duration, historyServiceAddr string) (*grpc.Server, error) {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.AuthUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptors.AuthStreamInterceptor,
		),
	)

	redisService := redisservice.NewService(redisClient, redisTTL)
	quizRepo := quizpg.NewRepository(pool)
	quizService := quiz.NewService(quizRepo)
	quizServer := quizgrpc.NewService(quizService)
	quizpb.RegisterQuizServiceServer(s, quizServer)

	questionRepo := questionpg.NewRepository(pool)
	questionService := question.NewService(questionRepo, redisService)
	questionServer, err := questiongrpc.NewService(questionService, quizService, historyServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create question service: %w", err)
	}
	questionpb.RegisterQuestionServiceServer(s, questionServer)

	reflection.Register(s)
	return s, nil
}

func Start(pool *pgxpool.Pool, redis *redis.Client, ttl time.Duration, addr string, historyServiceAddr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s, err := NewServer(pool, redis, ttl, historyServiceAddr)
	if err != nil {
		return err
	}

	log.Println("Serving gRPC on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
