package redis

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type Service struct {
	client *redis.Client
	ttl    time.Duration
}

func NewService(client *redis.Client, defaultTTL time.Duration) *Service {
	return &Service{
		client: client,
		ttl:    defaultTTL,
	}
}

func (s *Service) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (s *Service) Set(ctx context.Context, key string, value interface{}) error {
	return s.SetWithTTL(ctx, key, value, s.ttl)
}

func (s *Service) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Service) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *Service) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		if err := s.client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	return nil
}
