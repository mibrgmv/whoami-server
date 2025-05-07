package models

import (
	"github.com/google/uuid"
	pb "whoami-server/protogen/golang/quiz"
)

type Quiz struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Results []string  `json:"results"`
}

func (q *Quiz) ToProto() *pb.Quiz {
	return &pb.Quiz{
		Id:      q.ID.String(),
		Title:   q.Title,
		Results: q.Results,
	}
}
