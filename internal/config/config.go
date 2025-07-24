package config

import (
	"github.com/mibrgmv/whoami-server-shared/config/api/grpc"
	"github.com/mibrgmv/whoami-server-shared/config/api/http"
	"whoami-server-gateway/internal/auth/keycloak"
)

type Config struct {
	Keycloak       keycloak.Config `json:"keycloak"`
	HTTP           http.Config     `json:"http"`
	QuizzesService grpc.Config     `json:"quizzes_service"`
	UsersService   grpc.Config     `json:"users_service"`
	HistoryService grpc.Config     `json:"history_service"`
}
