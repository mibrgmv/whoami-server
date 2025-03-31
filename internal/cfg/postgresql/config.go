package postgresql

import "fmt"

type Config struct {
	Host     string `mapstructure:"host"`
	Port     uint16 `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SslMode  string `mapstructure:"ssl_mode"`
}

func (c *Config) GetConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SslMode,
	)
}
