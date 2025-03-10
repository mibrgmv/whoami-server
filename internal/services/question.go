package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"whoami-server/internal/models"
	"whoami-server/internal/persistence/repositories"
)

type QuestionService struct {
	repo *repositories.QuestionRepository
}

func NewQuestionService(repo *repositories.QuestionRepository) *QuestionService {
	return &QuestionService{repo: repo}
}

// AddQuestions godoc
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
func (s *QuestionService) AddQuestions(c *gin.Context) {
	var questions []models.Question
	if err := c.ShouldBindJSON(&questions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuestions, err := s.repo.AddQuestions(c.Request.Context(), questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add questions"})
		return
	}

	c.JSON(http.StatusCreated, createdQuestions)
}

// QueryQuestions godoc
// @Summary Query questions with parameters
// @Description Query questions with parameters such as an array of quiz IDs.
// @Tags questions
// @Produce json
// @Param quiz_ids query []int64 false "Array of Quiz IDs (comma-separated)"
// @Success 200 {array} models.Question
// @Failure 400
// @Failure 500
// @Router /question/q [get]
func (s *QuestionService) QueryQuestions(c *gin.Context) {
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

	questions, err := s.repo.QueryQuestions(c.Request.Context(), repositories.QuestionQuery{QuizIds: quizIDs})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	c.JSON(http.StatusOK, questions)
}
