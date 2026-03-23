package repository

import (
	"context"
	"time"

	"gordle/gateway/internal/models"
)

type SessionRepository interface {
	Create(ctx context.Context, data *models.SessionData, ttl time.Duration) (sessionID string, err error)
	Get(ctx context.Context, sessionID string) (*models.SessionData, error)
	Update(ctx context.Context, sessionID string, data *models.SessionData) error
	Delete(ctx context.Context, sessionID string) error
	SetState(ctx context.Context, state, returnTo string) error
	GetState(ctx context.Context, state string) (returnTo string, err error)
}
