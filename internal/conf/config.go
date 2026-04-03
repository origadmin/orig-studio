// Package conf provides configuration loading
package conf

import (
	"fmt"

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

// Config is a placeholder for runtime-based config scanning.
// The actual config struct lives in cmd/server/main.go since the monolith
// entry point defines its own Config matching the bootstrap.yaml structure.
// This file exists to satisfy the Transformer registration used by the
// runtime engine.
type Config struct{}
