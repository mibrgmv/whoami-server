package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gordle/libs/tools"
	"gordle/statistics/internal/domain/models"
	"gordle/statistics/internal/repository"
)

type HistoryService interface {
	CreateGameHistory(ctx context.Context, result *models.GameResult) error
	GetUserHistory(ctx context.Context, userID uuid.UUID, pageSize int32, pageToken string) ([]*models.GameHistory, string, error)
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

func (s *historyService) CreateGameHistory(ctx context.Context, result *models.GameResult) error {
	var gameDatePtr *string
	if result.GameDate != "" {
		gameDatePtr = &result.GameDate
	}

	history := &models.GameHistory{
		UserID:       result.UserID,
		SessionID:    result.SessionID,
		GameMode:     result.GameMode,
		GameDate:     gameDatePtr,
		TargetWord:   result.TargetWord,
		Guesses:      result.Guesses,
		Result:       result.Result,
		AttemptsUsed: result.AttemptsUsed,
		CreatedAt:    time.Now(),
	}

	_, err := s.historyRepo.Create(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	if err := s.statisticsService.UpdateFromGame(ctx, result.UserID, result.Result, result.AttemptsUsed, result.GameDate); err != nil {
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
