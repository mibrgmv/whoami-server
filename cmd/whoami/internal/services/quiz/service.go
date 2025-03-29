package quiz

import (
	"context"
	"errors"
	"whoami-server/cmd/whoami/internal/models"
)

var ErrQuizNotFound = errors.New("quiz not found")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, quizzes []models.Quiz) ([]models.Quiz, error) {
	return s.repo.Add(ctx, quizzes)
}

func (s *Service) GetAll() ([]models.Quiz, error) {
	return s.repo.Query(context.Background(), Query{})
}

func (s *Service) GetByID(ctx context.Context, quizID int64) (*models.Quiz, error) {
	quizzes, err := s.repo.Query(ctx, Query{Ids: []int64{quizID}})
	if err != nil {
		return nil, err
	}

	if len(quizzes) == 0 {
		return nil, ErrQuizNotFound
	}

	return &quizzes[0], nil
}
