package models

import (
	"fmt"
	"github.com/google/uuid"
	pb "whoami-server/protogen/golang/question"
)

type Question struct {
	ID             uuid.UUID            `json:"id"`
	QuizID         uuid.UUID            `json:"quiz_id"`
	Body           string               `json:"body"`
	OptionsWeights map[string][]float32 `json:"options_weights"`
}

func QuestionToModel(protoQuestion *pb.CreateQuestionRequest) (*Question, error) {
	optionsWeights := make(map[string][]float32)

	for option, protoWeights := range protoQuestion.OptionsWeights {
		weights := make([]float32, len(protoWeights.Weights))
		copy(weights, protoWeights.Weights)
		optionsWeights[option] = weights
	}

	var quizID uuid.UUID
	if protoQuestion.QuizId == "" {
		quizID = uuid.Nil
	} else {
		parsedID, err := uuid.Parse(protoQuestion.QuizId)
		if err != nil {
			return nil, fmt.Errorf("failed to parse uuid: %w", err)
		}
		quizID = parsedID
	}

	return &Question{
		QuizID:         quizID,
		Body:           protoQuestion.Body,
		OptionsWeights: optionsWeights,
	}, nil
}

func (q *Question) ToProto() *pb.Question {
	protoOptionsWeights := make(map[string]*pb.OptionWeights)

	for option, weights := range q.OptionsWeights {
		protoWeights := &pb.OptionWeights{
			Weights: make([]float32, len(weights)),
		}
		copy(protoWeights.Weights, weights)
		protoOptionsWeights[option] = protoWeights
	}

	return &pb.Question{
		Id:             q.ID.String(),
		QuizId:         q.QuizID.String(),
		Body:           q.Body,
		OptionsWeights: protoOptionsWeights,
	}
}

func (q *Question) ToProtoWithoutWeights() *pb.QuestionResponse {
	var options []string
	for option := range q.OptionsWeights {
		options = append(options, option)
	}

	return &pb.QuestionResponse{
		Id:      q.ID.String(),
		QuizId:  q.QuizID.String(),
		Body:    q.Body,
		Options: options,
	}
}
