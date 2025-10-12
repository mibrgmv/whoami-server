package models

import (
	"fmt"

	"github.com/google/uuid"
	questionv1 "github.com/mibrgmv/whoami-server/quiz/internal/protogen/question/v1"
)

type Answer struct {
	QuizID     uuid.UUID `json:"quiz_id"`
	QuestionID uuid.UUID `json:"question_id"`
	Body       string    `json:"body"`
}

func AnswerToModel(protoAnswer *questionv1.Answer) (*Answer, error) {
	quizID, err := uuid.Parse(protoAnswer.QuizId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quiz ID '%s': %w", protoAnswer.QuizId, err)
	}

	questionID, err := uuid.Parse(protoAnswer.QuestionId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse question ID '%s': %w", protoAnswer.QuestionId, err)
	}

	return &Answer{
		QuizID:     quizID,
		QuestionID: questionID,
		Body:       protoAnswer.Body,
	}, nil
}
