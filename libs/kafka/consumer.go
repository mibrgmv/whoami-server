package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers         []string
	GroupID         string
	Topic           string
	AutoOffsetReset string // "earliest" or "latest"
	MaxWait         time.Duration
	MinBytes        int
	MaxBytes        int
}

type MessageHandler func(ctx context.Context, message kafka.Message) error

type Consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	brokers []string
	topic   string
}

func NewConsumer(cfg ConsumerConfig, handler MessageHandler) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		CommitInterval: 0, // manual
		Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
		brokers: cfg.Brokers,
		topic:   cfg.Topic,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	if err := WaitForTopic(ctx, c.brokers, c.topic, 3*time.Second); err != nil {
		log.Printf("Failed waiting for topic %q: %v", c.topic, err)
		return
	}
	c.consumeLoop(ctx)
}

func (c *Consumer) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Error fetching message: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Error processing message: %v", err)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			} else {
				log.Printf("Successfully processed and committed message at offset %d", msg.Offset)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := c.handler(processCtx, msg); err != nil {
		return fmt.Errorf("message handler failed: %w", err)
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
