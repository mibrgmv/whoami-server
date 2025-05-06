package question

import (
	"context"
	"github.com/google/uuid"
	"whoami-server/cmd/whoami/internal/models"
)

type Repository interface {
	Add(ctx context.Context, quizID uuid.UUID, questions []*models.Question) ([]*models.Question, error)
	Query(ctx context.Context, query Query) ([]*models.Question, error)
}
