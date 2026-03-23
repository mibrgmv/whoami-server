package consumer

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"gordle/statistics/internal/domain/models"
	"gordle/statistics/internal/domain/service"
	statisticsv1 "gordle/statistics/pkg/protogen/statistics/v1"
)

type GameCompletedHandler struct {
	historyService service.HistoryService
}

func NewGameCompletedHandler(historyService service.HistoryService) *GameCompletedHandler {
	return &GameCompletedHandler{
		historyService: historyService,
	}
}

func (h *GameCompletedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event statisticsv1.GameCompletedEvent
	if err := proto.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal game completed event: %w", err)
	}

	userID, err := uuid.Parse(event.UserId)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	sessionID, err := uuid.Parse(event.SessionId)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	result := &models.GameResult{
		UserID:       userID,
		SessionID:    sessionID,
		GameMode:     event.GameMode,
		GameDate:     event.GameDate,
		TargetWord:   event.TargetWord,
		Guesses:      event.Guesses,
		Result:       event.Result,
		AttemptsUsed: int(event.AttemptsUsed),
	}

	if err := h.historyService.CreateGameHistory(ctx, result); err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	log.Printf("Successfully processed game completed event for user %s, session %s", event.UserId, event.SessionId)
	return nil
}
