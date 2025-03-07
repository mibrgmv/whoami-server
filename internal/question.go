package internal

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type question struct {
	ID      int64    `json:"id"`
	QuizID  int64    `json:"quiz_id"`
	Body    string   `json:"body"`
	Options []string `json:"options"`
}

var questions = []question{
	{ID: 1, QuizID: 2, Body: "Are you black?", Options: []string{"Yes", "No"}},
	{ID: 2, QuizID: 2, Body: "Do you drink gasoline?", Options: []string{"Yes", "No"}},
	{ID: 3, QuizID: 2, Body: "Have you ever betrayed your friends?", Options: []string{"Yes", "No"}},
}

func GetQuestionsByQuizID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	var result []question
	for _, q := range questions {
		if q.QuizID == quizID {
			result = append(result, q)
		}
	}

	if len(result) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "No questions found for the given quiz ID"})
		return
	}

	c.IndentedJSON(http.StatusOK, result)
}
