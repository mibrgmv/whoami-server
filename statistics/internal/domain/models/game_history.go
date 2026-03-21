package models

import (
	"time"

	"github.com/google/uuid"
)

type GameHistory struct {
	ID           uuid.UUID  `json:"history_id"`
	UserID       uuid.UUID  `json:"user_id"`
	SessionID    uuid.UUID  `json:"session_id"`
	GameMode     string     `json:"game_mode"`
	GameDate     *time.Time `json:"game_date"`
	TargetWord   string     `json:"target_word"`
	Guesses      []string   `json:"guesses"`
	Result       string     `json:"result"`
	AttemptsUsed int        `json:"attempts_used"`
	CreatedAt    time.Time  `json:"created_at"`
}

type GameHistoryQuery struct {
	UserID    uuid.UUID
	PageSize  int32
	PageToken string
}
