package config

import (
	"github.com/mibrgmv/whoami-server-shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server-shared/config/api/http"
	"whoami-server-gateway/internal/auth/keycloak"
)

type Config struct {
	Keycloak       keycloak.Config `mapstructure:"keycloak"`
	HTTP           http.Config     `mapstructure:"http"`
	QuizzesService grpc.Config     `mapstructure:"quizzes_service"`
	HistoryService grpc.Config     `mapstructure:"history_service"`
}
