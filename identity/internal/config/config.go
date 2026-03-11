package config

import (
	"gordle/libs/keycloak"
	"gordle/libs/server"
)

type Config struct {
	Grpc     server.Config   `mapstructure:"grpc"`
	Keycloak keycloak.Config `mapstructure:"keycloak"`
}
