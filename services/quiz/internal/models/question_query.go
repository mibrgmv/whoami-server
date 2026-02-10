package models

import "github.com/google/uuid"

type QuestionQuery struct {
	QuizIds []uuid.UUID
}
