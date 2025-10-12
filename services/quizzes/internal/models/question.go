package models

import (
	"fmt"

	"github.com/google/uuid"
	questionv1 "github.com/mibrgmv/whoami-server/quizzes/internal/protogen/question/v1"
)

type Question struct {
	ID             uuid.UUID            `json:"id"`
	QuizID         uuid.UUID            `json:"quiz_id"`
	Body           string               `json:"body"`
	OptionsWeights map[string][]float32 `json:"options_weights"`
}

func QuestionToModel(protoQuestion *questionv1.CreateQuestionRequest) (*Question, error) {
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

func (q *Question) ToProto() *questionv1.Question {
	protoOptionsWeights := make(map[string]*questionv1.OptionWeights)

	for option, weights := range q.OptionsWeights {
		protoWeights := &questionv1.OptionWeights{
			Weights: make([]float32, len(weights)),
		}
		copy(protoWeights.Weights, weights)
		protoOptionsWeights[option] = protoWeights
	}

	return &questionv1.Question{
		Id:             q.ID.String(),
		QuizId:         q.QuizID.String(),
		Body:           q.Body,
		OptionsWeights: protoOptionsWeights,
	}
}

func (q *Question) ToProtoWithoutWeights() *questionv1.QuestionResponse {
	var options []string
	for option := range q.OptionsWeights {
		options = append(options, option)
	}

	return &questionv1.QuestionResponse{
		Id:      q.ID.String(),
		QuizId:  q.QuizID.String(),
		Body:    q.Body,
		Options: options,
	}
}
