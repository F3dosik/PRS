package server

import (
	"fmt"
	"log"

	"github.com/F3dosik/PRS.git/pkg/logger"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Port    string `env:"APP_PORT"`
	LogMode string `env:"LOG_MODE"`

	DatabaseURL string `env:"DATABASE_URL"`
}

const (
	defaultPort    = ":8080"
	defaultLogMode = string(logger.ModeDevelopment)
)

func (c *ServerConfig) Validate() error {
	if c.Port == "" {
		c.Port = defaultPort
	}

	if c.Port[0] != ':' {
		c.Port = ":" + c.Port
	}

	switch c.LogMode {
	case string(logger.ModeDevelopment), string(logger.ModeProduction):
	default:
		c.LogMode = defaultLogMode
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL can not be empty")
	}

	return nil
}

func LoadServerConfig() (*ServerConfig, error) {
	config := parseEnvConfig()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func parseEnvConfig() *ServerConfig {
	var config ServerConfig
	err := env.Parse(&config)
	if err != nil {
		log.Printf("Warning: failed to parse env config: %v\n", err)
	}

	return &config
}
