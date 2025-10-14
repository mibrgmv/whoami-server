package config

import (
	"github.com/mibrgmv/whoami-server/shared/grpc"
	"github.com/mibrgmv/whoami-server/shared/http"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
)

type Config struct {
	Keycloak       keycloak.Config `mapstructure:"keycloak"`
	HTTP           http.Config     `mapstructure:"http"`
	AuthService    grpc.Config     `mapstructure:"auth_service"`
	QuizService    grpc.Config     `mapstructure:"quiz_service"`
	UserService    grpc.Config     `mapstructure:"user_service"`
	HistoryService grpc.Config     `mapstructure:"history_service"`
}
