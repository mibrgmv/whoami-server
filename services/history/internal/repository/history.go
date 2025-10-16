package repository

import (
	"context"

	"github.com/mibrgmv/whoami-server/history/internal/models"
)

type HistoryRepository interface {
	Add(ctx context.Context, historyItems []*models.QuizCompletionHistoryItem) ([]*models.QuizCompletionHistoryItem, error)
	Query(ctx context.Context, query Query) ([]*models.QuizCompletionHistoryItem, error)
}
