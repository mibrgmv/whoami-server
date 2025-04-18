package redis

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Address    string `mapstructure:"address"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	TTLMinutes int    `mapstructure:"ttl_minutes"`
}

func LoadDefault() (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetTTL() time.Duration {
	return time.Duration(c.TTLMinutes) * time.Minute
}
