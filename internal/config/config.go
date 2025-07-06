package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
	"strings"
	"whoami-server/internal/config/api/grpc"
	"whoami-server/internal/config/api/http"
	"whoami-server/internal/config/auth/jwt"
	"whoami-server/internal/config/cache/redis"
	"whoami-server/internal/config/dbs/postgresql"
)

type Config struct {
	Http     *http.Config
	Grpc     *grpc.Config
	Postgres *postgresql.Config
	Redis    *redis.Config
	JWT      *jwt.Config
}

func LoanDefault() (*Config, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return nil, fmt.Errorf("failed to get called information")
	}

	mainDir := filepath.Dir(filename)
	configDir := filepath.Join(mainDir, "configs")

	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
