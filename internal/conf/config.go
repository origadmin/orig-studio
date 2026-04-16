/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package conf implements the functions, types, and contracts for the application.
package conf

import (
	"fmt"
	"time"

	"github.com/origadmin/runtime/config"
	"github.com/origadmin/runtime/engine/bootstrap"
)

const (
	// APIPrefix is the prefix for all API routes.
	APIPrefix = "/api/v1"
)

// Config holds all runtime configuration parsed from bootstrap.yaml.
type Config struct {
	Data struct {
		Databases map[string]struct {
			Name    string `yaml:"name"`
			Dialect string `yaml:"dialect"`
			Source  string `yaml:"source"`
		} `yaml:"databases"`
	} `yaml:"data"`
	Server struct {
		HTTP struct {
			Network string `yaml:"network"`
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"http"`
	} `yaml:"server"`
	Security struct {
		Authn struct {
			Configs []struct {
				Type string `yaml:"type"`
				JWT  struct {
					SigningKey      string `yaml:"signing_key"`
					SigningMethod   string `yaml:"signing_method"`
					AccessTokenTTL  string `yaml:"access_token_ttl"`
					RefreshTokenTTL string `yaml:"refresh_token_ttl"`
				} `yaml:"jwt"`
			} `yaml:"configs"`
		} `yaml:"authn"`
	} `yaml:"security"`
}

// GetDefaultDB returns the "default" database config for convenience.
func (c *Config) GetDefaultDB() (dialect, source string) {
	if c.Data.Databases != nil {
		if db, ok := c.Data.Databases["default"]; ok {
			return db.Dialect, db.Source
		}
	}
	return "postgres", ""
}

// GetJWTConfig returns the first JWT authn config found.
func (c *Config) GetJWTConfig() (signingKey, signingMethod, accessTokenTTL, refreshTokenTTL string) {
	for _, cfg := range c.Security.Authn.Configs {
		if cfg.Type == "jwt" {
			return cfg.JWT.SigningKey, cfg.JWT.SigningMethod, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL
		}
	}
	return "change-me-in-production", "HS256", "3600s", "720h"
}

// ParseDuration parses a duration string or returns the fallback.
func ParseDuration(s string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}

// Transformer implements bootstrap.ConfigTransformer for orig-cms Config.
var Transformer bootstrap.ConfigTransformer = bootstrap.ConfigTransformFunc(transformer)

func transformer(cfg config.KConfig) (any, error) {
	var c Config
	if err := cfg.Scan(&c); err != nil {
		return nil, fmt.Errorf("failed to scan config: %w", err)
	}
	return &c, nil
}
