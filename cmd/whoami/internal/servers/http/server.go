package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"whoami-server/cmd/whoami/internal"
	questionHandler "whoami-server/cmd/whoami/internal/handlers/question"
	quizHandler "whoami-server/cmd/whoami/internal/handlers/quiz"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func DefaultConfig() ServerConfig {
	return ServerConfig{
		Port:            "8080",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

type Server struct {
	config     ServerConfig
	router     *gin.Engine
	httpServer *http.Server
}

func NewServer(config ServerConfig, quizHandler *quizHandler.Handler, questionHandler *questionHandler.Handler) *Server {
	router := gin.Default()

	router.Use(gin.Recovery())

	routerSetup := internal.RouterSetup{
		QuizHandler:     quizHandler,
		QuestionHandler: questionHandler,
	}
	internal.SetupRoutes(router, routerSetup)

	httpServer := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		config:     config,
		router:     router,
		httpServer: httpServer,
	}
}

func (s *Server) Start() error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on port %s", s.config.Port)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case <-shutdown:
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.httpServer.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		log.Println("Server stopped gracefully")
		return nil
	}
}
