/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package database

import (
	"github.com/origadmin/runtime/log"

	config "origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/entity"
)

// NewDatabaseBridge wraps NewDatabase to return a cleanup function
// instead of *sql.DB, which is required by wire's provider signature convention.
func NewDatabaseBridge(cfg *config.Config, logger log.Logger) (*entity.Client, error) {
	client, _, err := NewDatabase(cfg, logger)
	if err != nil {
		return nil, err
	}
	return client, nil
}
