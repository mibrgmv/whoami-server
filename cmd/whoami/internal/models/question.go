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

func QuestionToModel(protoQuestion *pb.Question) (*Question, error) {
	var id, quizID uuid.UUID
	var err error
	optionsWeights := make(map[string][]float32)

	for option, protoWeights := range protoQuestion.OptionsWeights {
		weights := make([]float32, len(protoWeights.Weights))
		for i, weight := range protoWeights.Weights {
			weights[i] = float32(weight)
		}

		optionsWeights[option] = weights
	}

	// todo stinky
	if protoQuestion.Id == "" {
		id = uuid.Nil
	} else {
		id, err = uuid.Parse(protoQuestion.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse question ID '%s': %w", protoQuestion.Id, err)
		}
	}

	if protoQuestion.QuizId == "" {
		quizID = uuid.Nil
	} else {
		quizID, err = uuid.Parse(protoQuestion.QuizId)
		if err != nil {
			return nil, fmt.Errorf("failed to parse question ID '%s': %w", protoQuestion.Id, err)
		}
	}

	return &Question{
		ID:             id,
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

		for i, weight := range weights {
			protoWeights.Weights[i] = float32(weight)
		}

		protoOptionsWeights[option] = protoWeights
	}

	return &pb.Question{
		Id:             q.ID.String(),
		QuizId:         q.QuizID.String(),
		Body:           q.Body,
		OptionsWeights: protoOptionsWeights,
	}
}
