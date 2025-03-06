package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type quiz struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var quizzes = []quiz{
	{ID: "1", Title: "Which sigma boy are you?"},
	{ID: "2", Title: "Who are you from GTA V."},
	{ID: "3", Title: "Are you gay?"},
}

func getQuizzes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, quizzes)
}

func getQuizByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range quizzes {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
}
