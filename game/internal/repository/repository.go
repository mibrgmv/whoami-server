package repository

import (
	"context"

	"github.com/google/uuid"
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

// SessionRepository handles game session operations (Redis)
type SessionRepository interface {
	// Create creates a new game session
	Create(ctx context.Context, session *models.GameSession) (*models.GameSession, error)
	// GetByID returns a game session by its ID
	GetByID(ctx context.Context, sessionID uuid.UUID) (*models.GameSession, error)
	// GetDailyByUser returns the user's daily game session for a specific date (only for registered users)
	GetDailyByUser(ctx context.Context, userID uuid.UUID, language, date string) (*models.GameSession, error)
	// Update updates a game session (including guesses)
	Update(ctx context.Context, session *models.GameSession) error
}
