package models

import "github.com/google/uuid"

type QuizQuery struct {
	Ids       []uuid.UUID
	PageSize  int32
	PageToken string
}
