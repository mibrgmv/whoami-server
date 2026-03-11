package models

import (
	"github.com/google/uuid"
	statisticsv1 "gordle/statistics/pkg/protogen/statistics/v1"
)

type GuessDistribution struct {
	One   int `json:"1"`
	Two   int `json:"2"`
	Three int `json:"3"`
	Four  int `json:"4"`
	Five  int `json:"5"`
	Six   int `json:"6"`
}

type UserStatistics struct {
	UserID            uuid.UUID         `json:"user_id"`
	GamesPlayed       int               `json:"games_played"`
	GamesWon          int               `json:"games_won"`
	CurrentStreak     int               `json:"current_streak"`
	MaxStreak         int               `json:"max_streak"`
	GuessDistribution GuessDistribution `json:"guess_distribution"`
	LastPlayedDate    *string           `json:"last_played_date"`
	LastWonDate       *string           `json:"last_won_date"`
}

func (s *UserStatistics) WinPercentage() float64 {
	if s.GamesPlayed == 0 {
		return 0
	}
	return float64(s.GamesWon) / float64(s.GamesPlayed) * 100
}

func (s *UserStatistics) ToProto() *statisticsv1.UserStatistics {
	stats := &statisticsv1.UserStatistics{
		UserId:        s.UserID.String(),
		GamesPlayed:   int32(s.GamesPlayed),
		GamesWon:      int32(s.GamesWon),
		WinPercentage: s.WinPercentage(),
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
		stats.LastPlayedDate = *s.LastPlayedDate
	}

	return stats
}

type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	AttemptsUsed int       `json:"attempts_used"`
	IsWon        bool      `json:"is_won"`
}

func (e *LeaderboardEntry) ToProto() *statisticsv1.LeaderboardEntry {
	return &statisticsv1.LeaderboardEntry{
		Rank:         int32(e.Rank),
		UserId:       e.UserID.String(),
		Username:     e.Username,
		AttemptsUsed: int32(e.AttemptsUsed),
		IsWon:        e.IsWon,
	}
}
