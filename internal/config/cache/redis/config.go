package redis

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
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
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	redisConfig := viper.Sub("redis")
	if redisConfig == nil {
		return nil, fmt.Errorf("redis configuration section not found")
	}

	var cfg Config
	if err := redisConfig.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetTTL() time.Duration {
	return time.Duration(c.TTLMinutes) * time.Minute
}
