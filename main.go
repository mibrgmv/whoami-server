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
	"whoami-server/configuration"
	"whoami-server/docs"
	"whoami-server/internal/persistence/repositories"
	"whoami-server/internal/services"
)

func main() {
	config, err := (&configuration.Config{}).Load("config.yaml")
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

	quizzes := repositories.NewQuizRepository(pool)
	questions := repositories.NewQuestionRepository(pool)

	quizService := services.NewQuizService(quizzes)
	questionService := services.NewQuestionService(questions)

	gin.ForceConsoleColor()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
	}))

	// todo make a database context and make all services createable with this database context.
	// something like
	//quizservice(context) {
	//	repo = context.repo
	//}

	docs.SwaggerInfo.BasePath = ""

	router.GET("/quiz", quizService.GetQuizzes)
	router.POST("/quiz/add", quizService.AddQuiz)
	router.GET("/quiz/:id", quizService.GetQuizByID)
	router.GET("/quiz/:id/questions", questionService.GetQuestionsByQuizID)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.ListenAndServe()
}
