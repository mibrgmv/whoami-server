package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"whoami-server/internal/persistence/repositories"
)

type QuestionService struct {
	repo *repositories.QuestionRepository
}

func NewQuestionService(repo *repositories.QuestionRepository) *QuestionService {
	return &QuestionService{repo: repo}
}

func (s *QuestionService) GetQuestionsByQuizID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	questions, err := s.repo.GetQuestionsByQuizID(c.Request.Context(), quizID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	if len(questions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No questions found for the given quiz ID"})
		return
	}

	c.JSON(http.StatusOK, questions)
}
