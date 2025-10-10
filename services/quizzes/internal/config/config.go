package config

import (
	"github.com/mibrgmv/whoami-server/shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server/shared/config/cache/redis"
	"github.com/mibrgmv/whoami-server/shared/config/dbs/postgresql"
)

type Config struct {
	Grpc           *grpc.Config       `mapstructure:"grpc"`
	Postgres       *postgresql.Config `mapstructure:"postgres"`
	Redis          *redis.Config      `mapstructure:"redis"`
	HistoryService *grpc.Config       `mapstructure:"history-service"`
}
