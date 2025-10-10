package history

import (
	"context"

	"github.com/google/uuid"
	"github.com/mibrgmv/whoami-server/services/history/internal/models"
	"github.com/mibrgmv/whoami-server/shared/tools"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateItem(ctx context.Context, item *models.QuizCompletionHistoryItem) ([]*models.QuizCompletionHistoryItem, error) {
	return s.repo.Add(ctx, []*models.QuizCompletionHistoryItem{item})
}

func (s *Service) GetItems(ctx context.Context, userIDs []*uuid.UUID, quizIDs []*uuid.UUID, pageSize int32, pageToken string) ([]*models.QuizCompletionHistoryItem, string, error) {
	query := Query{
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
