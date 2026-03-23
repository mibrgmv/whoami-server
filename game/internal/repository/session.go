package repository

import (
	"context"

	"github.com/google/uuid"
	"gordle/game/internal/models"
)

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
