package history

import "github.com/google/uuid"

type Query struct {
	IDs     []uuid.UUID
	UserIDs []uuid.UUID
	QuizIDs []uuid.UUID
}
