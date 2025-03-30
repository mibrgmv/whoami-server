package grpc

type Config struct {
	Host string `mapstructure:"host"`
	Port uint16 `mapstructure:"port"`
}
