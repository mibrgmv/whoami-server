package repository

import (
	"context"

	"github.com/google/uuid"
	"whoami-server/statistics/internal/models"
)

type StatisticsRepository interface {
	// GetByUserID returns statistics for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error)
	// Upsert creates or updates user statistics
	Upsert(ctx context.Context, stats *models.UserStatistics) error
	// GetDailyLeaderboard returns the leaderboard for a specific date
	GetDailyLeaderboard(ctx context.Context, date, language string, limit int) ([]*models.LeaderboardEntry, error)
}
