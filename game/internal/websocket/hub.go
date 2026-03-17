package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"gordle/game/internal/models"
)

type Hub struct {
	rooms      map[string]map[string]*Connection // roomCode -> playerID -> conn
	mu         sync.RWMutex
	register   chan *Registration
	unregister chan *Registration
	broadcast  chan *BroadcastMessage
	logger     *slog.Logger
	pubsub     *PubSub
}

type Registration struct {
	RoomCode   string
	PlayerID   string
	Connection *Connection
}

type BroadcastMessage struct {
	RoomCode string
	PlayerID string // empty for broadcast to all
	Event    *models.WSEvent
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		rooms:      make(map[string]map[string]*Connection),
		register:   make(chan *Registration, 256),
		unregister: make(chan *Registration, 256),
		broadcast:  make(chan *BroadcastMessage, 256),
		logger:     logger,
	}
}

func (h *Hub) SetPubSub(pubsub *PubSub) {
	h.pubsub = pubsub
}

func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			h.handleRegister(reg)
		case reg := <-h.unregister:
			h.handleUnregister(reg)
		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *Hub) handleRegister(reg *Registration) {
	h.mu.Lock()

	isFirstInRoom := false
	if _, ok := h.rooms[reg.RoomCode]; !ok {
		h.rooms[reg.RoomCode] = make(map[string]*Connection)
		isFirstInRoom = true
	}

	if existing, ok := h.rooms[reg.RoomCode][reg.PlayerID]; ok {
		existing.Close()
	}

	h.rooms[reg.RoomCode][reg.PlayerID] = reg.Connection
	h.mu.Unlock()

	// Subscribe to Redis channel when first local player joins
	if isFirstInRoom && h.pubsub != nil {
		if err := h.pubsub.Subscribe(context.Background(), reg.RoomCode); err != nil {
			h.logger.Error("failed to subscribe to room channel",
				"room", reg.RoomCode,
				"error", err,
			)
		}
	}

	h.logger.Info("player connected",
		"room", reg.RoomCode,
		"player", reg.PlayerID,
	)
}

func (h *Hub) handleUnregister(reg *Registration) {
	h.mu.Lock()

	isLastInRoom := false
	if room, ok := h.rooms[reg.RoomCode]; ok {
		if conn, ok := room[reg.PlayerID]; ok {
			conn.Close()
			delete(room, reg.PlayerID)
			h.logger.Info("player disconnected",
				"room", reg.RoomCode,
				"player", reg.PlayerID,
			)
		}

		if len(room) == 0 {
			delete(h.rooms, reg.RoomCode)
			isLastInRoom = true
		}
	}
	h.mu.Unlock()

	// Unsubscribe from Redis channel when last local player leaves
	if isLastInRoom && h.pubsub != nil {
		if err := h.pubsub.Unsubscribe(reg.RoomCode); err != nil {
			h.logger.Error("failed to unsubscribe from room channel",
				"room", reg.RoomCode,
				"error", err,
			)
		}
	}
}

func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	// If sending to specific player, deliver locally only
	if msg.PlayerID != "" {
		h.deliverToPlayer(msg.RoomCode, msg.PlayerID, msg.Event)
		return
	}

	// Broadcast to all: publish via Redis if available
	if h.pubsub != nil {
		if err := h.pubsub.Publish(context.Background(), msg.RoomCode, msg.Event); err != nil {
			h.logger.Error("failed to publish to redis",
				"room", msg.RoomCode,
				"error", err,
			)
			h.deliverLocal(msg.RoomCode, msg.Event)
		}
	} else {
		h.deliverLocal(msg.RoomCode, msg.Event)
	}
}

// deliverLocal sends an event to all local connections in a room
func (h *Hub) deliverLocal(roomCode string, event *models.WSEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[roomCode]
	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event", "error", err)
		return
	}

	for _, conn := range room {
		conn.Send(data)
	}
}

// deliverToPlayer sends an event to a specific player (local only)
func (h *Hub) deliverToPlayer(roomCode, playerID string, event *models.WSEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[roomCode]
	if !ok {
		return
	}

	conn, ok := room[playerID]
	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event", "error", err)
		return
	}

	conn.Send(data)
}

func (h *Hub) Register(roomCode, playerID string, conn *Connection) {
	h.register <- &Registration{
		RoomCode:   roomCode,
		PlayerID:   playerID,
		Connection: conn,
	}
}

func (h *Hub) Unregister(roomCode, playerID string) {
	h.unregister <- &Registration{
		RoomCode: roomCode,
		PlayerID: playerID,
	}
}

func (h *Hub) BroadcastToRoom(roomCode string, event *models.WSEvent) {
	h.broadcast <- &BroadcastMessage{
		RoomCode: roomCode,
		Event:    event,
	}
}

func (h *Hub) SendToPlayer(roomCode, playerID string, event *models.WSEvent) {
	h.broadcast <- &BroadcastMessage{
		RoomCode: roomCode,
		PlayerID: playerID,
		Event:    event,
	}
}

func (h *Hub) GetConnectionCount(roomCode string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[roomCode]; ok {
		return len(room)
	}
	return 0
}
