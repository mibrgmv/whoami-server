package models

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	roomv1 "gordle/game/pkg/protogen/room/v1"
)

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

type RoomMode string

const (
	RoomModeSingleRound RoomMode = "single_round"
)

const (
	RoomCodeLength    = 6
	RoomCodeAlphabet  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // No I/O/0/1
	DefaultMaxPlayers = 6
	RoomTTL           = 24 * time.Hour
)

type RoomSettings struct {
	Mode        RoomMode `json:"mode"`
	MaxPlayers  int      `json:"maxPlayers"`
	ShowGuesses bool     `json:"showGuesses"`
}

type Room struct {
	ID          uuid.UUID    `json:"id"`
	Code        string       `json:"code"`
	HostID      string       `json:"hostId"`
	Status      RoomStatus   `json:"status"`
	Settings    RoomSettings `json:"settings"`
	CurrentWord string       `json:"currentWord,omitempty"`
	RoundNumber int          `json:"roundNumber"`
	Language    string       `json:"language"`
	CreatedAt   time.Time    `json:"createdAt"`
	ExpiresAt   time.Time    `json:"expiresAt"`
}

func GenerateRoomCode() (string, error) {
	b := make([]byte, RoomCodeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := make([]byte, RoomCodeLength)
	for i := 0; i < RoomCodeLength; i++ {
		code[i] = RoomCodeAlphabet[int(b[i])%len(RoomCodeAlphabet)]
	}
	return string(code), nil
}

func NewRoom(hostID, language string, settings RoomSettings) *Room {
	now := time.Now()
	if settings.MaxPlayers == 0 {
		settings.MaxPlayers = DefaultMaxPlayers
	}
	if settings.Mode == "" {
		settings.Mode = RoomModeSingleRound
	}
	return &Room{
		ID:          uuid.New(),
		HostID:      hostID,
		Status:      RoomStatusWaiting,
		Settings:    settings,
		RoundNumber: 0,
		Language:    language,
		CreatedAt:   now,
		ExpiresAt:   now.Add(RoomTTL),
	}
}

func (r *Room) ToProto() *roomv1.Room {
	room := &roomv1.Room{
		Id:          r.ID.String(),
		Code:        r.Code,
		HostId:      r.HostID,
		Status:      r.Status.ToProto(),
		Settings:    r.Settings.ToProto(),
		RoundNumber: int32(r.RoundNumber),
		Language:    r.Language,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		ExpiresAt:   r.ExpiresAt.Format(time.RFC3339),
	}
	return room
}

func (s RoomStatus) ToProto() roomv1.RoomStatus {
	switch s {
	case RoomStatusWaiting:
		return roomv1.RoomStatus_ROOM_STATUS_WAITING
	case RoomStatusPlaying:
		return roomv1.RoomStatus_ROOM_STATUS_PLAYING
	case RoomStatusFinished:
		return roomv1.RoomStatus_ROOM_STATUS_FINISHED
	default:
		return roomv1.RoomStatus_ROOM_STATUS_UNSPECIFIED
	}
}

func RoomModeFromProto(mode roomv1.RoomMode) RoomMode {
	return RoomModeSingleRound
}

func (m RoomMode) ToProto() roomv1.RoomMode {
	return roomv1.RoomMode_ROOM_MODE_SINGLE_ROUND
}

func (s *RoomSettings) ToProto() *roomv1.RoomSettings {
	return &roomv1.RoomSettings{
		Mode:        s.Mode.ToProto(),
		MaxPlayers:  int32(s.MaxPlayers),
		ShowGuesses: s.ShowGuesses,
	}
}

func RoomSettingsFromProto(s *roomv1.RoomSettings) RoomSettings {
	return RoomSettings{
		Mode:        RoomModeFromProto(s.GetMode()),
		MaxPlayers:  int(s.GetMaxPlayers()),
		ShowGuesses: s.GetShowGuesses(),
	}
}
