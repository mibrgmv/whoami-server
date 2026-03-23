package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"gordle/gateway/internal/models"
	"gordle/gateway/internal/repository"
)

const (
	sessionKeyPrefix = "gordle:session:"
	stateKeyPrefix   = "gordle:auth_state:"
	stateTTL         = 5 * time.Minute
)

type sessionRepo struct {
	client *goredis.Client
}

func NewSessionRepository(client *goredis.Client) repository.SessionRepository {
	return &sessionRepo{client: client}
}

func (r *sessionRepo) Create(ctx context.Context, data *models.SessionData, ttl time.Duration) (string, error) {
	sessionID := uuid.New().String()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if err := r.client.Set(ctx, sessionKeyPrefix+sessionID, jsonData, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	return sessionID, nil
}

func (r *sessionRepo) Get(ctx context.Context, sessionID string) (*models.SessionData, error) {
	data, err := r.client.Get(ctx, sessionKeyPrefix+sessionID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepo) Update(ctx context.Context, sessionID string, data *models.SessionData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ttl, err := r.client.TTL(ctx, sessionKeyPrefix+sessionID).Result()
	if err != nil || ttl <= 0 {
		ttl = 24 * time.Hour
	}

	if err := r.client.Set(ctx, sessionKeyPrefix+sessionID, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (r *sessionRepo) Delete(ctx context.Context, sessionID string) error {
	if err := r.client.Del(ctx, sessionKeyPrefix+sessionID).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *sessionRepo) SetState(ctx context.Context, state, returnTo string) error {
	if err := r.client.Set(ctx, stateKeyPrefix+state, returnTo, stateTTL).Err(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

func (r *sessionRepo) GetState(ctx context.Context, state string) (string, error) {
	returnTo, err := r.client.GetDel(ctx, stateKeyPrefix+state).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", fmt.Errorf("invalid or expired state")
		}
		return "", fmt.Errorf("failed to get state: %w", err)
	}
	return returnTo, nil
}
