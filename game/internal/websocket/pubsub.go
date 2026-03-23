package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"gordle/game/internal/models"
)

const roomChannelPrefix = "ws:room:"

type PubSub struct {
	client        *redis.Client
	subscriptions map[string]*redis.PubSub // roomCode -> subscription
	mu            sync.RWMutex
	hub           *Hub
	logger        *slog.Logger
}

func NewPubSub(client *redis.Client, hub *Hub, logger *slog.Logger) *PubSub {
	return &PubSub{
		client:        client,
		subscriptions: make(map[string]*redis.PubSub),
		hub:           hub,
		logger:        logger,
	}
}

func (p *PubSub) channelName(roomCode string) string {
	return roomChannelPrefix + roomCode
}

// Publish sends an event to all instances via Redis
func (p *PubSub) Publish(ctx context.Context, roomCode string, event *models.WSEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := p.channelName(roomCode)
	if err := p.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish to redis: %w", err)
	}

	return nil
}

// Subscribe starts listening for events on a room channel
func (p *PubSub) Subscribe(ctx context.Context, roomCode string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.subscriptions[roomCode]; exists {
		return nil // already subscribed
	}

	channel := p.channelName(roomCode)
	sub := p.client.Subscribe(ctx, channel)

	// Wait for subscription confirmation
	if _, err := sub.Receive(ctx); err != nil {
		sub.Close()
		return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}

	p.subscriptions[roomCode] = sub
	p.logger.Info("subscribed to room channel", "room", roomCode)

	// Start goroutine to handle messages
	go p.handleMessages(ctx, roomCode, sub)

	return nil
}

// Unsubscribe stops listening for events on a room channel
func (p *PubSub) Unsubscribe(roomCode string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub, exists := p.subscriptions[roomCode]
	if !exists {
		return nil
	}

	if err := sub.Close(); err != nil {
		return fmt.Errorf("failed to unsubscribe from channel: %w", err)
	}

	delete(p.subscriptions, roomCode)
	p.logger.Info("unsubscribed from room channel", "room", roomCode)

	return nil
}

func (p *PubSub) handleMessages(ctx context.Context, roomCode string, sub *redis.PubSub) {
	ch := sub.Channel()

	for msg := range ch {
		var event models.WSEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			p.logger.Error("failed to unmarshal event from redis",
				"room", roomCode,
				"error", err,
			)
			continue
		}

		// Deliver to local connections only
		p.hub.deliverLocal(roomCode, &event)
	}

	p.logger.Debug("message handler stopped", "room", roomCode)
}

// Close closes all subscriptions
func (p *PubSub) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for roomCode, sub := range p.subscriptions {
		if err := sub.Close(); err != nil {
			p.logger.Error("failed to close subscription",
				"room", roomCode,
				"error", err,
			)
		}
	}

	p.subscriptions = make(map[string]*redis.PubSub)
	return nil
}
