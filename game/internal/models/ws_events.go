package models

import (
	"encoding/json"
	"time"
)

type WSEventType string

const (
	WSEventPlayerJoined  WSEventType = "player_joined"
	WSEventPlayerLeft    WSEventType = "player_left"
	WSEventPlayerReady   WSEventType = "player_ready"
	WSEventGameStarted   WSEventType = "game_started"
	WSEventPlayerAttempt WSEventType = "player_attempt"
	WSEventPlayerGuess   WSEventType = "player_guess"
	WSEventRoundEnded    WSEventType = "round_ended"
	WSEventGameEnded     WSEventType = "game_ended"
	WSEventRoomUpdated   WSEventType = "room_updated"
	WSEventError         WSEventType = "error"
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
	PlayerID    string `json:"playerId"`
	DisplayName string `json:"displayName"`
}

type PlayerReadyPayload struct {
	PlayerID    string `json:"playerId"`
	DisplayName string `json:"displayName"`
	Ready       bool   `json:"ready"`
}

type GameStartedPayload struct {
	RoundNumber int `json:"roundNumber"`
	WordLength  int `json:"wordLength"`
}

type PlayerAttemptPayload struct {
	PlayerID    string `json:"playerId"`
	DisplayName string `json:"displayName"`
	Attempts    int    `json:"attempts"`
	Solved      bool   `json:"solved"`
}

type PlayerGuessPayload struct {
	PlayerID    string `json:"playerId"`
	DisplayName string `json:"displayName"`
	GuessWord   string `json:"guessWord"`
	Result      string `json:"result"`
	Attempts    int    `json:"attempts"`
	Solved      bool   `json:"solved"`
}

type RoundEndedPayload struct {
	RoundNumber int           `json:"roundNumber"`
	TargetWord  string        `json:"targetWord"`
	Results     []PlayerScore `json:"results"`
}

type PlayerScore struct {
	PlayerID    string `json:"playerId"`
	DisplayName string `json:"displayName"`
	Result      string `json:"result"`
	Attempts    int    `json:"attempts"`
	Score       int    `json:"score"`
}

type GameEndedPayload struct {
	Reason      string        `json:"reason"`
	FinalScores []PlayerScore `json:"finalScores"`
	TargetWord  string        `json:"targetWord,omitempty"`
}

type RoomUpdatedPayload struct {
	Room *Room `json:"room"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
