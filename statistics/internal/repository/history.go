package repository

import (
	"context"

	"gordle/statistics/internal/models"
)

type HistoryRepository interface {
	// Create creates a new game history entry
	Create(ctx context.Context, history *models.GameHistory) (*models.GameHistory, error)
	// Query returns game history entries for a user
	Query(ctx context.Context, query models.GameHistoryQuery) ([]*models.GameHistory, error)
}
