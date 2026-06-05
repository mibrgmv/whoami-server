package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

func EnsureTopic(ctx context.Context, brokers []string, topic string, numPartitions int) error {
	var conn *kafka.Conn
	var err error

	for _, broker := range brokers {
		conn, err = kafka.DialContext(ctx, "tcp", broker)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to connect to kafka brokers: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: 1,
	})
	if err != nil {
		var kafkaErr kafka.Error
		if errors.As(err, &kafkaErr) && kafkaErr == kafka.TopicAlreadyExists {
			log.Printf("Topic %q already exists", topic)
			return nil
		}
		return fmt.Errorf("failed to create topic %q: %w", topic, err)
	}

	log.Printf("Created topic %q with %d partitions", topic, numPartitions)
	return nil
}

func WaitForTopic(ctx context.Context, brokers []string, topic string, pollInterval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if topicExists(brokers, topic) {
			log.Printf("Topic %q is available", topic)
			return nil
		}

		log.Printf("Topic %q not found, retrying in %v...", topic, pollInterval)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func topicExists(brokers []string, topic string) bool {
	for _, broker := range brokers {
		conn, err := kafka.Dial("tcp", broker)
		if err != nil {
			continue
		}

		partitions, err := conn.ReadPartitions(topic)
		conn.Close()
		if err == nil && len(partitions) > 0 {
			return true
		}
	}
	return false
}
