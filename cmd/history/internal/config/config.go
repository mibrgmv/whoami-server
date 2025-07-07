package config

import (
	"whoami-server/internal/config/api/grpc"
	"whoami-server/internal/config/dbs/postgresql"
)

type Config struct {
	Grpc     *grpc.Config       `mapstructure:"grpc"`
	Postgres *postgresql.Config `mapstructure:"postgres"`
}
