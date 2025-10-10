package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadWithOptions(cfg interface{}, opts *Options) error {
	for _, envFile := range opts.EnvFiles {
		if err := loadEnvFile(envFile, opts.Silent); err != nil && !opts.IgnoreNotFound {
			return fmt.Errorf("failed to load env file %s: %w", envFile, err)
		}
	}

	viper.SetConfigName(opts.ConfigName)
	viper.SetConfigType(opts.ConfigType)

	for _, path := range opts.ConfigPaths {
		viper.AddConfigPath(path)
	}

	viper.AutomaticEnv()

	if opts.EnvPrefix != "" {
		viper.SetEnvPrefix(opts.EnvPrefix)
	}

	if opts.EnvKeyReplacer != nil {
		viper.SetEnvKeyReplacer(opts.EnvKeyReplacer)
	}

	if err := viper.ReadInConfig(); err != nil {
		if opts.IgnoreNotFound {
			if !opts.Silent {
				log.Printf("Config file not found, using environment variables only: %v", err)
			}
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else if !opts.Silent {
		log.Printf("Loaded config from: %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func loadEnvFile(filename string, silent bool) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if !silent {
			log.Printf("Environment file %s not found, skipping", filename)
		}
		return nil
	}

	if err := godotenv.Load(filename); err != nil {
		return err
	}

	if !silent {
		log.Printf("Loaded environment file: %s", filename)
	}
	return nil
}
