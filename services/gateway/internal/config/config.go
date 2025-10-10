package config

import (
	"github.com/mibrgmv/whoami-server/services/gateway/internal/auth/keycloak"
	"github.com/mibrgmv/whoami-server/shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server/shared/config/api/http"
)

type Config struct {
	Keycloak       keycloak.Config `mapstructure:"keycloak"`
	HTTP           http.Config     `mapstructure:"http"`
	QuizzesService grpc.Config     `mapstructure:"quizzes_service"`
	HistoryService grpc.Config     `mapstructure:"history_service"`
}
