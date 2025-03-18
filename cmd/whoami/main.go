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
	"whoami-server/cmd/whoami/internal"
	hQuestion "whoami-server/cmd/whoami/internal/handlers/question"
	hQuiz "whoami-server/cmd/whoami/internal/handlers/quiz"
	hUser "whoami-server/cmd/whoami/internal/handlers/user"
	sQuestion "whoami-server/cmd/whoami/internal/services/question"
	pgquestion "whoami-server/cmd/whoami/internal/services/question/postgresql"
	sQuiz "whoami-server/cmd/whoami/internal/services/quiz"
	pgquiz "whoami-server/cmd/whoami/internal/services/quiz/postgresql"
	sUser "whoami-server/cmd/whoami/internal/services/user"
	pguser "whoami-server/cmd/whoami/internal/services/user/postgresql"
	"whoami-server/docs"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	config, err := (&Config{}).Load("configs/default.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	//if config.Server.Mode == "release" {
	//	gin.SetMode(gin.ReleaseMode)
	//} else if config.Server.Mode == "test" {
	//	gin.SetMode(gin.TestMode)
	//} else {
	//	gin.SetMode(gin.DebugMode)
	//}

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
	users := pguser.NewRepository(pool)

	quizService := sQuiz.NewService(quizzes)
	questionService, err := sQuestion.NewService(questions, "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create question service: %v", err)
	}
	userService := sUser.NewService(users)

	questionHandler := hQuestion.NewHandler(questionService)
	quizHandler := hQuiz.NewHandler(quizService, questionService)
	userHandler := hUser.NewHandler(userService)

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
		UserHandler:     userHandler,
	})

	// todo need to close the grpc conn

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
