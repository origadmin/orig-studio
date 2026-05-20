/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	dbschema "entgo.io/ent/dialect/sql/schema"
	"origadmin/application/origstudio/internal/dal/entity/migrate"

	"github.com/origadmin/runtime/log"
)

// EnsurePostgresDB ensures the PostgreSQL database exists, creating it if necessary.
func EnsurePostgresDB(dsn string, logger log.Logger) error {
	_, dbName := ParsePostgresDSN(dsn)
	if dbName == "" {
		return nil
	}

	var defaultDSN string
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		defaultDSN = replaceDBNameInURI(dsn, "postgres")
	} else {
		defaultDSN = dsn + " dbname=postgres sslmode=disable"
	}

	db, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return err
	}
	defer db.Close()

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

// PreparePostgresDSN adjusts the PostgreSQL DSN by ensuring sslmode is set.
func PreparePostgresDSN(dsn string) string {
	if !strings.Contains(dsn, "sslmode") {
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			if strings.Contains(dsn, "?") {
				return dsn + "&sslmode=disable"
			}
			return dsn + "?sslmode=disable"
		}
		return dsn + " sslmode=disable"
	}
	return dsn
}

// PostgresMigrateOptions returns migration options specific to PostgreSQL.
func PostgresMigrateOptions() []dbschema.MigrateOption {
	return []dbschema.MigrateOption{
		migrate.WithForeignKeys(false),
		dbschema.WithApplyHook(func(next dbschema.Applier) dbschema.Applier {
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
		}),
	}
}

// ParsePostgresDSN parses a PostgreSQL DSN to extract connection string and database name.
func ParsePostgresDSN(dsn string) (connStr, dbName string) {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return parsePostgresURIDSN(dsn)
	}
	return parsePostgresKVDSN(dsn)
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
		return dsn + "/" + newDBName
	}

	authority := rest[:slashIdx]
	remainder := rest[slashIdx+1:]

	qIdx := strings.Index(remainder, "?")
	var query string
	if qIdx >= 0 {
		query = "?" + remainder[qIdx+1:]
		remainder = remainder[:qIdx]
	}

	return scheme + authority + "/" + newDBName + query
}

// parsePostgresURIDSN parses a PostgreSQL URI-format DSN.
func parsePostgresURIDSN(dsn string) (connStr, dbName string) {
	rest := dsn
	if idx := strings.Index(rest, "://"); idx >= 0 {
		rest = rest[idx+3:]
	}

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

	connStr = dsn
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
	if i := strings.Index(dsn, "dbname="); i >= 0 {
		start := i + 7
		end := strings.IndexAny(dsn[start:], " ")
		if end < 0 {
			dbName = dsn[start:]
		} else {
			dbName = dsn[start : start+end]
		}
	}

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
