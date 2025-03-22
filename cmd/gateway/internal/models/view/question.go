package view

type Question struct {
	ID      int64    `json:"id"`
	QuizID  int64    `json:"quiz_id"`
	Body    string   `json:"body"`
	Options []string `json:"options"`
}
