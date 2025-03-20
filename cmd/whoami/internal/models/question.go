package models

import pb "whoami-server/protogen/golang/question"

type Question struct {
	ID             int64                `json:"id"`
	QuizID         int64                `json:"quiz_id"`
	Body           string               `json:"body"`
	OptionsWeights map[string][]float32 `json:"options_weights"`
}

func ToModel(protoQuestion *pb.Question) *Question {
	optionsWeights := make(map[string][]float32)

	for option, protoWeights := range protoQuestion.OptionsWeights {
		weights := make([]float32, len(protoWeights.Weights))
		for i, weight := range protoWeights.Weights {
			weights[i] = float32(weight)
		}

		optionsWeights[option] = weights
	}

	return &Question{
		ID:             protoQuestion.ID,
		QuizID:         protoQuestion.QuizID,
		Body:           protoQuestion.Body,
		OptionsWeights: optionsWeights,
	}
}

func ToProto(model Question) *pb.Question {
	protoOptionsWeights := make(map[string]*pb.OptionWeights)

	for option, weights := range model.OptionsWeights {
		protoWeights := &pb.OptionWeights{
			Weights: make([]float32, len(weights)),
		}

		for i, weight := range weights {
			protoWeights.Weights[i] = float32(weight)
		}

		protoOptionsWeights[option] = protoWeights
	}

	return &pb.Question{
		ID:             model.ID,
		QuizID:         model.QuizID,
		Body:           model.Body,
		OptionsWeights: protoOptionsWeights,
	}
}
