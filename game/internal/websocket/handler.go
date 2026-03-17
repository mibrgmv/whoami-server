package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"gordle/game/internal/models"
	"gordle/game/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now; configure in production
	},
}

type Handler struct {
	hub         *Hub
	roomService service.RoomService
	logger      *slog.Logger
}

func NewHandler(hub *Hub, roomService service.RoomService, logger *slog.Logger) *Handler {
	return &Handler{
		hub:         hub,
		roomService: roomService,
		logger:      logger,
	}
}

type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type GuessPayload struct {
	Word string `json:"word"`
}

type ReadyPayload struct {
	Ready bool `json:"ready"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Expected path: /rooms/{code}/ws
	path := strings.TrimPrefix(r.URL.Path, "/rooms/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "ws" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	roomCode := parts[0]
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		http.Error(w, "player_id required", http.StatusBadRequest)
		return
	}

	room, players, err := h.roomService.GetRoom(r.Context(), roomCode)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	playerFound := false
	for _, p := range players {
		if p.PlayerID == playerID {
			playerFound = true
			break
		}
	}
	if !playerFound {
		http.Error(w, "player not in room", http.StatusForbidden)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	conn := NewConnection(ws, h.hub, roomCode, playerID, h.logger)
	h.hub.Register(roomCode, playerID, conn)

	go conn.WritePump()
	go conn.ReadPump(func(message []byte) {
		h.handleMessage(room.Code, playerID, message)
	})
}

func (h *Handler) handleMessage(roomCode, playerID string, message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		h.sendError(roomCode, playerID, "invalid_message", "failed to parse message")
		return
	}

	switch msg.Type {
	case "guess":
		h.handleGuess(roomCode, playerID, msg.Payload)
	case "ready":
		h.handleReady(roomCode, playerID, msg.Payload)
	case "start_game":
		h.handleStartGame(roomCode, playerID)
	case "next_round":
		h.handleNextRound(roomCode, playerID)
	case "leave":
		h.handleLeave(roomCode, playerID)
	default:
		h.sendError(roomCode, playerID, "unknown_type", "unknown message type")
	}
}

func (h *Handler) handleGuess(roomCode, playerID string, payload json.RawMessage) {
	var p GuessPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		h.sendError(roomCode, playerID, "invalid_payload", "failed to parse guess payload")
		return
	}

	_, _, _, err := h.roomService.SubmitGuess(context.Background(), roomCode, playerID, p.Word)
	if err != nil {
		h.sendError(roomCode, playerID, "guess_failed", err.Error())
		return
	}
}

func (h *Handler) handleReady(roomCode, playerID string, payload json.RawMessage) {
	var p ReadyPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		h.sendError(roomCode, playerID, "invalid_payload", "failed to parse ready payload")
		return
	}

	if err := h.roomService.SetReady(context.Background(), roomCode, playerID, p.Ready); err != nil {
		h.sendError(roomCode, playerID, "ready_failed", err.Error())
		return
	}
}

func (h *Handler) handleStartGame(roomCode, playerID string) {
	if _, err := h.roomService.StartGame(context.Background(), roomCode, playerID); err != nil {
		h.sendError(roomCode, playerID, "start_game_failed", err.Error())
		return
	}
}

func (h *Handler) handleNextRound(roomCode, playerID string) {
	if _, err := h.roomService.NextRound(context.Background(), roomCode, playerID); err != nil {
		h.sendError(roomCode, playerID, "next_round_failed", err.Error())
		return
	}
}

func (h *Handler) handleLeave(roomCode, playerID string) {
	if err := h.roomService.LeaveRoom(context.Background(), roomCode, playerID); err != nil {
		h.sendError(roomCode, playerID, "leave_failed", err.Error())
		return
	}

	h.hub.Unregister(roomCode, playerID)
}

func (h *Handler) sendError(roomCode, playerID, code, message string) {
	event, err := models.NewWSEvent(models.WSEventError, models.ErrorPayload{
		Code:    code,
		Message: message,
	})
	if err != nil {
		return
	}

	h.hub.SendToPlayer(roomCode, playerID, event)
}
