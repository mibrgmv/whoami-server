package models

import (
	"github.com/google/uuid"
	pb "whoami-server/protogen/golang/question"
)

type Question struct {
	ID             uuid.UUID            `json:"id"`
	Body           string               `json:"body"`
	OptionsWeights map[string][]float32 `json:"options_weights"`
}

func QuestionToModel(protoQuestion *pb.CreateQuestionRequest) (*Question, error) {
	optionsWeights := make(map[string][]float32)

	for option, protoWeights := range protoQuestion.OptionsWeights {
		weights := make([]float32, len(protoWeights.Weights))
		for i, weight := range protoWeights.Weights {
			weights[i] = float32(weight)
		}

		optionsWeights[option] = weights
	}

	return &Question{
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
		Body:           q.Body,
		OptionsWeights: protoOptionsWeights,
	}
}

func (q *Question) ToProtoWithoutWeights() *pb.QuestionResponse {
	var options []string
	for option, _ := range q.OptionsWeights {
		options = append(options, option)
	}

	return &pb.QuestionResponse{
		Id:      q.ID.String(),
		Body:    q.Body,
		Options: options,
	}
}
