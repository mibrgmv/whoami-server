package services

import (
	"context"
	"whoami-server/cmd/history/internal/models"
)

type Service struct {
	history Repository
}

func NewService(repo Repository) *Service {
	return &Service{history: repo}
}

func (s *Service) Add(ctx context.Context, items []models.QuizCompletionHistoryItem) ([]models.QuizCompletionHistoryItem, error) {
	return s.history.Add(ctx, items)
}

func (s *Service) Query(ctx context.Context, query Query) ([]models.QuizCompletionHistoryItem, error) {
	return s.history.Query(ctx, query)
}
