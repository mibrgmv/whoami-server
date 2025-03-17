package quiz

import (
	"context"
	"errors"
	"whoami-server/cmd/whoami/internal/models"
)

var ErrNotFound = errors.New("quiz not found")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, quizzes []models.Quiz) ([]models.Quiz, error) {
	return s.repo.Add(ctx, quizzes)
}

func (s *Service) Query(ctx context.Context, query Query) ([]models.Quiz, error) {
	return s.repo.Query(ctx, query)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*models.Quiz, error) {
	quizzes, err := s.repo.Query(ctx, Query{Ids: []int64{id}})
	if err != nil {
		return nil, err
	}

	if len(quizzes) == 0 {
		return nil, ErrNotFound
	}

	return &quizzes[0], nil
}
