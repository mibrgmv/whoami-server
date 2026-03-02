package config

import (
	"whoami-server/libs/keycloak"
	"whoami-server/libs/server"
)

type Config struct {
	Keycloak       keycloak.Config   `mapstructure:"keycloak"`
	HTTP           server.HTTPConfig `mapstructure:"http"`
	Metrics        server.Config     `mapstructure:"metrics"`
	AuthService    server.Config     `mapstructure:"auth_service"`
	QuizService    server.Config     `mapstructure:"quiz_service"`
	UserService    server.Config     `mapstructure:"user_service"`
	HistoryService server.Config     `mapstructure:"history_service"`
}
