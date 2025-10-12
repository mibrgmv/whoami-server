package config

import (
	"github.com/mibrgmv/whoami-server/shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server/shared/dbs/postgresql"
)

type Config struct {
	Grpc     *grpc.Config       `mapstructure:"grpc"`
	Postgres *postgresql.Config `mapstructure:"postgres"`
}
