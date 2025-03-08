package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"whoami-server/internal/models"
)

var quizzes = []models.Quiz{
	{ID: 1, Title: "Which sigma boy are you?"},
	{ID: 2, Title: "Who are you from GTA V?"},
	{ID: 3, Title: "Are you gay?"},
}

func GetQuizzes(c *gin.Context) {
	c.JSON(http.StatusOK, quizzes)
}

func GetQuizByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	for _, q := range quizzes {
		if q.ID == id {
			c.JSON(http.StatusOK, q)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
}
