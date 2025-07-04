package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"whoami-server/internal/config/api/grpc"
	"whoami-server/internal/config/api/http"
	"whoami-server/internal/config/dbs/postgresql"
)

type Config struct {
	Http     http.Config
	Grpc     grpc.Config
	Postgres postgresql.Config
}

func GetDefaultForService(serviceName string) (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg map[string]Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	serviceCfg, ok := cfg[serviceName]
	if !ok {
		return nil, fmt.Errorf("service config not found: %s", serviceName)
	}

	return &serviceCfg, nil
}
