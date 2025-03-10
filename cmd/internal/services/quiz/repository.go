package quiz

import (
	"context"
	"whoami-server/cmd/internal/models"
)

type Repository interface {
	Add(ctx context.Context, quizzes []models.Quiz) ([]models.Quiz, error)
	Query(ctx context.Context, query Query) ([]models.Quiz, error)
}
