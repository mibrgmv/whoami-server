package quiz

import (
	"context"
	"errors"
	"github.com/google/uuid"
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

func (s *Service) Get(ctx context.Context, pageSize int32, pageToken string) ([]models.Quiz, string, error) {
	quizzes, err := s.repo.Query(ctx, Query{PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if pageSize > 0 && len(quizzes) > int(pageSize) {
		quizzes = quizzes[:len(quizzes)-1]
		nextPageToken = quizzes[len(quizzes)-1].ID.String()
	}

	return quizzes, nextPageToken, err
}

func (s *Service) GetByID(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error) {
	quizzes, err := s.repo.Query(ctx, Query{Ids: []uuid.UUID{quizID}, PageSize: 1})
	if err != nil {
		return nil, err
	}

	if len(quizzes) == 0 {
		return nil, ErrQuizNotFound
	}

	return &quizzes[0], nil
}
