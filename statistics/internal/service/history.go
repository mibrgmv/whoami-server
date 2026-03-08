package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"whoami-server/libs/tools"
	"whoami-server/statistics/internal/models"
	"whoami-server/statistics/internal/repository"
)

type HistoryService interface {
	CreateFromEvent(ctx context.Context, event GameCompletedEvent) error
	GetUserHistory(ctx context.Context, userID uuid.UUID, pageSize int32, pageToken string) ([]*models.GameHistory, string, error)
}

type GameCompletedEvent struct {
	UserID       string
	SessionID    string
	GameMode     string
	GameDate     string
	TargetWord   string
	Guesses      []string
	Result       string
	AttemptsUsed int
}

type historyService struct {
	historyRepo       repository.HistoryRepository
	statisticsService StatisticsService
}

func NewHistoryService(historyRepo repository.HistoryRepository, statisticsService StatisticsService) HistoryService {
	return &historyService{
		historyRepo:       historyRepo,
		statisticsService: statisticsService,
	}
}

func (s *historyService) CreateFromEvent(ctx context.Context, event GameCompletedEvent) error {
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	sessionID, err := uuid.Parse(event.SessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	var gameDate *string
	if event.GameDate != "" {
		gameDate = &event.GameDate
	}

	history := &models.GameHistory{
		UserID:       userID,
		SessionID:    sessionID,
		GameMode:     event.GameMode,
		GameDate:     gameDate,
		TargetWord:   event.TargetWord,
		Guesses:      event.Guesses,
		Result:       event.Result,
		AttemptsUsed: event.AttemptsUsed,
		CreatedAt:    time.Now(),
	}

	_, err = s.historyRepo.Create(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	if err := s.statisticsService.UpdateFromGame(ctx, userID, event.Result, event.AttemptsUsed, event.GameDate); err != nil {
		return fmt.Errorf("failed to update statistics: %w", err)
	}

	return nil
}

func (s *historyService) GetUserHistory(ctx context.Context, userID uuid.UUID, pageSize int32, pageToken string) ([]*models.GameHistory, string, error) {
	parsedToken, err := tools.ParsePageToken(pageToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	query := models.GameHistoryQuery{
		UserID:    userID,
		PageSize:  pageSize,
		PageToken: parsedToken,
	}

	history, err := s.historyRepo.Query(ctx, query)
	if err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if pageSize > 0 && len(history) > int(pageSize) {
		history = history[:len(history)-1]
		lastID := history[len(history)-1].ID
		nextPageToken = tools.CreatePageToken(lastID)
	}

	return history, nextPageToken, nil
}
