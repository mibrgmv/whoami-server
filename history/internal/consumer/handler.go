package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	libskafka "whoami-server/libs/kafka"

	"whoami-server/history/internal/models"
	"whoami-server/history/internal/service"
)

type QuizCompletedHandler struct {
	historyService service.HistoryService
}

func NewQuizCompletedHandler(historyService service.HistoryService) *QuizCompletedHandler {
	return &QuizCompletedHandler{
		historyService: historyService,
	}
}

func (h *QuizCompletedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event libskafka.QuizCompletedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal quiz completed event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	quizID, err := uuid.Parse(event.QuizID)
	if err != nil {
		return fmt.Errorf("failed to parse quiz ID: %w", err)
	}

	item := &models.QuizCompletionHistoryItem{
		UserID:     userID,
		QuizID:     quizID,
		QuizResult: event.QuizResult,
	}

	_, err = h.historyService.CreateItem(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to create history item: %w", err)
	}

	log.Printf("Successfully processed quiz completed event for user %s, quiz %s", event.UserID, event.QuizID)
	return nil
}
