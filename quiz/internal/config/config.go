package config

import (
	"whoami-server/libs/server"
	"whoami-server/libs/storage/postgres"
	"whoami-server/libs/storage/redis"
)

type Config struct {
	Grpc           *server.Config   `mapstructure:"grpc"`
	Postgres       *postgres.Config `mapstructure:"postgres"`
	Redis          *redis.Config    `mapstructure:"redis"`
	HistoryService *server.Config   `mapstructure:"history-service"`
}
