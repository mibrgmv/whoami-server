package redis

import (
	"context"

	config "github.com/mibrgmv/whoami-server/shared/config/cache/redis"
	"github.com/redis/go-redis/v9"
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
