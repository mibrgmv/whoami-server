package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"whoami-server/internal/models"
	"whoami-server/internal/persistence/repositories"
)

type QuizService struct {
	repo *repositories.QuizRepository
}

func NewQuizService(repo *repositories.QuizRepository) *QuizService {
	return &QuizService{repo: repo}
}

// AddQuizzes godoc
// @Summary Add multiple quizzes
// @Description Add multiple quizzes to the database
// @Tags quizzes
// @Accept json
// @Produce json
// @Param quizzes body []models.Quiz true "Array of quiz objects to add"
// @Success 201 {array} models.Quiz
// @Failure 400
// @Failure 500
// @Router /quiz/a [post]
func (s *QuizService) AddQuizzes(c *gin.Context) {
	var quizzes []models.Quiz
	if err := c.ShouldBindJSON(&quizzes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuizzes, err := s.repo.AddQuizzes(c.Request.Context(), quizzes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add quizzes"})
		return
	}

	c.JSON(http.StatusCreated, createdQuizzes)
}
