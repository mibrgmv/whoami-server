package config

import (
	"whoami-server/libs/keycloak"
	"whoami-server/libs/server"
)

type Config struct {
	Grpc     server.Config   `mapstructure:"grpc"`
	Keycloak keycloak.Config `mapstructure:"keycloak"`
}
