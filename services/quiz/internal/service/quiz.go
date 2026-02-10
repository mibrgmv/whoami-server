package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mibrgmv/whoami-server/quiz/internal/models"
	"github.com/mibrgmv/whoami-server/quiz/internal/repository"
	"github.com/mibrgmv/whoami-server/shared/tools"
)

var ErrQuizNotFound = errors.New("quiz not found")

type QuizService interface {
	Add(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error)
	Get(ctx context.Context, pageSize int32, pageToken string) ([]*models.Quiz, string, error)
	GetByID(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error)
}

type quizService struct {
	quizRepo repository.QuizRepository
}

func NewQuizService(quizRepo repository.QuizRepository) QuizService {
	return &quizService{quizRepo: quizRepo}
}

func (s *quizService) Add(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error) {
	return s.quizRepo.Add(ctx, quiz)
}

func (s *quizService) Get(ctx context.Context, pageSize int32, pageToken string) ([]*models.Quiz, string, error) {
	parsedToken, err := tools.ParsePageToken(pageToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	quizzes, err := s.quizRepo.Query(ctx, models.QuizQuery{PageSize: pageSize, PageToken: parsedToken})
	if err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if pageSize > 0 && len(quizzes) > int(pageSize) {
		quizzes = quizzes[:len(quizzes)-1]
		lastQuizID := quizzes[len(quizzes)-1].ID
		nextPageToken = tools.CreatePageToken(lastQuizID)
	}

	return quizzes, nextPageToken, err
}

func (s *quizService) GetByID(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error) {
	quizzes, err := s.quizRepo.Query(ctx, models.QuizQuery{Ids: []uuid.UUID{quizID}, PageSize: 1})
	if err != nil {
		return nil, err
	}

	if len(quizzes) == 0 {
		return nil, ErrQuizNotFound
	}

	return quizzes[0], nil
}
