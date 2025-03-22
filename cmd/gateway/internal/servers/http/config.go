package http

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Server struct {
		Host            string        `yaml:"host" json:"host"`
		Port            int           `yaml:"port" json:"port"`
		Mode            string        `yaml:"mode" json:"mode"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
		Cors            struct {
			AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
			AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods"`
			AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers"`
			AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
			MaxAge           int      `yaml:"max_age" json:"max_age"`
		} `yaml:"cors" json:"cors"`
	} `yaml:"server" json:"server"`

	Services struct {
		Quiz     string `yaml:"quiz" json:"quiz"`
		Question string `yaml:"question" json:"question"`
	} `yaml:"services" json:"services"`

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
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
