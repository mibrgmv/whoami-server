package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server struct {
		Host string `yaml:"host" json:"host"`
		Port int    `yaml:"port" json:"port"`
		Mode string `yaml:"mode" json:"mode"`
		Cors struct {
			AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
			AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods"`
			AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers"`
			AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
			MaxAge           int      `yaml:"max_age" json:"max_age"`
		} `yaml:"cors" json:"cors"`
	} `yaml:"server" json:"server"`

	Postgres struct {
		Host     string `yaml:"host" json:"host"`
		Port     int    `yaml:"port" json:"port"`
		Database string `yaml:"database" json:"database"`
		Username string `yaml:"username" json:"username"`
		Password string `yaml:"password" json:"password"`
		SslMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	} `yaml:"postgres" json:"postgres"`

	Swagger struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		BasePath    string `yaml:"base_path" json:"base_path"`
		Title       string `yaml:"title" json:"title"`
		Description string `yaml:"description" json:"description"`
		Version     string `yaml:"version" json:"version"`
	} `yaml:"swagger" json:"swagger"`
}

func (c *Config) Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Postgres.Username,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database,
		c.Postgres.SslMode,
	)
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
