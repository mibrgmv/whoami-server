package question

import (
	"context"
	"errors"
	"whoami-server/cmd/internal/models"
)

var ErrNotFound = errors.New("no questions found for quiz")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, questions []models.Question) ([]models.Question, error) {
	return s.repo.Add(ctx, questions)
}

func (s *Service) Query(ctx context.Context, query Query) ([]models.Question, error) {
	return s.repo.Query(ctx, query)
}

func (s *Service) GetByQuizID(ctx context.Context, quizID int64) (*[]models.Question, error) {
	questions, err := s.repo.Query(ctx, Query{QuizIds: []int64{quizID}})
	if err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return nil, ErrNotFound
	}

	return &questions, nil
}
