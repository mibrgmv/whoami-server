package repository

import (
	"context"

	"whoami-server/quiz/internal/models"
)

type QuizRepository interface {
	Add(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error)
	Query(ctx context.Context, query models.QuizQuery) ([]*models.Quiz, error)
}
