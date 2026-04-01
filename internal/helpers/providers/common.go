/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package providers

import (
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/wire"

	"github.com/origadmin/runtime"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"github.com/origadmin/runtime/helpers/comp"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"
	"origadmin/application/origcms/internal/auth"
)

const (
	// CategoryPublisher is the runtime container category for the Watermill publisher.
	CategoryPublisher = "publisher"
)

const (
	// DefaultJWTSecret is the default JWT secret for development.
	DefaultJWTSecret = "origcms-secret-key-change-in-production"
	// DefaultJWTTL is the default JWT token TTL (24 hours).
	DefaultJWTTL = 24 * time.Hour
)

// CommonSet provides common dependencies for all modules.
var CommonSet = wire.NewSet(
	ProvideLogger,
	ProvideHasher,
	ProvidePublisher,
	ProvideServers,
	ProvideJWTManager,
)

// ProvideLogger provides the project's runtime logger.
func ProvideLogger(app *runtime.App) log.Logger {
	return app.Logger()
}

// ProvideHasher provides a hasher instance.
func ProvideHasher() (hash.Crypto, error) {
	return hash.NewCrypto(hashtypes.BCRYPT)
}

// ProvidePublisher provides the Watermill publisher from the runtime container.
// Returns nil if not configured (graceful degradation for M1 - events won't be published).
func ProvidePublisher(app *runtime.App) (message.Publisher, error) {
	pub, err := comp.Get[message.Publisher](app.Context(), app.Container().In(CategoryPublisher))
	if err != nil {
		// M1: publisher may not be configured; return nil for graceful degradation
		return nil, nil
	}
	return pub, nil
}

// ProvideServers extracts server transport configuration from the runtime config.
func ProvideServers(app *runtime.App) *transportv1.Servers {
	type hasServers interface {
		GetServers() *transportv1.Servers
	}
	if cfg, ok := app.Config().(hasServers); ok {
		return cfg.GetServers()
	}
	return nil
}

// ProvideJWTManager provides a JWT manager instance with default config.
func ProvideJWTManager() *auth.Manager {
	return auth.NewManager(DefaultJWTSecret, DefaultJWTTL)
}
