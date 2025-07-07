package config

import (
	"whoami-server/internal/config/api/grpc"
	"whoami-server/internal/config/api/http"
	"whoami-server/internal/config/auth/jwt"
)

type Config struct {
	JWT            *jwt.Config  `mapstructure:"jwt"`
	HTTP           *http.Config `mapstructure:"http"`
	QuizzesService *grpc.Config `mapstructure:"quizzes_service"`
	UsersService   *grpc.Config `mapstructure:"users_service"`
	HistoryService *grpc.Config `mapstructure:"history_service"`
}
