package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"whoami-server/internal/persistence/repositories"
)

type QuizService struct {
	repo *repositories.QuizRepository
}

func NewQuizService(repo *repositories.QuizRepository) *QuizService {
	return &QuizService{repo: repo}
}

func (s *QuizService) GetQuizzes(c *gin.Context) {
	var quizzes, err = s.repo.GetQuizzes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to get quizzes"})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}

func (s *QuizService) GetQuizByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	quiz, err := s.repo.GetQuizByID(c.Request.Context(), id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to get quiz by ID"})
		return
	}

	c.JSON(http.StatusOK, quiz)
}
