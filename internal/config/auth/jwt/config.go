package jwt

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Config struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
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

	jwtConfig := viper.Sub("jwt")
	if jwtConfig == nil {
		return nil, fmt.Errorf("jwt configuration section not found")
	}

	var cfg Config
	if err := jwtConfig.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jwt config: %w", err)
	}

	return &cfg, nil
}
