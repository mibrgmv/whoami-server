package grpc

import (
	"math"
	"time"

	"gordle/statistics/internal/domain/models"
	statisticsv1 "gordle/statistics/pkg/protogen/statistics/v1"
)

func userStatisticsToProto(s *models.UserStatistics) *statisticsv1.UserStatistics {
	stats := &statisticsv1.UserStatistics{
		UserId:        s.UserID.String(),
		GamesPlayed:   int32(s.GamesPlayed),
		GamesWon:      int32(s.GamesWon),
		WinPercentage: roundToTwoDecimals(s.WinPercentage()),
		CurrentStreak: int32(s.CurrentStreak),
		MaxStreak:     int32(s.MaxStreak),
		GuessDistribution: &statisticsv1.GuessDistribution{
			One:   int32(s.GuessDistribution.One),
			Two:   int32(s.GuessDistribution.Two),
			Three: int32(s.GuessDistribution.Three),
			Four:  int32(s.GuessDistribution.Four),
			Five:  int32(s.GuessDistribution.Five),
			Six:   int32(s.GuessDistribution.Six),
		},
	}

	if s.LastPlayedDate != nil {
		stats.LastPlayedDate = s.LastPlayedDate.Format("2006-01-02")
	}

	return stats
}

func gameHistoryToProto(g *models.GameHistory) *statisticsv1.GameHistoryItem {
	item := &statisticsv1.GameHistoryItem{
		HistoryId:    g.ID.String(),
		SessionId:    g.SessionID.String(),
		GameMode:     g.GameMode,
		TargetWord:   g.TargetWord,
		Guesses:      g.Guesses,
		Result:       g.Result,
		AttemptsUsed: int32(g.AttemptsUsed),
		CreatedAt:    g.CreatedAt.Format(time.RFC3339),
	}

	if g.GameDate != nil {
		item.GameDate = *g.GameDate
	}

	return item
}

func leaderboardEntryToProto(e *models.LeaderboardEntry) *statisticsv1.LeaderboardEntry {
	return &statisticsv1.LeaderboardEntry{
		Rank:         int32(e.Rank),
		UserId:       e.UserID.String(),
		Username:     e.Username,
		AttemptsUsed: int32(e.AttemptsUsed),
		IsWon:        e.IsWon,
	}
}

func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}
