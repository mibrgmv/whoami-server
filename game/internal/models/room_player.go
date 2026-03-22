package models

import (
	"time"

	"github.com/google/uuid"
	roomv1 "gordle/game/pkg/protogen/room/v1"
)

type PlayerStatus string

const (
	PlayerStatusWaiting  PlayerStatus = "waiting"
	PlayerStatusReady    PlayerStatus = "ready"
	PlayerStatusPlaying  PlayerStatus = "playing"
	PlayerStatusFinished PlayerStatus = "finished"
)

type PlayerResult string

const (
	PlayerResultNone PlayerResult = ""
	PlayerResultWon  PlayerResult = "won"
	PlayerResultLost PlayerResult = "lost"
)

type RoomPlayer struct {
	RoomID          uuid.UUID    `json:"-"`
	PlayerID        string       `json:"playerId"`
	IsGuest         bool         `json:"isGuest"`
	DisplayName     string       `json:"displayName"`
	Status          PlayerStatus `json:"status"`
	Result          PlayerResult `json:"result"`
	CurrentAttempts int          `json:"currentAttempts"`
	Guesses         []Guess      `json:"guesses"`
	FinishedAt      *time.Time   `json:"finishedAt,omitempty"`
}

func NewRoomPlayer(roomID uuid.UUID, playerID uuid.UUID, isGuest bool, displayName string) *RoomPlayer {
	return &RoomPlayer{
		RoomID:          roomID,
		PlayerID:        playerID.String(),
		IsGuest:         isGuest,
		DisplayName:     displayName,
		Status:          PlayerStatusWaiting,
		Result:          PlayerResultNone,
		CurrentAttempts: 0,
		Guesses:         []Guess{},
	}
}

func (p *RoomPlayer) ToProto() *roomv1.RoomPlayer {
	player := &roomv1.RoomPlayer{
		PlayerId:        p.PlayerID,
		DisplayName:     p.DisplayName,
		Status:          p.Status.ToProto(),
		Result:          p.Result.ToProto(),
		CurrentAttempts: int32(p.CurrentAttempts),
	}
	if p.IsGuest {
		player.GuestId = p.PlayerID
	} else {
		player.UserId = p.PlayerID
	}
	for _, g := range p.Guesses {
		player.Guesses = append(player.Guesses, g.ToRoomProto())
	}
	if p.FinishedAt != nil {
		player.FinishedAt = p.FinishedAt.Format(time.RFC3339)
	}
	return player
}

func (s PlayerStatus) ToProto() roomv1.PlayerStatus {
	switch s {
	case PlayerStatusWaiting:
		return roomv1.PlayerStatus_PLAYER_STATUS_WAITING
	case PlayerStatusReady:
		return roomv1.PlayerStatus_PLAYER_STATUS_READY
	case PlayerStatusPlaying:
		return roomv1.PlayerStatus_PLAYER_STATUS_PLAYING
	case PlayerStatusFinished:
		return roomv1.PlayerStatus_PLAYER_STATUS_FINISHED
	default:
		return roomv1.PlayerStatus_PLAYER_STATUS_UNSPECIFIED
	}
}

func (r PlayerResult) ToProto() roomv1.PlayerResult {
	switch r {
	case PlayerResultWon:
		return roomv1.PlayerResult_PLAYER_RESULT_WON
	case PlayerResultLost:
		return roomv1.PlayerResult_PLAYER_RESULT_LOST
	default:
		return roomv1.PlayerResult_PLAYER_RESULT_UNSPECIFIED
	}
}
