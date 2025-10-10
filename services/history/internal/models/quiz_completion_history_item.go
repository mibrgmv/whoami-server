package models

import (
	"fmt"

	"github.com/google/uuid"
	pb "github.com/mibrgmv/whoami-server/history/internal/protogen/history"
)

type QuizCompletionHistoryItem struct {
	ID         uuid.UUID `json:"id"`
	QuizID     uuid.UUID `json:"quiz_id"`
	UserID     uuid.UUID `json:"user_id"`
	QuizResult string    `json:"quiz_result"`
}

func ToModel(protoItem *pb.QuizCompletionHistoryItem) (*QuizCompletionHistoryItem, error) {
	userID, err := uuid.Parse(protoItem.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UserID '%s': %w", protoItem.UserId, err)
	}

	quizID, err := uuid.Parse(protoItem.QuizId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse QuizID '%s': %w", protoItem.QuizId, err)
	}

	return &QuizCompletionHistoryItem{
		UserID:     userID,
		QuizID:     quizID,
		QuizResult: protoItem.QuizResult,
	}, nil
}

func (item *QuizCompletionHistoryItem) ToProto() *pb.QuizCompletionHistoryItem {
	return &pb.QuizCompletionHistoryItem{
		Id:         item.ID.String(),
		UserId:     item.UserID.String(),
		QuizId:     item.QuizID.String(),
		QuizResult: item.QuizResult,
	}
}
