// File: internal/handlers/question/handler.go
package question

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"whoami-server/cmd/internal/models"

	questionService "whoami-server/cmd/internal/services/question"
)

type Handler struct {
	service *questionService.Service
}

func NewHandler(service *questionService.Service) *Handler {
	return &Handler{service: service}
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
func (h *Handler) Add(c *gin.Context) {
	var questions []models.Question
	if err := c.ShouldBindJSON(&questions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuestions, err := h.service.Add(c.Request.Context(), questions)
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
// @Success 200 {array} QuestionView
// @Failure 400
// @Failure 500
// @Router /question/q [get]
func (h *Handler) Query(c *gin.Context) {
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

	questions, err := h.service.Query(c.Request.Context(), questionService.Query{QuizIds: quizIDs})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	var result []question
	for _, q := range questions {
		result = append(result, ToView(&q))
	}

	c.JSON(http.StatusOK, result)
}

// GetByQuizID godoc
// @Summary Get questions by quiz ID
// @Description Retrieve questions associated with a specific quiz ID.
// @Tags questions
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Success 200 {array} QuestionView
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quiz/{id}/questions [get]
func (h *Handler) GetByQuizID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	questions, err := h.service.GetByQuizID(c.Request.Context(), quizID)
	if err != nil {
		if errors.Is(err, questionService.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "No questions found for the given quiz ID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	var result []question
	for _, q := range *questions {
		result = append(result, ToView(&q))
	}

	c.JSON(http.StatusOK, result)
}
