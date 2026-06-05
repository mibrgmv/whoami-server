package config

import (
	"time"

	"gordle/libs/kafka"
	"gordle/libs/server"
	"gordle/libs/storage/postgres"
)

type Config struct {
	Grpc      *server.Config   `mapstructure:"grpc"`
	Websocket *server.Config   `mapstructure:"websocket"`
	Postgres  *postgres.Config `mapstructure:"postgres"`
	Redis     *struct {
		Address  string        `mapstructure:"address"`
		Password string        `mapstructure:"password"`
		DB       int           `mapstructure:"db"`
		TTL      time.Duration `mapstructure:"ttl"`
	} `mapstructure:"redis"`
	Kafka *kafka.Config `mapstructure:"kafka"`
}
