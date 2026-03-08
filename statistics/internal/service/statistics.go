package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"whoami-server/statistics/internal/models"
	"whoami-server/statistics/internal/repository"
)

type StatisticsService interface {
	GetUserStatistics(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error)
	UpdateFromGame(ctx context.Context, userID uuid.UUID, result string, attemptsUsed int, gameDate string) error
	GetDailyLeaderboard(ctx context.Context, language string, limit int) (string, []*models.LeaderboardEntry, error)
}

type statisticsService struct {
	statsRepo repository.StatisticsRepository
}

func NewStatisticsService(statsRepo repository.StatisticsRepository) StatisticsService {
	return &statisticsService{
		statsRepo: statsRepo,
	}
}

func (s *statisticsService) GetUserStatistics(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error) {
	return s.statsRepo.GetByUserID(ctx, userID)
}

func (s *statisticsService) UpdateFromGame(ctx context.Context, userID uuid.UUID, result string, attemptsUsed int, gameDate string) error {
	stats, err := s.statsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current statistics: %w", err)
	}

	stats.GamesPlayed++

	today := time.Now().Format("2006-01-02")
	stats.LastPlayedDate = &today

	if result == "won" {
		stats.GamesWon++
		stats.LastWonDate = &today

		switch attemptsUsed {
		case 1:
			stats.GuessDistribution.One++
		case 2:
			stats.GuessDistribution.Two++
		case 3:
			stats.GuessDistribution.Three++
		case 4:
			stats.GuessDistribution.Four++
		case 5:
			stats.GuessDistribution.Five++
		case 6:
			stats.GuessDistribution.Six++
		}

		stats.CurrentStreak++
		if stats.CurrentStreak > stats.MaxStreak {
			stats.MaxStreak = stats.CurrentStreak
		}
	} else {
		stats.CurrentStreak = 0
	}

	if err := s.statsRepo.Upsert(ctx, stats); err != nil {
		return fmt.Errorf("failed to save statistics: %w", err)
	}

	return nil
}

func (s *statisticsService) GetDailyLeaderboard(ctx context.Context, language string, limit int) (string, []*models.LeaderboardEntry, error) {
	if language == "" {
		language = "en"
	}
	if limit <= 0 {
		limit = 10
	}

	today := time.Now().Format("2006-01-02")
	entries, err := s.statsRepo.GetDailyLeaderboard(ctx, today, language, limit)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return today, entries, nil
}
