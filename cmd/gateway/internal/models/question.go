package models

type Question struct {
	ID             int64                `json:"id"`
	QuizID         int64                `json:"quiz_id"`
	Body           string               `json:"body"`
	OptionsWeights map[string][]float32 `json:"options_weights"`
}
