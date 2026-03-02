package repository

import "github.com/google/uuid"

type Query struct {
	UserIDs   []*uuid.UUID
	QuizIDs   []*uuid.UUID
	PageSize  int32
	PageToken string
}
