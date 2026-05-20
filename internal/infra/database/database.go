/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	dbschema "entgo.io/ent/dialect/sql/schema"

	"github.com/origadmin/runtime/log"

	config "origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/migrate"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	contentdal "origadmin/application/origstudio/internal/features/content/dal"
	mediadal "origadmin/application/origstudio/internal/features/media/dal"
	systemdal "origadmin/application/origstudio/internal/features/system/dal"
)

// NewDatabase creates a new database client.
func NewDatabase(cfg *config.Config, logger log.Logger) (*entity.Client, *sql.DB, error) {
	dbDialect, dbSource := cfg.GetDefaultDB()
	db, err := openDB(dbSource, dbDialect, logger)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	if dbDialect == "sqlite3" {
		if err := PatchSQLiteSchema(dbSource, logger); err != nil {
			log.Warnf("SQLite pre-migration patch failed (non-fatal): %v", err)
		}
	}

	migrateOpts := []dbschema.MigrateOption{
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	}

	if dbDialect == "postgres" {
		migrateOpts = append(migrateOpts, PostgresMigrateOptions()...)
	}

	if err := db.Schema.Create(ctx, migrateOpts...); err != nil {
		return nil, nil, err
	}

	if err := mediadal.SeedEncodeProfiles(ctx, db); err != nil {
		return nil, nil, err
	}

	if err := systemdal.SeedSettings(ctx, db); err != nil {
		return nil, nil, err
	}

	if err := contentdal.SeedCategories(ctx, db); err != nil {
		return nil, nil, err
	}

	if err := contentdal.SeedTags(ctx, db); err != nil {
		return nil, nil, err
	}

	sqlDB, err := openSQLDB(dbSource, dbDialect)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open raw SQL connection: %w", err)
	}

	// Register the entity-layer IsNotFound checker so that domain/types.IsNotFound()
	// works without biz/service importing internal/dal/entity directly.
	repotypes.RegisterNotFoundChecker(entity.IsNotFound)

	return db, sqlDB, nil
}

// EnvInt reads an environment variable as an integer, returning the default if not set or invalid.
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

// openDB opens a database connection based on the DSN and database type.
func openDB(dsn, dbType string, logger log.Logger) (*entity.Client, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		if err := EnsurePostgresDB(dsn, logger); err != nil {
			return nil, err
		}
		dsn = PreparePostgresDSN(dsn)
	} else {
		if err := EnsureSQLiteDir(dsn); err != nil {
			return nil, fmt.Errorf("failed to create sqlite data directory: %w", err)
		}
		dsn = PrepareSQLiteDSN(dsn)
	}
	return entity.Open(driverName, dsn)
}

// openSQLDB opens a raw sql.DB connection for health checks and similar purposes.
func openSQLDB(dsn, dbType string) (*sql.DB, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		dsn = PreparePostgresDSN(dsn)
	} else {
		dsn = PrepareSQLiteDSN(dsn)
	}
	return sql.Open(driverName, dsn)
}
