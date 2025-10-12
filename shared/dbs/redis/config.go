package redis

import (
	"time"
)

type Config struct {
	Address    string `mapstructure:"address"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	TTLMinutes int    `mapstructure:"ttl_minutes"`
}

func (c *Config) GetTTL() time.Duration {
	return time.Duration(c.TTLMinutes) * time.Minute
}
