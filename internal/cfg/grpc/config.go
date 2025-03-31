package grpc

import "fmt"

type Config struct {
	Host string `mapstructure:"host"`
	Port uint16 `mapstructure:"port"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
