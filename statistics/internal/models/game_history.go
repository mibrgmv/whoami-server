package models

import (
	"time"

	"github.com/google/uuid"
	statisticsv1 "gordle/statistics/pkg/protogen/statistics/v1"
)

type GameHistory struct {
	ID           uuid.UUID `json:"history_id"`
	UserID       uuid.UUID `json:"user_id"`
	SessionID    uuid.UUID `json:"session_id"`
	GameMode     string    `json:"game_mode"`
	GameDate     *string   `json:"game_date"`
	TargetWord   string    `json:"target_word"`
	Guesses      []string  `json:"guesses"`
	Result       string    `json:"result"`
	AttemptsUsed int       `json:"attempts_used"`
	CreatedAt    time.Time `json:"created_at"`
}

func (g *GameHistory) ToProto() *statisticsv1.GameHistoryItem {
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

type GameHistoryQuery struct {
	UserID    uuid.UUID
	PageSize  int32
	PageToken string
}
