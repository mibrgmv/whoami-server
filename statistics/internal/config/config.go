package config

import (
	"gordle/libs/kafka"
	"gordle/libs/server"
	"gordle/libs/storage/postgres"
)

type Config struct {
	Grpc     *server.Config   `mapstructure:"grpc"`
	Postgres *postgres.Config `mapstructure:"postgres"`
	Kafka    *kafka.Config    `mapstructure:"kafka"`
}
