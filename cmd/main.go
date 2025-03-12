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

	connStr := config.GetPostgresConnectionString()
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse connStr: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	quizzes := pgquiz.NewRepository(pool)
	questions := pgquestion.NewRepository(pool)

	quizService := sQuiz.NewService(quizzes)
	questionService := sQuestion.NewService(questions)

	questionHandler := hQuestion.NewHandler(questionService)
	quizHandler := hQuiz.NewHandler(quizService)

	gin.ForceConsoleColor()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
	}))

	docs.SwaggerInfo.BasePath = ""

	internal.SetupRoutes(router, internal.RouterSetup{
		QuizHandler:     quizHandler,
		QuestionHandler: questionHandler,
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.ListenAndServe()
}
