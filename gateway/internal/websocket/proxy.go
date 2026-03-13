package websocket

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly in production
	},
}

type Proxy struct {
	targetURL *url.URL
	logger    *slog.Logger
}

func NewProxy(targetAddr string, logger *slog.Logger) (*Proxy, error) {
	targetURL, err := url.Parse("ws://" + targetAddr)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		targetURL: targetURL,
		logger:    logger,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Build target URL
	targetURL := *p.targetURL
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery

	// Remove /api/v1 prefix if present
	targetURL.Path = strings.TrimPrefix(targetURL.Path, "/api/v1")

	p.logger.Info("proxying WebSocket",
		slog.String("path", r.URL.Path),
		slog.String("target", targetURL.String()),
	)

	// Connect to backend
	backendConn, resp, err := websocket.DefaultDialer.Dial(targetURL.String(), nil)
	if err != nil {
		p.logger.Error("failed to connect to backend",
			slog.String("error", err.Error()),
			slog.String("target", targetURL.String()),
		)
		if resp != nil {
			http.Error(w, "Failed to connect to game server", resp.StatusCode)
		} else {
			http.Error(w, "Failed to connect to game server", http.StatusBadGateway)
		}
		return
	}
	defer backendConn.Close()

	// Upgrade client connection
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Error("failed to upgrade client connection", slog.String("error", err.Error()))
		return
	}
	defer clientConn.Close()

	// Proxy messages bidirectionally
	errChan := make(chan error, 2)

	// Client -> Backend
	go func() {
		errChan <- p.copyMessages(backendConn, clientConn)
	}()

	// Backend -> Client
	go func() {
		errChan <- p.copyMessages(clientConn, backendConn)
	}()

	// Wait for either direction to fail
	<-errChan
}

func (p *Proxy) copyMessages(dst, src *websocket.Conn) error {
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := dst.WriteMessage(messageType, message); err != nil {
			return err
		}
	}
}
