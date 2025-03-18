package quiz

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
	"whoami-server/cmd/whoami/internal/services/quiz"
)

type Handler struct {
	quizService     *quiz.Service
	questionService *question.Service
}

func NewHandler(quizService *quiz.Service, questionService *question.Service) *Handler {
	return &Handler{quizService: quizService, questionService: questionService}
}

// Add godoc
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
func (h *Handler) Add(c *gin.Context) {
	var quizzes []models.Quiz
	if err := c.ShouldBindJSON(&quizzes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuizzes, err := h.quizService.Add(c.Request.Context(), quizzes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add quizzes"})
		return
	}

	c.JSON(http.StatusCreated, createdQuizzes)
}

// Query godoc
// @Summary Get quizzes
// @Description Retrieve quizzes based on optional quiz IDs. If no quiz IDs are provided, retrieves all quizzes.
// @Tags quizzes
// @Produce json
// @Param quiz_ids query []int64 false "List of Quiz IDs (comma-separated)"
// @Success 200 {array} models.Quiz
// @Failure 400
// @Failure 500
// @Router /quiz/q [get]
func (h *Handler) Query(c *gin.Context) {
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

	quizzes, err := h.quizService.Query(c.Request.Context(), quiz.Query{Ids: ids})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quizzes"})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}

// GetByID godoc
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
func (h *Handler) GetByID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	q, err := h.quizService.GetByID(c.Request.Context(), quizID)
	if err != nil {
		if errors.Is(err, quiz.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quiz"})
		return
	}

	c.JSON(http.StatusOK, q)
}

// EvaluateAnswers godoc
// @Summary Evaluate answers for a quiz
// @Description Evaluate submitted answers against a quiz and return result weights
// @Tags questions
// @Accept json
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Param answers body []models.Answer true "Array of answer objects to evaluate"
// @Success 200 {object} map[string]interface{}
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quiz/{id}/evaluate [post]
func (h *Handler) EvaluateAnswers(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	var answers []models.Answer
	if err := c.ShouldBindJSON(&answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid answers format", "error": err.Error()})
		return
	}

	if len(answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No answers provided"})
		return
	}

	quiz, err := h.quizService.GetByID(c.Request.Context(), quizID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
		return
	}

	result, err := h.questionService.EvaluateAnswers(c.Request.Context(), answers, *quiz)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response := gin.H{
		"result": result,
	}

	c.JSON(http.StatusOK, response)
}
