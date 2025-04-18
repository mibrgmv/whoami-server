package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	config "whoami-server/internal/config/cache/redis"
)

func NewClient(ctx context.Context, config *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})

	err := client.Ping(ctx).Err()
	return client, err
}
