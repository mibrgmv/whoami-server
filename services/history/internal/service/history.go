package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mibrgmv/whoami-server/history/internal/models"
	"github.com/mibrgmv/whoami-server/history/internal/repository"
	"github.com/mibrgmv/whoami-server/shared/tools"
)

type HistoryService interface {
	CreateItem(ctx context.Context, item *models.QuizCompletionHistoryItem) ([]*models.QuizCompletionHistoryItem, error)
	GetItems(ctx context.Context, userIDs []*uuid.UUID, quizIDs []*uuid.UUID, pageSize int32, pageToken string) ([]*models.QuizCompletionHistoryItem, string, error)
}

type historyService struct {
	repo repository.HistoryRepository
}

func NewHistoryService(repo repository.HistoryRepository) HistoryService {
	return &historyService{
		repo: repo,
	}
}

func (s *historyService) CreateItem(ctx context.Context, item *models.QuizCompletionHistoryItem) ([]*models.QuizCompletionHistoryItem, error) {
	return s.repo.Add(ctx, []*models.QuizCompletionHistoryItem{item})
}

func (s *historyService) GetItems(ctx context.Context, userIDs []*uuid.UUID, quizIDs []*uuid.UUID, pageSize int32, pageToken string) ([]*models.QuizCompletionHistoryItem, string, error) {
	query := repository.Query{
		UserIDs:   userIDs,
		QuizIDs:   quizIDs,
		PageSize:  pageSize,
		PageToken: pageToken,
	}

	items, err := s.repo.Query(ctx, query)
	if err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if pageSize > 0 && len(items) > int(pageSize) {
		items = items[:len(items)-1]
		lastItemID := items[len(items)-1].ID
		nextPageToken = tools.CreatePageToken(lastItemID)
	}

	return items, nextPageToken, nil
}
