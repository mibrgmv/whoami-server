package config

import (
	"github.com/mibrgmv/whoami-server/shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
)

type Config struct {
	Grpc     grpc.Config     `mapstructure:"grpc"`
	Keycloak keycloak.Config `mapstructure:"keycloak"`
}
