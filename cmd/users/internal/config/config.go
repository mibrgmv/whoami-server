package config

import (
	"whoami-server/internal/config/api/grpc"
	"whoami-server/internal/config/auth/jwt"
	"whoami-server/internal/config/cache/redis"
	"whoami-server/internal/config/dbs/postgresql"
)

type Config struct {
	JWT      *jwt.Config
	Grpc     *grpc.Config
	Postgres *postgresql.Config
	Redis    *redis.Config
}
