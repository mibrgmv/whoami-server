package models

import (
	"encoding/json"
	"time"
)

type WSEventType string

const (
	WSEventPlayerJoined WSEventType = "player_joined"
	WSEventPlayerLeft   WSEventType = "player_left"
	WSEventPlayerReady  WSEventType = "player_ready"
	WSEventGameStarted  WSEventType = "game_started"
	WSEventPlayerGuess  WSEventType = "player_guess"
	WSEventRoundEnded   WSEventType = "round_ended"
	WSEventGameEnded    WSEventType = "game_ended"
	WSEventRoomUpdated  WSEventType = "room_updated"
	WSEventError        WSEventType = "error"
)

type WSEvent struct {
	Type      WSEventType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

func NewWSEvent(eventType WSEventType, payload any) (*WSEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &WSEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   data,
	}, nil
}

type PlayerJoinedPayload struct {
	Player *RoomPlayer `json:"player"`
}

type PlayerLeftPayload struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
}

type PlayerReadyPayload struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	Ready       bool   `json:"ready"`
}

type GameStartedPayload struct {
	RoundNumber int `json:"round_number"`
	WordLength  int `json:"word_length"`
}

type PlayerGuessPayload struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	GuessWord   string `json:"guess_word,omitempty"`
	Result      string `json:"result,omitempty"`
	Attempts    int    `json:"attempts"`
	Solved      bool   `json:"solved"`
}

type RoundEndedPayload struct {
	RoundNumber int           `json:"round_number"`
	TargetWord  string        `json:"target_word"`
	Results     []PlayerScore `json:"results"`
}

type PlayerScore struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	Result      string `json:"result"`
	Attempts    int    `json:"attempts"`
	Score       int    `json:"score"`
}

type GameEndedPayload struct {
	Reason      string        `json:"reason"`
	FinalScores []PlayerScore `json:"final_scores"`
	TargetWord  string        `json:"target_word,omitempty"`
}

type RoomUpdatedPayload struct {
	Room *Room `json:"room"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
