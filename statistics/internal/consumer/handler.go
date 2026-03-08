package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	libskafka "whoami-server/libs/kafka"
	"whoami-server/statistics/internal/service"
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
	var event libskafka.GameCompletedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal game completed event: %w", err)
	}

	serviceEvent := service.GameCompletedEvent{
		UserID:       event.UserID,
		SessionID:    event.SessionID,
		GameMode:     event.GameMode,
		GameDate:     event.GameDate,
		TargetWord:   event.TargetWord,
		Guesses:      event.Guesses,
		Result:       event.Result,
		AttemptsUsed: event.AttemptsUsed,
	}

	if err := h.historyService.CreateFromEvent(ctx, serviceEvent); err != nil {
		return fmt.Errorf("failed to create history from event: %w", err)
	}

	log.Printf("Successfully processed game completed event for user %s, session %s", event.UserID, event.SessionID)
	return nil
}
