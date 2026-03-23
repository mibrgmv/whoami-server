package models

import "github.com/google/uuid"

type GameResult struct {
	UserID       uuid.UUID
	SessionID    uuid.UUID
	GameMode     string
	GameDate     string
	TargetWord   string
	Guesses      []string
	Result       string
	AttemptsUsed int
}
