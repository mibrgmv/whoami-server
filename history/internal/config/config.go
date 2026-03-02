package config

import (
	"whoami-server/libs/server"
	"whoami-server/libs/storage/postgres"
)

type Config struct {
	Grpc     *server.Config   `mapstructure:"grpc"`
	Postgres *postgres.Config `mapstructure:"postgres"`
}
