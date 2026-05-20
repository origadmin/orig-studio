/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"origadmin/application/origstudio/internal/conf"
)

// AdminConfig holds the runtime configuration needed by AdminHandler
// that cannot be auto-injected by wire (multiple string parameters).
type AdminConfig struct {
	AppVersion string
	DBDialect  string
}

// NewAdminConfig creates an AdminConfig from the application Config.
func NewAdminConfig(cfg *conf.Config) *AdminConfig {
	dbDialect, _ := cfg.GetDefaultDB()
	return &AdminConfig{
		AppVersion: "", // Set by wire bridge from Version var
		DBDialect:  dbDialect,
	}
}
