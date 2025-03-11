package quiz

import (
	"github.com/gin-gonic/gin"
	"whoami-server/cmd/internal/services/quiz"
)

type Handler struct {
	s *quiz.Service
}

func NewHandler(s *quiz.Service) *Handler {
	return &Handler{s: s}
}

func (h *Handler) Setup(r *gin.Engine) {
	g := r.Group("/quiz")
	g.GET("/q", h.s.Query)
	g.POST("/a", h.s.Add)
	g.GET("/:id", h.s.GetQuizByID)
}
