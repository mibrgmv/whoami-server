package models

import (
	"github.com/google/uuid"
	pb "whoami-server/protogen/golang/quiz"
)

type Quiz struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Results  []string  `json:"results"`
	ImageURL string    `json:"image_url"`
}

func QuizToModel(protoQuiz *pb.Quiz) (*Quiz, error) {
	var quizID uuid.UUID
	var err error

	if protoQuiz.Id == "" {
		quizID = uuid.Nil
	} else {
		quizID, err = uuid.Parse(protoQuiz.Id)
		if err != nil {
			return nil, err
		}
	}

	return &Quiz{
		ID:       quizID,
		Title:    protoQuiz.Title,
		Results:  protoQuiz.Results,
		ImageURL: protoQuiz.ImageUrl,
	}, nil
}

func (q *Quiz) ToProto() *pb.Quiz {
	return &pb.Quiz{
		Id:       q.ID.String(),
		Title:    q.Title,
		Results:  q.Results,
		ImageUrl: q.ImageURL,
	}
}
