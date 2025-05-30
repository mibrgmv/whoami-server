package question

import (
	"context"
	"whoami-server/cmd/quizzes/internal/models"
)

type Repository interface {
	Add(ctx context.Context, questions []*models.Question) ([]*models.Question, error)
	Query(ctx context.Context, query Query) ([]*models.Question, error)
}
