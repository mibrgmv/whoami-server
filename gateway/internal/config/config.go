package config

import (
	"gordle/libs/keycloak"
	"gordle/libs/server"
)

type Config struct {
	Keycloak          keycloak.Config   `mapstructure:"keycloak"`
	HTTP              server.HTTPConfig `mapstructure:"http"`
	Metrics           server.Config     `mapstructure:"metrics"`
	GameService       server.Config     `mapstructure:"game_service"`
	GameWebSocket     server.Config     `mapstructure:"game_websocket"`
	StatisticsService server.Config     `mapstructure:"statistics_service"`
	GuestSecret       string            `mapstructure:"guest_secret"`
}
