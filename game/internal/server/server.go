package server

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	gamegrpc "gordle/game/internal/grpc"
	"gordle/game/internal/repository/postgres"
	"gordle/game/internal/repository/redis"
	"gordle/game/internal/service"
	"gordle/game/internal/websocket"
	gamev1 "gordle/game/pkg/protogen/game/v1"
	roomv1 "gordle/game/pkg/protogen/room/v1"
	libsgrpc "gordle/libs/grpc"
	"gordle/libs/kafka"
	"gordle/libs/logging"
)

type Server struct {
	grpcServer  *grpc.Server
	httpServer  *http.Server
	wsHub       *websocket.Hub
	wsPubSub    *websocket.PubSub
	producer    kafka.Producer
	logger      *slog.Logger
	roomService service.RoomService
}

func NewServer(pool *pgxpool.Pool, redisClient *goredis.Client, kafkaCfg *kafka.Config) *Server {
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

	wsHub := websocket.NewHub(logger)
	wsPubSub := websocket.NewPubSub(redisClient, wsHub, logger)
	wsHub.SetPubSub(wsPubSub)

	roomRepo := redis.NewRoomRepository(redisClient)
	roomService := service.NewRoomService(roomRepo, wordService, producer, kafkaCfg.Topics.GameCompleted, wsHub)

	gameServer := gamegrpc.NewGameServer(gameService, wordService)
	gamev1.RegisterGameServiceServer(s, gameServer)

	roomServer := gamegrpc.NewRoomServer(roomService)
	roomv1.RegisterRoomServiceServer(s, roomServer)

	reflection.Register(s)
	return &Server{
		grpcServer:  s,
		wsHub:       wsHub,
		wsPubSub:    wsPubSub,
		producer:    producer,
		logger:      logger,
		roomService: roomService,
	}
}

func (s *Server) Start(addr string) error {
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

func (s *Server) StartWebSocket(addr string) error {
	go s.wsHub.Run()

	mux := http.NewServeMux()
	wsHandler := websocket.NewHandler(s.wsHub, s.roomService, s.logger)
	mux.HandleFunc("/rooms/", wsHandler.ServeHTTP)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.logger.Info("serving WebSocket", slog.String("addr", addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("failed to serve WebSocket", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()

	if s.httpServer != nil {
		if err := s.httpServer.Close(); err != nil {
			s.logger.Error("error closing HTTP server", slog.String("error", err.Error()))
		}
	}

	if s.wsPubSub != nil {
		if err := s.wsPubSub.Close(); err != nil {
			s.logger.Error("error closing WebSocket pubsub", slog.String("error", err.Error()))
		}
	}

	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			s.logger.Error("error closing Kafka producer", slog.String("error", err.Error()))
		}
	}

	s.logger.Info("servers stopped")
}
