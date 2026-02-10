package repository

import (
	"context"

	"github.com/mibrgmv/whoami-server/quiz/internal/models"
)

type QuestionRepository interface {
	Add(ctx context.Context, questions []*models.Question) ([]*models.Question, error)
	Query(ctx context.Context, query models.QuestionQuery) ([]*models.Question, error)
}
