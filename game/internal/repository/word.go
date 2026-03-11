package repository

import (
	"context"

	"gordle/game/internal/models"
)

// WordRepository handles word dictionary operations (PostgreSQL)
type WordRepository interface {
	// GetByWord returns a word if it exists in the dictionary
	GetByWord(ctx context.Context, word, language string) (*models.Word, error)
	// GetRandomSolution returns a random word that can be used as a solution
	GetRandomSolution(ctx context.Context, language string) (*models.Word, error)
	// GetSolutionByOffset returns a solution word at a given offset (for deterministic daily word)
	GetSolutionByOffset(ctx context.Context, language string, offset int) (*models.Word, error)
	// WordExists checks if a word exists in the dictionary
	WordExists(ctx context.Context, word, language string) (bool, error)
}
