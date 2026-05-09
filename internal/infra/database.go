/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	dbschema "entgo.io/ent/dialect/sql/schema"

	"github.com/origadmin/runtime/log"

	config "origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/migrate"
	mediadal "origadmin/application/origcms/internal/features/media/dal"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemdal "origadmin/application/origcms/internal/features/system/dal"
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
		if err := patchSQLiteSchema(dbSource, logger); err != nil {
			log.Warnf("SQLite pre-migration patch failed (non-fatal): %v", err)
		}
	}

	migrateOpts := []dbschema.MigrateOption{
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	}

	if dbDialect == "postgres" {
		migrateOpts = append(migrateOpts, migrate.WithForeignKeys(false))
		migrateOpts = append(migrateOpts, dbschema.WithApplyHook(func(next dbschema.Applier) dbschema.Applier {
			return dbschema.ApplyFunc(func(ctx context.Context, conn dialect.ExecQuerier, plan *atlasmigrate.Plan) error {
				for i, c := range plan.Changes {
					if strings.HasPrefix(c.Cmd, "CREATE INDEX ") || strings.HasPrefix(c.Cmd, "CREATE UNIQUE INDEX ") {
						c.Cmd = strings.Replace(c.Cmd, "CREATE INDEX ", "CREATE INDEX IF NOT EXISTS ", 1)
						c.Cmd = strings.Replace(c.Cmd, "CREATE UNIQUE INDEX ", "CREATE UNIQUE INDEX IF NOT EXISTS ", 1)
						plan.Changes[i] = c
					}
				}
				return next.Apply(ctx, conn, plan)
			})
		}))
	}

	if err := db.Schema.Create(ctx, migrateOpts...); err != nil {
		return nil, nil, err
	}

	if err := mediadal.SeedEncodeProfiles(ctx, db); err != nil {
		return nil, nil, err
	}

	if err := seedSettings(ctx, db); err != nil {
		return nil, nil, err
	}

	sqlDB, err := openSQLDB(dbSource, dbDialect)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open raw SQL connection: %w", err)
	}

	return db, sqlDB, nil
}

// patchSQLiteSchema adds missing columns to existing SQLite tables before Ent auto-migration.
// This prevents NOT NULL constraint violations when Ent copies data from old tables to new tables.
func patchSQLiteSchema(dsn string, logger log.Logger) error {
	// Extract file path from DSN (remove query parameters if present)
	dbPath := dsn
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dbPath = dsn[:idx]
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite for patching: %w", err)
	}
	defer db.Close()

	// Define columns that need to be added to existing tables.
	// These correspond to fields added in recent schema changes that old databases lack.
	// For DATETIME columns that need non-constant defaults (e.g. CURRENT_TIMESTAMP),
	// we add them as nullable first, then UPDATE existing rows, because SQLite
	// ALTER TABLE does not support non-constant defaults.
	patches := []struct {
		table      string
		column     string
		colType    string
		defaultVal string // constant default for ALTER TABLE; empty = nullable
		updateVal  string // value to UPDATE existing rows; empty = skip UPDATE
	}{
		// users table: new fields added in schema evolution
		{"users", "slug", "TEXT", "", ""},
		{"users", "status", "TEXT", "'ACTIVE'", ""},
		{"users", "nickname", "TEXT", "", ""},
		{"users", "phone", "TEXT", "", ""},
		{"users", "avatar", "TEXT", "", ""},
		{"users", "last_login_ip", "TEXT", "", ""},
		{"users", "login_ip", "TEXT", "", ""},
		{"users", "last_login_time", "DATETIME", "", ""},
		{"users", "login_time", "DATETIME", "", ""},
		// created_at/updated_at: add as nullable, then set to date_joined/date_added for existing rows
		{"users", "create_time", "DATETIME", "", "date_joined"},
		{"users", "update_time", "DATETIME", "", "date_added"},
		{"users", "create_author", "TEXT", "", ""},
		{"users", "update_author", "TEXT", "", ""},
		// content_media table: new fields
		{"content_media", "share_count", "INTEGER", "0", ""},
		{"content_media", "uuid", "TEXT", "", ""},
		{"content_media", "sprite_status", "TEXT", "'pending'", ""},
		{"content_media", "sprite_path", "TEXT", "", ""},
		{"content_media", "vtt_path", "TEXT", "", ""},
		{"content_media", "thumbnail_time", "REAL", "", ""},
		{"content_media", "tags", "JSON", "", ""},
		{"content_media", "create_author", "TEXT", "", ""},
		{"content_media", "update_author", "TEXT", "", ""},
	}

	for _, p := range patches {
		// Check if column already exists
		var count int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?",
			p.table, p.column,
		).Scan(&count)
		if err != nil {
			// Table might not exist yet (fresh database), skip
			continue
		}
		if count > 0 {
			continue // Column already exists
		}

		// Add the missing column (nullable, with constant default if provided)
		var alterSQL string
		if p.defaultVal != "" {
			alterSQL = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s DEFAULT %s", p.table, p.column, p.colType, p.defaultVal)
		} else {
			alterSQL = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", p.table, p.column, p.colType)
		}
		if _, err := db.Exec(alterSQL); err != nil {
			log.Warnf("SQLite patch: failed to add %s.%s: %v", p.table, p.column, err)
			continue
		}
		log.Infof("SQLite patch: added column %s.%s", p.table, p.column)

		// For columns that need non-constant defaults, UPDATE existing rows
		if p.updateVal != "" {
			updateSQL := fmt.Sprintf("UPDATE `%s` SET `%s` = `%s` WHERE `%s` IS NULL", p.table, p.column, p.updateVal, p.column)
			if result, err := db.Exec(updateSQL); err != nil {
				log.Warnf("SQLite patch: failed to update %s.%s: %v", p.table, p.column, err)
			} else if rows, _ := result.RowsAffected(); rows > 0 {
				log.Infof("SQLite patch: updated %d rows in %s.%s", rows, p.table, p.column)
			}
		}
	}

	return nil
}

// seedSettings seeds default settings into the database.
func seedSettings(ctx context.Context, db *entity.Client) error {
	repo := systemdal.NewSettingRepo(db)
	uc := systembiz.NewSettingUseCase(repo)
	return uc.SeedDefaults(ctx)
}

// openDB opens a database connection based on the DSN and database type.
func openDB(dsn, dbType string, logger log.Logger) (*entity.Client, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		// Ensure database exists before connecting
		if err := ensurePostgresDB(dsn, logger); err != nil {
			return nil, err
		}
		// Add sslmode if not present
		if !strings.Contains(dsn, "sslmode") {
			if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
				// URI format: append as query param
				if strings.Contains(dsn, "?") {
					dsn = dsn + "&sslmode=disable"
				} else {
					dsn = dsn + "?sslmode=disable"
				}
			} else {
				// key=value format: append as param
				dsn = dsn + " sslmode=disable"
			}
		}
	} else {
		// SQLite3: ensure the parent directory for the database file exists
		if err := ensureSQLiteDir(dsn); err != nil {
			return nil, fmt.Errorf("failed to create sqlite data directory: %w", err)
		}
		// Enable foreign keys pragma if not already set
		if !strings.Contains(dsn, "_fk=") {
			if strings.Contains(dsn, "?") {
				dsn = dsn + "&_fk=1"
			} else {
				dsn = dsn + "?_fk=1"
			}
		}
	}
	return entity.Open(driverName, dsn)
}

// ensureSQLiteDir ensures the parent directory for the SQLite database file exists.
func ensureSQLiteDir(dsn string) error {
	// Extract file path from DSN (remove query parameters if present)
	dbPath := dsn
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dbPath = dsn[:idx]
	}
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// ensurePostgresDB ensures the PostgreSQL database exists, creating it if necessary.
func ensurePostgresDB(dsn string, logger log.Logger) error {
	// Parse DSN to extract connection info
	_, dbName := parsePostgresDSN(dsn)
	if dbName == "" {
		return nil
	}

	// Build a DSN pointing to the default 'postgres' database
	var defaultDSN string
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// URI format: replace the database name in the path
		defaultDSN = replaceDBNameInURI(dsn, "postgres")
	} else {
		// key=value format: append/override dbname
		defaultDSN = dsn + " dbname=postgres sslmode=disable"
	}

	db, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).
		Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("create database %s: %w", dbName, err)
		}
		log.Infof("Created database: %s", dbName)
	}
	return nil
}

// replaceDBNameInURI replaces the database name in a PostgreSQL URI DSN.
func replaceDBNameInURI(dsn, newDBName string) string {
	scheme := "postgres://"
	if strings.HasPrefix(dsn, "postgresql://") {
		scheme = "postgresql://"
	}
	rest := dsn[len(scheme):]

	slashIdx := strings.Index(rest, "/")
	if slashIdx < 0 {
		// No path, append /newDBName
		return dsn + "/" + newDBName
	}

	authority := rest[:slashIdx]
	remainder := rest[slashIdx+1:]

	// Separate path from query
	qIdx := strings.Index(remainder, "?")
	var query string
	if qIdx >= 0 {
		query = "?" + remainder[qIdx+1:]
		remainder = remainder[:qIdx]
	}

	return scheme + authority + "/" + newDBName + query
}

// parsePostgresDSN parses a PostgreSQL DSN to extract connection string and database name.
func parsePostgresDSN(dsn string) (connStr, dbName string) {
	// URI format: postgres://user:pass@host/db?opts
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return parsePostgresURIDSN(dsn)
	}
	// key=value format: host=x dbname=y
	return parsePostgresKVDSN(dsn)
}

// parsePostgresURIDSN parses a PostgreSQL URI-format DSN.
func parsePostgresURIDSN(dsn string) (connStr, dbName string) {
	// Remove scheme
	rest := dsn
	if idx := strings.Index(rest, "://"); idx >= 0 {
		rest = rest[idx+3:]
	}

	// Split authority and path
	var _, pathPart, queryPart string
	if slashIdx := strings.Index(rest, "/"); slashIdx >= 0 {
		remainder := rest[slashIdx+1:]
		if qIdx := strings.Index(remainder, "?"); qIdx >= 0 {
			pathPart = remainder[:qIdx]
			queryPart = remainder[qIdx+1:]
		} else {
			pathPart = remainder
		}
	}

	dbName = pathPart

	// Rebuild connection string pointing to 'postgres' default database
	connStr = dsn
	// Replace dbname in URI
	if dbName != "" {
		if queryPart != "" {
			connStr = strings.Replace(connStr, "/"+dbName+"?", "/postgres?", 1)
		} else {
			connStr = strings.Replace(connStr, "/"+dbName, "/postgres", 1)
		}
	}
	return connStr, dbName
}

// parsePostgresKVDSN parses a PostgreSQL key=value format DSN.
func parsePostgresKVDSN(dsn string) (connStr, dbName string) {
	// Find dbname
	if i := strings.Index(dsn, "dbname="); i >= 0 {
		start := i + 7
		end := strings.IndexAny(dsn[start:], " ")
		if end < 0 {
			dbName = dsn[start:]
		} else {
			dbName = dsn[start : start+end]
		}
	}

	// Extract connection params for default DB (remove dbname)
	connParts := []string{}
	for _, part := range strings.Split(dsn, " ") {
		if strings.HasPrefix(part, "dbname=") {
			continue
		}
		connParts = append(connParts, part)
	}
	connStr = strings.Join(connParts, " ")
	return connStr, dbName
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

func openSQLDB(dsn, dbType string) (*sql.DB, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		if !strings.Contains(dsn, "sslmode") {
			if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
				if strings.Contains(dsn, "?") {
					dsn = dsn + "&sslmode=disable"
				} else {
					dsn = dsn + "?sslmode=disable"
				}
			} else {
				dsn = dsn + " sslmode=disable"
			}
		}
	} else {
		if !strings.Contains(dsn, "_fk=") {
			if strings.Contains(dsn, "?") {
				dsn = dsn + "&_fk=1"
			} else {
				dsn = dsn + "?_fk=1"
			}
		}
	}
	return sql.Open(driverName, dsn)
}
