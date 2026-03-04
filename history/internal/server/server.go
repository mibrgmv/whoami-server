package server

import (
	"context"
	"log/slog"
	"net"
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
	"whoami-server/libs/logging"
)

type Server struct {
	grpcServer    *grpc.Server
	kafkaConsumer *kafka.Consumer
	cancelFunc    context.CancelFunc
	logger        *slog.Logger
}

func New(pool *pgxpool.Pool, kafkaCfg *kafka.Config) *Server {
	logger := logging.NewLogger("history-service")

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
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
		logger:        logger,
	}
}

func (s *Server) Start(ctx context.Context, grpcAddr string) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	go func() {
		s.logger.Info("starting Kafka consumer for quiz-completed events")
		s.kafkaConsumer.Start(ctx)
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		s.logger.Error("failed to listen", slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("serving gRPC", slog.String("addr", lis.Addr().String()))
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
		s.logger.Error("error closing Kafka consumer", slog.String("error", err.Error()))
	}

	s.logger.Info("server stopped")
}
