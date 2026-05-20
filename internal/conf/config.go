/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package conf implements the functions, types, and contracts for the application.
package conf

import (
	"fmt"
	"time"

	discoveryv1 "github.com/origadmin/runtime/api/gen/go/config/discovery/v1"
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
		GRPC struct {
			Network string `yaml:"network"`
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"grpc"`
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
	Asynq *AsynqConfig `yaml:"asynq,omitempty"`
	// Discovery configuration for service registration and discovery
	Discovery *discoveryv1.Discoveries `yaml:"discovery,omitempty"`
	// GRPCClients configuration for gRPC client connections
	GRPCClients map[string]*GRPCClientConfig `yaml:"grpc_clients,omitempty"`
}

// GRPCClientConfig holds configuration for a gRPC client connection.
type GRPCClientConfig struct {
	// Endpoint is the service endpoint to connect to, can be direct address or discovery URI
	Endpoint string `yaml:"endpoint"`
	// Timeout is the default request timeout
	Timeout string `yaml:"timeout"`
	// DiscoveryName is the name of the discovery client to use
	DiscoveryName string `yaml:"discovery_name,omitempty"`
}

// AsynqConfig holds asynq distributed task queue configuration.
type AsynqConfig struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
	Concurrency   int32  `yaml:"concurrency"`
}

func (c *Config) GetAsynq() *AsynqConfig {
	if c.Asynq == nil {
		return nil
	}
	return c.Asynq
}

func (a *AsynqConfig) GetRedisAddr() string     { return a.RedisAddr }
func (a *AsynqConfig) GetRedisPassword() string  { return a.RedisPassword }
func (a *AsynqConfig) GetConcurrency() int32     { return a.Concurrency }

// GetDefaultDB returns the "default" database config for convenience.
func (c *Config) GetDefaultDB() (dialect, source string) {
	if c.Data.Databases != nil {
		if db, ok := c.Data.Databases["default"]; ok {
			return db.Dialect, db.Source
		}
	}
	return "sqlite3", ""
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

// Transformer implements bootstrap.ConfigTransformer for orig-studio Config.
var Transformer bootstrap.ConfigTransformer = bootstrap.ConfigTransformFunc(transformer)

func transformer(cfg config.KConfig) (any, error) {
	var c Config
	if err := cfg.Scan(&c); err != nil {
		return nil, fmt.Errorf("failed to scan config: %w", err)
	}
	return &c, nil
}
