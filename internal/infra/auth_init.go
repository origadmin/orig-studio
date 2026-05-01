/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"time"

	config "origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/infra/auth"
)

// NewJWTManager creates a new JWT manager.
func NewJWTManager(cfg *config.Config) *auth.Manager {
	jwtSigningKey, _, jwtTTL, refreshTokenTTL := cfg.GetJWTConfig()
	jwtExpire := config.ParseDuration(jwtTTL, 3600*time.Second)
	refreshTokenExpire := config.ParseDuration(refreshTokenTTL, 720*time.Hour)
	return auth.NewManager(
		jwtSigningKey,
		jwtExpire,
		refreshTokenExpire,
	)
}
