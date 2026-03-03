package server

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"whoami-server/history/internal/consumer"
	historygrpc "whoami-server/history/internal/grpc"
	"whoami-server/history/internal/repository/postgres"
	"whoami-server/history/internal/service"
	historyv1 "whoami-server/history/pkg/protogen/history/v1"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/kafka"
)

type Server struct {
	grpcServer    *grpc.Server
	kafkaConsumer *kafka.Consumer
	cancelFunc    context.CancelFunc
}

func New(pool *pgxpool.Pool, kafkaCfg *kafka.Config) *Server {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	grpcSrv := grpc.NewServer(
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
	historyv1.RegisterHistoryServiceServer(grpcSrv, historyGrpc)

	reflection.Register(grpcSrv)

	handler := consumer.NewQuizCompletedHandler(historyService)
	kafkaConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: kafkaCfg.Brokers,
		GroupID: kafkaCfg.GroupID,
		Topic:   kafkaCfg.Topics.QuizCompleted,
		MaxWait: 10 * time.Second,
	}, handler.Handle)

	return &Server{
		grpcServer:    grpcSrv,
		kafkaConsumer: kafkaConsumer,
	}
}

func (s *Server) Start(ctx context.Context, grpcAddr string) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	go func() {
		log.Println("Starting Kafka consumer for quiz-completed events")
		s.kafkaConsumer.Start(ctx)
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	log.Println("Serving gRPC on", lis.Addr())
	if err := s.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	s.grpcServer.GracefulStop()

	if err := s.kafkaConsumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	log.Println("Server stopped")
}
