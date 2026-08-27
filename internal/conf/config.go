/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package conf implements the functions, types, and contracts for the application.
package conf

import (
	"fmt"
	"os"
	"strconv"
	"time"

	discoveryv1 "github.com/origadmin/runtime/api/gen/go/config/discovery/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"github.com/origadmin/runtime/config"
	"github.com/origadmin/runtime/engine/bootstrap"
)

const (
	// APIPrefix is the prefix for all API routes.
	APIPrefix = "/api/v1"
)

// Config holds all runtime configuration parsed from bootstrap.yaml.
//
// This is the single source of truth for configuration in the application.
// It replaces the former confpb.Bootstrap (which was a parallel structure
// using json tags and could not parse security.authn correctly).
//
// The runtime framework registers it via conf.Transformer, and rt.Config()
// returns *Config. All wire.go providers should use *Config, not confpb.Bootstrap.
type Config struct {
	// Servers holds proto-style server configs (http, grpc).
	// YAML path: servers.configs[]
	// Uses json tag because Kratos Scan internally uses JSON unmarshal,
	// and transportv1.Servers (proto-generated) only has json tags.
	Servers *transportv1.Servers `json:"servers,omitempty"`
	// Clients holds proto-style gRPC client configs for the gateway.
	// YAML path: clients.configs[]
	Clients *transportv1.Clients `json:"clients,omitempty"`
	// Data holds the database configuration in a simplified shape.
	// YAML path: data.databases.{name}
	Data struct {
		Databases map[string]struct {
			Name    string `json:"name"`
			Dialect string `json:"dialect"`
			Source  string `json:"source"`
		} `json:"databases"`
	} `json:"data"`
	// Server holds the CE single-instance server config (cmd/server only).
	// EE microservices use Servers (proto-style) instead.
	Server struct {
		HTTP struct {
			Network string `json:"network"`
			Addr    string `json:"addr"`
			Timeout string `json:"timeout"`
		} `json:"http"`
		GRPC struct {
			Network string `json:"network"`
			Addr    string `json:"addr"`
			Timeout string `json:"timeout"`
		} `json:"grpc"`
	} `json:"server"`
	// Security holds JWT and other authn configs.
	// YAML path: security.authn.configs[]
	Security struct {
		Authn struct {
			Configs []struct {
				Type string `json:"type"`
				JWT  struct {
					SigningKey      string `json:"signing_key"`
					SigningMethod   string `json:"signing_method"`
					AccessTokenTTL  string `json:"access_token_ttl"`
					RefreshTokenTTL string `json:"refresh_token_ttl"`
				} `json:"jwt"`
			} `json:"configs"`
		} `json:"authn"`
	} `json:"security"`
	Asynq *AsynqConfig `json:"asynq,omitempty"`
	// Discovery configuration for service registration and discovery
	Discovery *discoveryv1.Discoveries `json:"discovery,omitempty"`
	// GRPCClients configuration for gRPC client connections
	GRPCClients map[string]*GRPCClientConfig `json:"grpc_clients,omitempty"`
}

// GetServers returns the proto-style servers config.
func (c *Config) GetServers() *transportv1.Servers {
	if c == nil {
		return nil
	}
	return c.Servers
}

// GetClients returns the proto-style clients config (used by gateway).
func (c *Config) GetClients() *transportv1.Clients {
	if c == nil {
		return nil
	}
	return c.Clients
}

// GRPCClientConfig holds configuration for a gRPC client connection.
type GRPCClientConfig struct {
	// Endpoint is the service endpoint to connect to, can be direct address or discovery URI
	Endpoint string `json:"endpoint"`
	// Timeout is the default request timeout
	Timeout string `json:"timeout"`
	// DiscoveryName is the name of the discovery client to use
	DiscoveryName string `json:"discovery_name,omitempty"`
}

// AsynqConfig holds asynq distributed task queue configuration.
type AsynqConfig struct {
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	Concurrency   int32  `json:"concurrency"`
}

func (c *Config) GetAsynq() *AsynqConfig {
	if c.Asynq == nil {
		return nil
	}
	return c.Asynq
}

func (a *AsynqConfig) GetRedisAddr() string {
	if v := os.Getenv("ASYNQ_REDIS_ADDR"); v != "" {
		return v
	}
	return a.RedisAddr
}
func (a *AsynqConfig) GetRedisPassword() string {
	if v := os.Getenv("ASYNQ_REDIS_PASSWORD"); v != "" {
		return v
	}
	return a.RedisPassword
}
func (a *AsynqConfig) GetConcurrency() int32 {
	if v := os.Getenv("ASYNQ_CONCURRENCY"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			return int32(n)
		}
	}
	return a.Concurrency
}

// GetDefaultDB returns the "default" database config for convenience.
func (c *Config) GetDefaultDB() (dialect, source string) {
	if c.Data.Databases != nil {
		if db, ok := c.Data.Databases["default"]; ok {
			dialect = db.Dialect
			source = db.Source
		}
	}
	if v := os.Getenv("DATABASE_DIALECT"); v != "" {
		dialect = v
	}
	if v := os.Getenv("DATABASE_SOURCE"); v != "" {
		source = v
	}
	if dialect == "" {
		dialect = "sqlite3"
	}
	return dialect, source
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
