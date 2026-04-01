// Package conf provides configuration loading
package conf

import (
	"fmt"
	"os"

	runtimeconfig "github.com/origadmin/runtime/config"
	runtimebootstrap "github.com/origadmin/runtime/engine/bootstrap"
)

// Transformer implements bootstrap.ConfigTransformer for orig-cms Config.
var Transformer runtimebootstrap.ConfigTransformer = runtimebootstrap.ConfigTransformFunc(transformer)

func transformer(cfg runtimeconfig.KConfig) (any, error) {
	var c Config
	if err := cfg.Scan(&c); err != nil {
		return nil, fmt.Errorf("failed to scan config: %w", err)
	}
	return &c, nil
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Driver string
	Source string
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "9090"),
		},
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "postgres"),
			Source: getEnv("DB_SOURCE", "host=localhost user=postgres password=postgres dbname=origcms port=5432 sslmode=disable"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

