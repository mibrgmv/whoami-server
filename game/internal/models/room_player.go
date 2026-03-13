package models

import (
	"fmt"
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
	RoomID          uuid.UUID    `json:"room_id"`
	UserID          *uuid.UUID   `json:"user_id,omitempty"`
	GuestID         *string      `json:"guest_id,omitempty"`
	DisplayName     string       `json:"display_name"`
	Status          PlayerStatus `json:"status"`
	Result          PlayerResult `json:"result"`
	CurrentAttempts int          `json:"current_attempts"`
	TotalScore      int          `json:"total_score"`
	Guesses         []Guess      `json:"guesses"`
	FinishedAt      *time.Time   `json:"finished_at,omitempty"`
}

func (p *RoomPlayer) PlayerID() string {
	if p.UserID != nil {
		return p.UserID.String()
	}
	if p.GuestID != nil {
		return "guest:" + *p.GuestID
	}
	return ""
}

func NewRoomPlayer(roomID uuid.UUID, userID *uuid.UUID, guestID *string, displayName string) *RoomPlayer {
	return &RoomPlayer{
		RoomID:          roomID,
		UserID:          userID,
		GuestID:         guestID,
		DisplayName:     displayName,
		Status:          PlayerStatusWaiting,
		Result:          PlayerResultNone,
		CurrentAttempts: 0,
		TotalScore:      0,
		Guesses:         []Guess{},
	}
}

func (p *RoomPlayer) ToProto() *roomv1.RoomPlayer {
	player := &roomv1.RoomPlayer{
		PlayerId:        p.PlayerID(),
		DisplayName:     p.DisplayName,
		Status:          p.Status.ToProto(),
		Result:          p.Result.ToProto(),
		CurrentAttempts: int32(p.CurrentAttempts),
		TotalScore:      int32(p.TotalScore),
	}
	if p.UserID != nil {
		player.UserId = p.UserID.String()
	}
	if p.GuestID != nil {
		player.GuestId = *p.GuestID
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

func ParsePlayerID(playerID string) (*uuid.UUID, *string, error) {
	if len(playerID) > 6 && playerID[:6] == "guest:" {
		guestID := playerID[6:]
		return nil, &guestID, nil
	}
	userID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid player ID: %w", err)
	}
	return &userID, nil, nil
}
