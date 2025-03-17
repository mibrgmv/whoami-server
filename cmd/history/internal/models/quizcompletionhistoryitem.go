package models

type QuizCompletionHistoryItem struct {
	ID         int64  `json:"id"`
	QuizID     int64  `json:"quiz_id"`
	UserID     int64  `json:"user_id"`
	QuizResult string `json:"quiz_result"`
}
