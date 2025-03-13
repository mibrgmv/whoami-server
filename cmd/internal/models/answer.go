package models

type Answer struct {
	QuizID     int64  `json:"quiz_id"`
	QuestionID int64  `json:"question_id"`
	Body       string `json:"body"`
}
