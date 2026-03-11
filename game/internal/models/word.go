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
