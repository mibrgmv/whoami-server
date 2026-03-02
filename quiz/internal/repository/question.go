package repository

import (
	"context"

	"whoami-server/quiz/internal/models"
)

type QuestionRepository interface {
	Add(ctx context.Context, questions []*models.Question) ([]*models.Question, error)
	Query(ctx context.Context, query models.QuestionQuery) ([]*models.Question, error)
}
