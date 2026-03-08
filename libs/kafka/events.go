package kafka

type QuizCompletedEvent struct {
	UserID     string `json:"user_id"`
	QuizID     string `json:"quiz_id"`
	QuizResult string `json:"quiz_result"`
}

// GameCompletedEvent is published when a Wordle game is completed
type GameCompletedEvent struct {
	UserID       string   `json:"user_id"`
	SessionID    string   `json:"session_id"`
	GameMode     string   `json:"game_mode"`
	GameDate     string   `json:"game_date"`
	TargetWord   string   `json:"target_word"`
	Guesses      []string `json:"guesses"`
	Result       string   `json:"result"`
	AttemptsUsed int      `json:"attempts_used"`
}
