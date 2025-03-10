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

// QueryQuizzes godoc
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

// GetQuizByID godoc
// @Summary Get quiz by ID
// @Description Retrieve a quiz by its ID.
// @Tags quizzes
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Success 200 {object} models.Quiz
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quiz/{id} [get]
func (s *QuizService) GetQuizByID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	quizIDs := []int64{quizID}
	quizzes, err := s.repo.QueryQuizzes(c.Request.Context(), repositories.QuizQuery{Ids: quizIDs})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quiz"})
		return
	}

	if len(quizzes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
		return
	}

	c.JSON(http.StatusOK, quizzes[0])
}
