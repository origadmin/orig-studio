/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/runtime/log"
)

// EnsureSQLiteDir ensures the parent directory for the SQLite database file exists.
func EnsureSQLiteDir(dsn string) error {
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

// PatchSQLiteSchema adds missing columns to existing SQLite tables before Ent auto-migration.
// This prevents NOT NULL constraint violations when Ent copies data from old tables to new tables.
func PatchSQLiteSchema(dsn string, logger log.Logger) error {
	dbPath := dsn
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dbPath = dsn[:idx]
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite for patching: %w", err)
	}
	defer db.Close()

	patches := []struct {
		table      string
		column     string
		colType    string
		defaultVal string
		updateVal  string
	}{
		{"users", "slug", "TEXT", "", ""},
		{"users", "status", "TEXT", "'ACTIVE'", ""},
		{"users", "nickname", "TEXT", "", ""},
		{"users", "phone", "TEXT", "", ""},
		{"users", "avatar", "TEXT", "", ""},
		{"users", "last_login_ip", "TEXT", "", ""},
		{"users", "login_ip", "TEXT", "", ""},
		{"users", "last_login_time", "DATETIME", "", ""},
		{"users", "login_time", "DATETIME", "", ""},
		{"users", "create_time", "DATETIME", "", "date_joined"},
		{"users", "update_time", "DATETIME", "", "date_added"},
		{"users", "create_author", "TEXT", "", ""},
		{"users", "update_author", "TEXT", "", ""},
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
		var count int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?",
			p.table, p.column,
		).Scan(&count)
		if err != nil {
			continue
		}
		if count > 0 {
			continue
		}

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

// PrepareSQLiteDSN adjusts the SQLite DSN by ensuring foreign keys are enabled.
func PrepareSQLiteDSN(dsn string) string {
	if !strings.Contains(dsn, "_fk=") {
		if strings.Contains(dsn, "?") {
			return dsn + "&_fk=1"
		}
		return dsn + "?_fk=1"
	}
	return dsn
}
