package server

import (
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/kafka"
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
}

func NewGrpcServer(pool *pgxpool.Pool, redisClient *redis.Client, kafkaCfg *kafka.Config) *GrpcServer {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:  kafkaCfg.Brokers,
		ClientID: "quiz-service",
	})

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
	}
}

func (s *GrpcServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Serving gRPC on", lis.Addr())
	if err := s.grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}

func (s *GrpcServer) Stop() {
	s.grpcServer.GracefulStop()

	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}

	log.Println("gRPC server stopped")
}
