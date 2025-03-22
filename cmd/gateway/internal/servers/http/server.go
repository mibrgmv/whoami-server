package http

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"whoami-server/cmd/gateway/internal"
	"whoami-server/cmd/gateway/internal/handlers/question"
	"whoami-server/cmd/gateway/internal/handlers/quiz"
	"whoami-server/cmd/gateway/internal/handlers/user"
	sUser "whoami-server/cmd/gateway/internal/services/user"
	pguser "whoami-server/cmd/gateway/internal/services/user/postgresql"
	"whoami-server/docs"
)

type Server struct {
	server *http.Server
	router *gin.Engine
	config *Config
	pool   *pgxpool.Pool
}

func NewServer(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) setupRouter() error {
	//switch s.config.Server.Mode {
	//case "release":
	//	gin.SetMode(gin.ReleaseMode)
	//case "test":
	//	gin.SetMode(gin.TestMode)
	//default:
	//	gin.SetMode(gin.DebugMode)
	//}

	gin.ForceConsoleColor()
	s.router = gin.Default()
	s.router.Use(gin.Recovery())

	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.Server.Cors.AllowedOrigins,
		AllowMethods:     s.config.Server.Cors.AllowedMethods,
		AllowHeaders:     s.config.Server.Cors.AllowedHeaders,
		AllowCredentials: s.config.Server.Cors.AllowCredentials,
		MaxAge:           time.Duration(s.config.Server.Cors.MaxAge) * time.Second,
	}))

	quizHandler, err := quiz.NewHandler("localhost:50051")
	if err != nil {
		return fmt.Errorf("failed to create quiz handler: %w", err)
	}

	questionHandler, err := question.NewHandler("localhost:50051")
	if err != nil {
		return fmt.Errorf("failed to create question handler: %w", err)
	}

	userRepo := pguser.NewRepository(s.pool)
	userService := sUser.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	routerSetup := internal.RouterSetup{
		QuizHandler:     quizHandler,
		QuestionHandler: questionHandler,
		UserHandler:     userHandler,
	}

	internal.SetupRoutes(s.router, routerSetup)

	if s.config.Swagger.Enabled {
		docs.SwaggerInfo.BasePath = s.config.Swagger.BasePath
		docs.SwaggerInfo.Title = s.config.Swagger.Title
		docs.SwaggerInfo.Description = s.config.Swagger.Description
		docs.SwaggerInfo.Version = s.config.Swagger.Version
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		log.Println("Swagger enabled at /swagger/index.html")
	}

	return nil
}

func (s *Server) connectDB() error {
	// TODO: Replace with actual connection string from config
	poolConfig, err := pgxpool.ParseConfig("postgresql://postgres:postgres@localhost:5432/postgres?sslmode=prefer")
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	s.pool = pool
	log.Println("Connected to database successfully")
	return nil
}

func (s *Server) Start() error {
	if err := s.connectDB(); err != nil {
		return err
	}
	defer s.pool.Close()

	if err := s.setupRouter(); err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.config.GetServerAddress(),
		Handler: s.router,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on %s", s.config.GetServerAddress())
		serverErrors <- s.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case <-shutdown:
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			s.server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		log.Println("Server stopped gracefully")
		return nil
	}
}
