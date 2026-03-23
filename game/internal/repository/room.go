package repository

import (
	"context"

	"github.com/google/uuid"
	"gordle/game/internal/models"
)

// RoomRepository handles multiplayer room operations (Redis)
type RoomRepository interface {
	// Create creates a new room
	Create(ctx context.Context, room *models.Room) (*models.Room, error)
	// GetByID returns a room by its ID
	GetByID(ctx context.Context, roomID uuid.UUID) (*models.Room, error)
	// GetByCode returns a room by its code
	GetByCode(ctx context.Context, code string) (*models.Room, error)
	// Update updates a room
	Update(ctx context.Context, room *models.Room) error
	// Delete deletes a room
	Delete(ctx context.Context, roomID uuid.UUID) error
	// SetCode sets the room code and creates a lookup
	SetCode(ctx context.Context, roomID uuid.UUID, code string) error
	// CodeExists checks if a room code already exists
	CodeExists(ctx context.Context, code string) (bool, error)

	// AddPlayer adds a player to a room
	AddPlayer(ctx context.Context, player *models.RoomPlayer) error
	// GetPlayer returns a player from a room
	GetPlayer(ctx context.Context, roomID uuid.UUID, playerID string) (*models.RoomPlayer, error)
	// GetPlayers returns all players in a room
	GetPlayers(ctx context.Context, roomID uuid.UUID) ([]models.RoomPlayer, error)
	// UpdatePlayer updates a player in a room
	UpdatePlayer(ctx context.Context, player *models.RoomPlayer) error
	// RemovePlayer removes a player from a room
	RemovePlayer(ctx context.Context, roomID uuid.UUID, playerID string) error
	// GetPlayerCount returns the number of players in a room
	GetPlayerCount(ctx context.Context, roomID uuid.UUID) (int, error)
}
