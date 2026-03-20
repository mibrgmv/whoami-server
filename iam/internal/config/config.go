package config

import (
	"gordle/libs/keycloak"
	"gordle/libs/server"
)

type RecaptchaConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	VerifyURL string `mapstructure:"verify_url"`
}

type Config struct {
	Grpc        server.Config   `mapstructure:"grpc"`
	Keycloak    keycloak.Config `mapstructure:"keycloak"`
	GuestSecret string          `mapstructure:"guest_secret"`
	SmtpHost    string          `mapstructure:"smtp_host"`
	Recaptcha   RecaptchaConfig `mapstructure:"recaptcha"`
}
