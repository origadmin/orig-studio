/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"database/sql"
	"os"
	"strconv"

	"github.com/origadmin/runtime/log"

	config "origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/infra/database"
)

// NewDatabase creates a new database client.
// Delegates to the database sub-package for all initialization logic.
func NewDatabase(cfg *config.Config, logger log.Logger) (*entity.Client, *sql.DB, error) {
	return database.NewDatabase(cfg, logger)
}

// EnvInt reads an environment variable as an integer, returning the default if not set or invalid.
// Kept here for backward compatibility with existing callers.
func EnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
