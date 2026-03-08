package models

import (
	"time"

	"github.com/google/uuid"
	gamev1 "whoami-server/game/pkg/protogen/game/v1"
)

type GameStatus string

const (
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusWon        GameStatus = "won"
	GameStatusLost       GameStatus = "lost"
)

type GameMode string

const (
	GameModeDaily  GameMode = "daily"
	GameModeRandom GameMode = "random"
)

const MaxAttempts = 6

type GameSession struct {
	ID           uuid.UUID  `json:"session_id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Language     string     `json:"language"`
	GameMode     GameMode   `json:"game_mode"`
	GameDate     *string    `json:"game_date,omitempty"`
	Status       GameStatus `json:"status"`
	AttemptsUsed int        `json:"attempts_used"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	TargetWord   string     `json:"target_word"`
	Guesses      []Guess    `json:"guesses"`
}

func (g *GameSession) ToProto() *gamev1.GameSession {
	session := &gamev1.GameSession{
		SessionId:    g.ID.String(),
		Language:     g.Language,
		GameMode:     gameModeToProto(g.GameMode),
		Status:       g.Status.ToProto(),
		AttemptsUsed: int32(g.AttemptsUsed),
		MaxAttempts:  MaxAttempts,
		StartedAt:    g.StartedAt.Format(time.RFC3339),
	}

	if g.UserID != nil {
		session.UserId = g.UserID.String()
	}

	if g.GameDate != nil {
		session.GameDate = *g.GameDate
	}

	if g.CompletedAt != nil {
		session.CompletedAt = g.CompletedAt.Format(time.RFC3339)
	}

	if g.Status == GameStatusWon || g.Status == GameStatusLost {
		session.TargetWord = g.TargetWord
	}

	for _, guess := range g.Guesses {
		session.Guesses = append(session.Guesses, guess.ToProto())
	}

	return session
}

func (g *GameSession) IsGuest() bool {
	return g.UserID == nil
}

func gameModeToProto(mode GameMode) gamev1.GameMode {
	switch mode {
	case GameModeDaily:
		return gamev1.GameMode_GAME_MODE_DAILY
	case GameModeRandom:
		return gamev1.GameMode_GAME_MODE_RANDOM
	default:
		return gamev1.GameMode_GAME_MODE_UNSPECIFIED
	}
}

func (s GameStatus) ToProto() gamev1.GameStatus {
	switch s {
	case GameStatusInProgress:
		return gamev1.GameStatus_GAME_STATUS_IN_PROGRESS
	case GameStatusWon:
		return gamev1.GameStatus_GAME_STATUS_WON
	case GameStatusLost:
		return gamev1.GameStatus_GAME_STATUS_LOST
	default:
		return gamev1.GameStatus_GAME_STATUS_UNSPECIFIED
	}
}

func GameModeFromProto(mode gamev1.GameMode) GameMode {
	switch mode {
	case gamev1.GameMode_GAME_MODE_DAILY:
		return GameModeDaily
	case gamev1.GameMode_GAME_MODE_RANDOM:
		return GameModeRandom
	default:
		return GameModeDaily
	}
}
