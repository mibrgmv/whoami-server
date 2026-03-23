package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"gordle/game/internal/models"
	"gordle/game/internal/repository"
)

const (
	roomKeyPrefix        = "wordle:room:"
	roomCodeKeyPrefix    = "wordle:room:code:"
	roomPlayersKeySuffix = ":players"
	roomActiveKey        = "wordle:room:active"
)

type roomRepo struct {
	client *goredis.Client
}

func NewRoomRepository(client *goredis.Client) repository.RoomRepository {
	return &roomRepo{client: client}
}

func (r *roomRepo) roomKey(roomID uuid.UUID) string {
	return roomKeyPrefix + roomID.String()
}

func (r *roomRepo) roomCodeKey(code string) string {
	return roomCodeKeyPrefix + code
}

func (r *roomRepo) playersKey(roomID uuid.UUID) string {
	return roomKeyPrefix + roomID.String() + roomPlayersKeySuffix
}

func (r *roomRepo) Create(ctx context.Context, room *models.Room) (*models.Room, error) {
	data, err := json.Marshal(room)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal room: %w", err)
	}

	key := r.roomKey(room.ID)
	ttl := room.ExpiresAt.Sub(room.CreatedAt)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, roomActiveKey, room.ID.String())

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to save room: %w", err)
	}

	return room, nil
}

func (r *roomRepo) GetByID(ctx context.Context, roomID uuid.UUID) (*models.Room, error) {
	key := r.roomKey(roomID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	var room models.Room
	if err := json.Unmarshal(data, &room); err != nil {
		return nil, fmt.Errorf("failed to unmarshal room: %w", err)
	}

	return &room, nil
}

func (r *roomRepo) GetByCode(ctx context.Context, code string) (*models.Room, error) {
	codeKey := r.roomCodeKey(code)
	roomIDStr, err := r.client.Get(ctx, codeKey).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get room ID by code: %w", err)
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid room ID in code key: %w", err)
	}

	return r.GetByID(ctx, roomID)
}

func (r *roomRepo) Update(ctx context.Context, room *models.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}

	key := r.roomKey(room.ID)

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		ttl = models.RoomTTL
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}

	return nil
}

func (r *roomRepo) Delete(ctx context.Context, roomID uuid.UUID) error {
	room, err := r.GetByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return nil
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.roomKey(roomID))
	pipe.Del(ctx, r.playersKey(roomID))
	if room.Code != "" {
		pipe.Del(ctx, r.roomCodeKey(room.Code))
	}
	pipe.SRem(ctx, roomActiveKey, roomID.String())

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
}

func (r *roomRepo) SetCode(ctx context.Context, roomID uuid.UUID, code string) error {
	room, err := r.GetByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return fmt.Errorf("room not found")
	}

	ttl := room.ExpiresAt.Sub(room.CreatedAt)
	codeKey := r.roomCodeKey(code)

	if err := r.client.Set(ctx, codeKey, roomID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to set room code: %w", err)
	}

	room.Code = code
	return r.Update(ctx, room)
}

func (r *roomRepo) CodeExists(ctx context.Context, code string) (bool, error) {
	codeKey := r.roomCodeKey(code)
	exists, err := r.client.Exists(ctx, codeKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}
	return exists > 0, nil
}

func (r *roomRepo) AddPlayer(ctx context.Context, player *models.RoomPlayer) error {
	data, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("failed to marshal player: %w", err)
	}

	key := r.playersKey(player.RoomID)
	if err := r.client.HSet(ctx, key, player.PlayerID, data).Err(); err != nil {
		return fmt.Errorf("failed to add player: %w", err)
	}

	return nil
}

func (r *roomRepo) GetPlayer(ctx context.Context, roomID uuid.UUID, playerID string) (*models.RoomPlayer, error) {
	key := r.playersKey(roomID)
	data, err := r.client.HGet(ctx, key, playerID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	var player models.RoomPlayer
	if err := json.Unmarshal(data, &player); err != nil {
		return nil, fmt.Errorf("failed to unmarshal player: %w", err)
	}

	player.RoomID = roomID
	return &player, nil
}

func (r *roomRepo) GetPlayers(ctx context.Context, roomID uuid.UUID) ([]models.RoomPlayer, error) {
	key := r.playersKey(roomID)
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	players := make([]models.RoomPlayer, 0, len(data))
	for _, playerData := range data {
		var player models.RoomPlayer
		if err := json.Unmarshal([]byte(playerData), &player); err != nil {
			return nil, fmt.Errorf("failed to unmarshal player: %w", err)
		}
		player.RoomID = roomID
		players = append(players, player)
	}

	return players, nil
}

func (r *roomRepo) UpdatePlayer(ctx context.Context, player *models.RoomPlayer) error {
	return r.AddPlayer(ctx, player)
}

func (r *roomRepo) RemovePlayer(ctx context.Context, roomID uuid.UUID, playerID string) error {
	key := r.playersKey(roomID)
	if err := r.client.HDel(ctx, key, playerID).Err(); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}
	return nil
}

func (r *roomRepo) GetPlayerCount(ctx context.Context, roomID uuid.UUID) (int, error) {
	key := r.playersKey(roomID)
	count, err := r.client.HLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get player count: %w", err)
	}
	return int(count), nil
}
