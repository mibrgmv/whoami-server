package server

import (
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/kafka"
	"whoami-server/libs/logging"
	"whoami-server/libs/storage/redis"
	quizgrpc "whoami-server/quiz/internal/grpc"
	"whoami-server/quiz/internal/repository/postgres"
	"whoami-server/quiz/internal/service"
	questionv1 "whoami-server/quiz/pkg/protogen/question/v1"
	quizv1 "whoami-server/quiz/pkg/protogen/quiz/v1"
)

type GrpcServer struct {
	grpcServer *grpc.Server
	producer   kafka.Producer
	logger     *slog.Logger
}

func NewGrpcServer(pool *pgxpool.Pool, redisClient *redis.Client, kafkaCfg *kafka.Config) *GrpcServer {
	logger := logging.NewLogger("quiz-service")

	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:  kafkaCfg.Brokers,
		ClientID: "quiz-service",
	})

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
	)

	quizRepo := postgres.NewQuizRepository(pool)
	quizService := service.NewQuizService(quizRepo)
	quizServer := quizgrpc.NewQuizServer(quizService)
	quizv1.RegisterQuizServiceServer(s, quizServer)

	questionRepo := postgres.NewQuestionRepository(pool)
	questionService := service.NewQuestionService(questionRepo, redisClient)
	questionServer := quizgrpc.NewQuestionServer(questionService, quizService, producer, kafkaCfg.Topics.QuizCompleted)
	questionv1.RegisterQuestionServiceServer(s, questionServer)

	reflection.Register(s)
	return &GrpcServer{
		grpcServer: s,
		producer:   producer,
		logger:     logger,
	}
}

func (s *GrpcServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("failed to listen", slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("serving gRPC", slog.String("addr", lis.Addr().String()))
	if err := s.grpcServer.Serve(lis); err != nil {
		s.logger.Error("failed to serve", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *GrpcServer) Stop() {
	s.grpcServer.GracefulStop()

	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			s.logger.Error("error closing Kafka producer", slog.String("error", err.Error()))
		}
	}

	s.logger.Info("gRPC server stopped")
}
