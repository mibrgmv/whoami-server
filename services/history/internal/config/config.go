package config

import (
	"github.com/mibrgmv/whoami-server/shared/grpc"
	"github.com/mibrgmv/whoami-server/shared/storage/postgres"
)

type Config struct {
	Grpc     *grpc.Config     `mapstructure:"grpc"`
	Postgres *postgres.Config `mapstructure:"postgres"`
}
