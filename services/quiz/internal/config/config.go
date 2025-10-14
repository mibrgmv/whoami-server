package config

import (
	"github.com/mibrgmv/whoami-server/shared/grpc"
	"github.com/mibrgmv/whoami-server/shared/storage/postgres"
	"github.com/mibrgmv/whoami-server/shared/storage/redis"
)

type Config struct {
	Grpc           *grpc.Config     `mapstructure:"grpc"`
	Postgres       *postgres.Config `mapstructure:"postgres"`
	Redis          *redis.Config    `mapstructure:"redis"`
	HistoryService *grpc.Config     `mapstructure:"history-service"`
}
