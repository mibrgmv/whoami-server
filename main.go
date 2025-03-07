package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"whoami-server/configuration"
	"whoami-server/internal/services"
)

func main() {
	config, err := (&configuration.Config{}).Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	connStr := config.GetPostgresConnectionString()
	log.Printf("Postgres connection string: %s", connStr)

	gin.ForceConsoleColor()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
	}))

	router.GET("/quiz", services.GetQuizzes)
	router.GET("/quiz/:id", services.GetQuizByID)
	router.GET("/quiz/:id/questions", services.GetQuestionsByQuizID)

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.ListenAndServe()
}
