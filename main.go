package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"whoami-server/configuration"
	"whoami-server/internal"
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

	router.GET("/quiz", internal.GetQuizzes)
	router.GET("/quiz/:id", internal.GetQuizByID)
	router.GET("/quiz/:id/questions", internal.GetQuestionsByQuizID)

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.ListenAndServe()
}
