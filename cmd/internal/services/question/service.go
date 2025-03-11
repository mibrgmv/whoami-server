package question

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"whoami-server/cmd/internal/models"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Add godoc
// @Summary Add multiple questions
// @Description Add multiple questions to the database
// @Tags questions
// @Accept json
// @Produce json
// @Param questions body []models.Question true "Array of question objects to add"
// @Success 201 {array} models.Question
// @Failure 400
// @Failure 500
// @Router /question/a [post]
func (s *Service) Add(c *gin.Context) {
	var questions []models.Question
	if err := c.ShouldBindJSON(&questions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuestions, err := s.repo.Add(c.Request.Context(), questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add questions"})
		return
	}

	c.JSON(http.StatusCreated, createdQuestions)
}

// Query godoc
// @Summary Query questions with parameters
// @Description Query questions with parameters such as an array of quiz IDs.
// @Tags questions
// @Produce json
// @Param quiz_ids query []int64 false "Array of Quiz IDs (comma-separated)"
// @Success 200 {array} models.Question
// @Failure 400
// @Failure 500
// @Router /question/q [get]
func (s *Service) Query(c *gin.Context) {
	quizIDsStr := c.Query("quiz_ids")

	var quizIDs []int64
	if quizIDsStr != "" {
		idStrSlice := strings.Split(quizIDsStr, ",")
		for _, idStr := range idStrSlice {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz IDs"})
				return
			}
			quizIDs = append(quizIDs, id)
		}
	}

	questions, err := s.repo.Query(c.Request.Context(), Query{QuizIds: quizIDs})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	c.JSON(http.StatusOK, questions)
}

// GetQuestionsByQuizID godoc
// @Summary Get questions by quiz ID
// @Description Retrieve questions associated with a specific quiz ID.
// @Tags questions
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Success 200 {array} models.Question
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quiz/{id}/questions [get]
func (s *Service) GetQuestionsByQuizID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	questionIDs := []int64{quizID}
	questions, err := s.repo.Query(c.Request.Context(), Query{QuizIds: questionIDs})
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
