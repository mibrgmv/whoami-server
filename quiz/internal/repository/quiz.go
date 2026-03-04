package repository

import (
	"context"

	"github.com/google/uuid"
	"whoami-server/quiz/internal/models"
)

type QuizRepository interface {
	Add(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error)
	Query(ctx context.Context, query models.QuizQuery) ([]*models.Quiz, error)
	Delete(ctx context.Context, quizID uuid.UUID) error
}
