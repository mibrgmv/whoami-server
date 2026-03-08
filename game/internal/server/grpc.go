package server

import (
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/kafka"
	"whoami-server/libs/logging"

	gamegrpc "whoami-server/game/internal/grpc"
	"whoami-server/game/internal/repository/postgres"
	"whoami-server/game/internal/repository/redis"
	"whoami-server/game/internal/service"
	gamev1 "whoami-server/game/pkg/protogen/game/v1"
)

type GrpcServer struct {
	grpcServer *grpc.Server
	producer   kafka.Producer
	logger     *slog.Logger
}

func NewGrpcServer(pool *pgxpool.Pool, redisClient *goredis.Client, kafkaCfg *kafka.Config) *GrpcServer {
	logger := logging.NewLogger("game-service")

	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:  kafkaCfg.Brokers,
		ClientID: "game-service",
	})

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(libsgrpc.DefaultUnaryInterceptors(logger)...),
		grpc.ChainStreamInterceptor(libsgrpc.DefaultStreamInterceptors(logger)...),
	)

	wordRepo := postgres.NewWordRepository(pool)
	wordService := service.NewWordService(wordRepo, redisClient)

	sessionRepo := redis.NewSessionRepository(redisClient)
	gameService := service.NewGameService(sessionRepo, wordService, producer, kafkaCfg.Topics.GameCompleted)
	gameServer := gamegrpc.NewGameServer(gameService, wordService)
	gamev1.RegisterGameServiceServer(s, gameServer)

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
