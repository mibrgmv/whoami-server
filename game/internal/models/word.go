package models

import (
	"github.com/google/uuid"
)

type Word struct {
	ID         uuid.UUID `json:"word_id"`
	Word       string    `json:"word"`
	Language   string    `json:"language"`
	IsSolution bool      `json:"is_solution"`
}

type DailyWord struct {
	ID       uuid.UUID `json:"daily_word_id"`
	WordID   uuid.UUID `json:"word_id"`
	Language string    `json:"language"`
	GameDate string    `json:"game_date"` // YYYY-MM-DD
	Word     string    `json:"word"`      // Joined from words table
}
