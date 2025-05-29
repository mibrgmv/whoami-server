package grpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"time"
	"whoami-server/cmd/whoami/internal/services/question"
	questiongrpc "whoami-server/cmd/whoami/internal/services/question/grpc"
	questionpg "whoami-server/cmd/whoami/internal/services/question/postgresql"
	"whoami-server/cmd/whoami/internal/services/quiz"
	quizgrpc "whoami-server/cmd/whoami/internal/services/quiz/grpc"
	quizpg "whoami-server/cmd/whoami/internal/services/quiz/postgresql"
	redisservice "whoami-server/internal/cache/redis"
	sharedInterceptors "whoami-server/internal/grpc/interceptors"
	"whoami-server/internal/grpc/interceptors/metadata"
	questionpb "whoami-server/protogen/golang/question"
	quizpb "whoami-server/protogen/golang/quiz"
)

type Server struct {
	grpcServer     *grpc.Server
	questionServer *questiongrpc.QuestionService
}

func authSkip(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo" &&
		c.FullMethod() != "/quiz.QuizService/BatchGetQuizzes"
}

func NewServer(pool *pgxpool.Pool, redisClient *redis.Client, redisTTL time.Duration, historyServiceAddr string) (*Server, error) {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	interceptorCfg := sharedInterceptors.NewConfig(logger)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			append(
				interceptorCfg.BuildUnaryInterceptors(),
				selector.UnaryServerInterceptor(metadata.UnaryInterceptor, selector.MatchFunc(authSkip)),
			)...,
		),
		grpc.ChainStreamInterceptor(
			append(
				interceptorCfg.BuildStreamInterceptors(),
				selector.StreamServerInterceptor(metadata.StreamInterceptor, selector.MatchFunc(authSkip)),
			)...,
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
	return &Server{
		grpcServer:     s,
		questionServer: questionServer,
	}, nil
}

func (s *Server) Start(addr string) error {
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

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()

	if s.questionServer != nil {
		if err := s.questionServer.Close(); err != nil {
			log.Printf("Error closing history service connection: %v", err)
		}
	}

	log.Println("gRPC server stopped")
}
