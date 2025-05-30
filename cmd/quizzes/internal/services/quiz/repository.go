package quiz

import (
	"context"
	"whoami-server/cmd/quizzes/internal/models"
)

type Repository interface {
	Add(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error)
	Query(ctx context.Context, query Query) ([]*models.Quiz, error)
}
