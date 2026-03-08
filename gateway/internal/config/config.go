package config

import (
	"whoami-server/libs/keycloak"
	"whoami-server/libs/server"
)

type Config struct {
	Keycloak          keycloak.Config   `mapstructure:"keycloak"`
	HTTP              server.HTTPConfig `mapstructure:"http"`
	Metrics           server.Config     `mapstructure:"metrics"`
	IdentityService   server.Config     `mapstructure:"identity_service"`
	GameService       server.Config     `mapstructure:"game_service"`
	StatisticsService server.Config     `mapstructure:"statistics_service"`
}
