package config

import (
	"gordle/libs/kafka"
	"gordle/libs/server"
	"gordle/libs/storage/postgres"
	"gordle/libs/storage/redis"
)

type Config struct {
	Grpc      *server.Config   `mapstructure:"grpc"`
	Websocket *server.Config   `mapstructure:"websocket"`
	Postgres  *postgres.Config `mapstructure:"postgres"`
	Redis     *redis.Config    `mapstructure:"redis"`
	Kafka     *kafka.Config    `mapstructure:"kafka"`
}
