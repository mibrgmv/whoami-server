package question

import (
	"github.com/gin-gonic/gin"
	"whoami-server/cmd/internal/services/question"
)

type Handler struct {
	s *question.Service
}

func NewHandler(s *question.Service) *Handler {
	return &Handler{s: s}
}

func (h *Handler) Setup(r *gin.Engine) {
	g := r.Group("/question")
	g.GET("/q", h.s.Query)
	g.POST("/a", h.s.Add)

	r.GET("/quiz/:id/questions", h.s.GetQuestionsByQuizID)
}
