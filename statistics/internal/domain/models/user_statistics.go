package models

import (
	"time"

	"github.com/google/uuid"
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
	LastPlayedDate    *time.Time        `json:"last_played_date"`
	LastWonDate       *time.Time        `json:"last_won_date"`
}

func (s *UserStatistics) WinPercentage() float64 {
	if s.GamesPlayed == 0 {
		return 0
	}
	return float64(s.GamesWon) / float64(s.GamesPlayed) * 100
}

type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	AttemptsUsed int       `json:"attempts_used"`
	IsWon        bool      `json:"is_won"`
}
