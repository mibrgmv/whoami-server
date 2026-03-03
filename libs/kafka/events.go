package kafka

type QuizCompletedEvent struct {
	UserID     string `json:"user_id"`
	QuizID     string `json:"quiz_id"`
	QuizResult string `json:"quiz_result"`
}
