package redis

import (
	"time"
)

type Config struct {
	Address  string        `mapstructure:"address"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TTL      time.Duration `mapstructure:"ttl"`
}
