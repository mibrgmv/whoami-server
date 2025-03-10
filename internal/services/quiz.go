package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
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

// GetQuizzes godoc
// @Summary Get quizzes
// @Description Retrieve quizzes based on optional quiz IDs. If no quiz IDs are provided, retrieves all quizzes.
// @Tags quizzes
// @Produce json
// @Param quiz_ids query []int64 false "List of Quiz IDs (comma-separated)"
// @Success 200 {array} models.Quiz
// @Failure 400
// @Failure 500
// @Router /quiz/q [get]
func (s *QuizService) QueryQuizzes(c *gin.Context) {
	idsStr := c.Query("quiz_ids")

	var ids []int64
	if idsStr != "" {
		idStrSlice := strings.Split(idsStr, ",")
		for _, idStr := range idStrSlice {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid IDs"})
				return
			}
			ids = append(ids, id)
		}
	}

	quizzes, err := s.repo.QueryQuizzes(c.Request.Context(), repositories.QuizQuery{Ids: ids})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quizzes"})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}
