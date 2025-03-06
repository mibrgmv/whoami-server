package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()
	router.GET("/quiz", getQuizzes)
	router.GET("/quiz/:id", getQuizByID)

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.ListenAndServe()
}
