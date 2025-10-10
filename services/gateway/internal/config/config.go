package config

import (
	"github.com/mibrgmv/whoami-server/shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server/shared/config/api/http"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
)

type Config struct {
	Keycloak       keycloak.Config `mapstructure:"keycloak"`
	HTTP           http.Config     `mapstructure:"http"`
	QuizzesService grpc.Config     `mapstructure:"quizzes_service"`
	UserService    grpc.Config     `mapstructure:"user_service"`
	HistoryService grpc.Config     `mapstructure:"history_service"`
}
