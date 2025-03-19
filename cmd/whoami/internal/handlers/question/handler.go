package question

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"whoami-server/cmd/whoami/internal/models"
	sQuestion "whoami-server/cmd/whoami/internal/services/question"
)

type Handler struct {
	service *sQuestion.Service
}

func NewHandler(service *sQuestion.Service) *Handler {
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

// All godoc
// @Summary Query questions with parameters
// @Description Retrieve all quizzes.
// @Tags questions
// @Produce json
// @Success 200 {array} question
// @Failure 400
// @Failure 500
// @Router /question/ [get]
func (h *Handler) All(c *gin.Context) {
	questions, err := h.service.Query(c.Request.Context(), sQuestion.Query{})
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
// @Success 200 {array} question
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
		if errors.Is(err, sQuestion.ErrNotFound) {
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
