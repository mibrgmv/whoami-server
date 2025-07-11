package internal

import (
	"whoami-server/cmd/gateway/internal/keycloak"
	"whoami-server/internal/config/api/grpc"
	"whoami-server/internal/config/api/http"
)

type Config struct {
	Keycloak       *keycloak.Config `mapstructure:"keycloak"`
	HTTP           *http.Config     `mapstructure:"http"`
	QuizzesService *grpc.Config     `mapstructure:"quizzes_service"`
	UsersService   *grpc.Config     `mapstructure:"users_service"`
	HistoryService *grpc.Config     `mapstructure:"history_service"`
}
