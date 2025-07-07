package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
	"strings"
)

func LoanConfig(cfg interface{}) error {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("failed to get called information")
	}

	mainDir := filepath.Dir(filename)
	configDir := filepath.Join(mainDir, "configs")

	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
