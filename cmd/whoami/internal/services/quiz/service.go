package quiz

import (
	"context"
	"errors"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/internal/cache"
)

var ErrQuizNotFound = errors.New("quiz not found")

const (
	allQuizzesCacheKey = "quizzes:all"
)

type Service struct {
	repo  Repository
	cache cache.Interface
}

func NewService(repo Repository, cache cache.Interface) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

func (s *Service) Add(ctx context.Context, quizzes []models.Quiz) ([]models.Quiz, error) {
	err := s.cache.Delete(ctx, allQuizzesCacheKey)
	if err != nil {
		return nil, err
	}
	return s.repo.Add(ctx, quizzes)
}

func (s *Service) GetAll(ctx context.Context) ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := s.cache.Get(ctx, allQuizzesCacheKey, &quizzes)
	if err == nil {
		return quizzes, nil
	}

	quizzes, err = s.repo.Query(ctx, Query{})

	if err := s.cache.Set(ctx, allQuizzesCacheKey, &quizzes); err != nil {
		// todo ignored error
	}

	return quizzes, err
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
