package repository

import (
	"context"

	"whoami-server/game/internal/models"
)

// WordRepository handles word dictionary operations (PostgreSQL)
type WordRepository interface {
	// GetByWord returns a word if it exists in the dictionary
	GetByWord(ctx context.Context, word, language string) (*models.Word, error)
	// GetRandomSolution returns a random word that can be used as a solution
	GetRandomSolution(ctx context.Context, language string) (*models.Word, error)
	// GetDailyWord returns the daily word for a given date and language
	GetDailyWord(ctx context.Context, date, language string) (*models.DailyWord, error)
	// WordExists checks if a word exists in the dictionary
	WordExists(ctx context.Context, word, language string) (bool, error)
}
