package models

import (
	"github.com/google/uuid"
	quizv1 "github.com/mibrgmv/whoami-server/quizzes/internal/protogen/quiz/v1"
)

type Quiz struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Results []string  `json:"results"`
}

func (q *Quiz) ToProto() *quizv1.Quiz {
	return &quizv1.Quiz{
		Id:      q.ID.String(),
		Title:   q.Title,
		Results: q.Results,
	}
}
