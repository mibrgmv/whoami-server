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

// AddQuiz godoc
// @Summary Add a new quiz
// @Description Add a new quiz to the database
// @Tags quizzes
// @Accept json
// @Produce json
// @Param quiz body models.Quiz true "Quiz object to add"
// @Success 201 {object} models.Quiz
// @Failure 400
// @Failure 500
// @Router /quiz/add [post]
func (s *QuizService) AddQuiz(c *gin.Context) {
	var quiz models.Quiz
	if err := c.ShouldBindJSON(&quiz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuiz, err := s.repo.AddQuiz(c.Request.Context(), quiz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add quiz"})
		return
	}

	c.JSON(http.StatusCreated, createdQuiz)
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
