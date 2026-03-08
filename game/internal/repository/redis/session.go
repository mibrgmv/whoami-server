package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"whoami-server/game/internal/models"
	"whoami-server/game/internal/repository"
)

const (
	sessionKeyPrefix   = "wordle:session:"
	userDailyKeyPrefix = "wordle:daily:"
	sessionTTL         = 24 * time.Hour
)

type sessionRepo struct {
	client *goredis.Client
}

func NewSessionRepository(client *goredis.Client) repository.SessionRepository {
	return &sessionRepo{client: client}
}

func (r *sessionRepo) Create(ctx context.Context, session *models.GameSession) (*models.GameSession, error) {
	session.ID = uuid.New()
	session.StartedAt = time.Now()
	session.Status = models.GameStatusInProgress
	session.AttemptsUsed = 0
	session.Guesses = []models.Guess{}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sessionKeyPrefix + session.ID.String()
	if err := r.client.Set(ctx, key, data, sessionTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	if session.GameMode == models.GameModeDaily && session.UserID != nil && session.GameDate != nil {
		dailyKey := r.dailyKey(*session.UserID, session.Language, *session.GameDate)
		if err := r.client.Set(ctx, dailyKey, session.ID.String(), sessionTTL).Err(); err != nil {
			return nil, fmt.Errorf("failed to save daily reference: %w", err)
		}
	}

	return session, nil
}

func (r *sessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (*models.GameSession, error) {
	key := sessionKeyPrefix + sessionID.String()
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.GameSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepo) GetDailyByUser(ctx context.Context, userID uuid.UUID, language, date string) (*models.GameSession, error) {
	dailyKey := r.dailyKey(userID, language, date)
	sessionIDStr, err := r.client.Get(ctx, dailyKey).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get daily session ID: %w", err)
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID in daily key: %w", err)
	}

	return r.GetByID(ctx, sessionID)
}

func (r *sessionRepo) Update(ctx context.Context, session *models.GameSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sessionKeyPrefix + session.ID.String()

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		ttl = sessionTTL
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (r *sessionRepo) dailyKey(userID uuid.UUID, language, date string) string {
	return fmt.Sprintf("%s%s:%s:%s", userDailyKeyPrefix, userID.String(), language, date)
}
