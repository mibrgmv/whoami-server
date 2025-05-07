package http

import "fmt"

type Config struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ShutdownTimeout string `mapstructure:"shutdown_timeout"`
	CORS            CORS   `mapstructure:"cors"`
}

type CORS struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
