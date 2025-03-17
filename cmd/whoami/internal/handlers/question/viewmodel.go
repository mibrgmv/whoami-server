package question

import (
	"whoami-server/cmd/whoami/internal/models"
)

type question struct {
	ID      int64    `json:"id"`
	QuizID  int64    `json:"quiz_id"`
	Body    string   `json:"body"`
	Options []string `json:"options"`
}

func ToView(m *models.Question) question {
	options := make([]string, 0, len(m.OptionsWeights))
	for option := range m.OptionsWeights {
		options = append(options, option)
	}

	return question{
		ID:      m.ID,
		QuizID:  m.QuizID,
		Body:    m.Body,
		Options: options,
	}
}
