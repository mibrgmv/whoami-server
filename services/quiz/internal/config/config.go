package config

import (
	"github.com/mibrgmv/whoami-server/shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server/shared/dbs/postgresql"
	"github.com/mibrgmv/whoami-server/shared/dbs/redis"
)

type Config struct {
	Grpc           *grpc.Config       `mapstructure:"grpc"`
	Postgres       *postgresql.Config `mapstructure:"postgres"`
	Redis          *redis.Config      `mapstructure:"redis"`
	HistoryService *grpc.Config       `mapstructure:"history-service"`
}
