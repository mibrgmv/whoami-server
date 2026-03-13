package websocket

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Connection struct {
	ws       *websocket.Conn
	send     chan []byte
	hub      *Hub
	roomCode string
	playerID string
	closed   bool
	mu       sync.Mutex
	logger   *slog.Logger
}

func NewConnection(ws *websocket.Conn, hub *Hub, roomCode, playerID string, logger *slog.Logger) *Connection {
	return &Connection{
		ws:       ws,
		send:     make(chan []byte, 256),
		hub:      hub,
		roomCode: roomCode,
		playerID: playerID,
		logger:   logger,
	}
}

func (c *Connection) Send(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	select {
	case c.send <- data:
	default:
		// Buffer full, close connection
		c.closed = true
		close(c.send)
	}
}

func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	close(c.send)
	c.ws.Close()
}

func (c *Connection) ReadPump(onMessage func([]byte)) {
	defer func() {
		c.hub.Unregister(c.roomCode, c.playerID)
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Warn("unexpected websocket close",
					slog.String("room", c.roomCode),
					slog.String("player", c.playerID),
					slog.String("error", err.Error()),
				)
			}
			break
		}

		if onMessage != nil {
			onMessage(message)
		}
	}
}

func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Write queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
