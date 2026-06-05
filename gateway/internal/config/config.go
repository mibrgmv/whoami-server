package config

import (
	"time"

	"gordle/libs/keycloak"
	"gordle/libs/server"
)

type Config struct {
	Keycloak          keycloak.Config   `mapstructure:"keycloak"`
	HTTP              server.HTTPConfig `mapstructure:"http"`
	Metrics           server.Config     `mapstructure:"metrics"`
	GameService       server.Config     `mapstructure:"game_service"`
	GameWebSocket     server.Config     `mapstructure:"game_websocket"`
	StatisticsService server.Config     `mapstructure:"statistics_service"`
	GuestSecret       string            `mapstructure:"guest_secret"`
	Redis             struct {
		Address  string        `mapstructure:"address"`
		Password string        `mapstructure:"password"`
		DB       int           `mapstructure:"db"`
		TTL      time.Duration `mapstructure:"ttl"`
	} `mapstructure:"redis"`
	Session SessionConfig `mapstructure:"session"`
}

type SessionConfig struct {
	CookieName   string `mapstructure:"cookie_name"`
	CookieSecure bool   `mapstructure:"cookie_secure"`
	CallbackURL  string `mapstructure:"callback_url"`
}
