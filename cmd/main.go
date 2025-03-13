package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"time"
	"whoami-server/cmd/internal"
	hQuestion "whoami-server/cmd/internal/handlers/question"
	hQuiz "whoami-server/cmd/internal/handlers/quiz"
	sQuestion "whoami-server/cmd/internal/services/question"
	pgquestion "whoami-server/cmd/internal/services/question/postgresql"
	sQuiz "whoami-server/cmd/internal/services/quiz"
	pgquiz "whoami-server/cmd/internal/services/quiz/postgresql"
	"whoami-server/docs"
)

func main() {
	config, err := (&Config{}).Load("configs/default.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if config.Server.Mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	connString := config.GetPostgresConnectionString()
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse connStr: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	quizzes := pgquiz.NewRepository(pool)
	questions := pgquestion.NewRepository(pool)

	quizService := sQuiz.NewService(quizzes)
	questionService := sQuestion.NewService(questions)

	questionHandler := hQuestion.NewHandler(questionService)
	quizHandler := hQuiz.NewHandler(quizService)

	gin.ForceConsoleColor()
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.Server.Cors.AllowedOrigins,
		AllowMethods:     config.Server.Cors.AllowedMethods,
		AllowHeaders:     config.Server.Cors.AllowedHeaders,
		AllowCredentials: config.Server.Cors.AllowCredentials,
		MaxAge:           time.Duration(config.Server.Cors.MaxAge) * time.Second,
	}))

	if config.Swagger.Enabled {
		docs.SwaggerInfo.BasePath = config.Swagger.BasePath
		docs.SwaggerInfo.Title = config.Swagger.Title
		docs.SwaggerInfo.Description = config.Swagger.Description
		docs.SwaggerInfo.Version = config.Swagger.Version
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		log.Println("Swagger enabled at /swagger/index.html")
	}

	internal.SetupRoutes(router, internal.RouterSetup{
		QuizHandler:     quizHandler,
		QuestionHandler: questionHandler,
	})

	serverAddr := config.GetServerAddress()
	s := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	log.Printf("Server starting on %s", serverAddr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
