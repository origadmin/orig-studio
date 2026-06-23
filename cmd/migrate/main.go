package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"origadmin/application/origstudio/internal/migrate"
)

func main() {
	var (
		sourceType   string
		sourceDSN    string
		sourceMedia  string
		targetDSN    string
		targetDialect string
		targetMedia  string
		stateDir     string
		dryRun       bool
		overwrite    bool
		resume       bool
		showHelp     bool
	)

	flag.StringVar(&sourceType, "source-type", "mediacms", "Source platform type (mediacms)")
	flag.StringVar(&sourceDSN, "source-dsn", "", "Source database DSN")
	flag.StringVar(&sourceMedia, "source-media", "", "Source media files directory")
	flag.StringVar(&targetDSN, "target-dsn", "./data/origstudio.db?_fk=1", "Target database DSN")
	flag.StringVar(&targetDialect, "target-dialect", "sqlite3", "Target dialect (sqlite3|postgres)")
	flag.StringVar(&targetMedia, "target-media", "./data/uploads", "Target media directory")
	flag.StringVar(&stateDir, "state-dir", "./data/migration", "Migration state directory")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run mode (no writes)")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite existing files")
	flag.BoolVar(&resume, "resume", false, "Resume from last checkpoint")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Data Migration Tool for orig-cms-ee

Usage:
  migrate [options]

Options:
  -source-type string   Source platform type (default: mediacms)
  -source-dsn string    Source database connection string
  -source-media string  Source media files directory
  -target-dsn string    Target database DSN (default: ./data/origstudio.db?_fk=1)
  -target-dialect      Target dialect: sqlite3|postgres (default: sqlite3)
  -target-media string  Target media directory (default: ./data/uploads)
  -state-dir string     Migration state dir (default: ./data/migration)
  -dry-run              Dry run mode, no actual writes (default: false)
  -overwrite            Overwrite existing files (default: false)
  -resume               Resume from last checkpoint (default: false)
  -help                 Show this help

Examples:
  # Migrate from MediaCMS (PostgreSQL) to local SQLite
  migrate -source-dsn "postgres://user:pass@localhost/mediacms" \
          -source-media /var/lib/mediacms/media

  # Migrate from MediaCMS (SQLite) to PostgreSQL target
  migrate -source-dsn "/data/mediacms/db.sqlite3" \
          -target-dsn "postgres://user:pass@localhost/origcms" \
          -target-dialect postgres

  # Dry run to preview migration plan
  migrate -dry-run -source-dsn "..." -source-media "..."

  # Resume interrupted migration
  migrate -resume

Supported Source Platforms:
  - mediacms: MediaCMS (https://github.com/mediacms-io/mediacms)
`)
	}

	flag.Parse()

	if showHelp || sourceDSN == "" && !resume {
		flag.Usage()
		os.Exit(0)
	}

	logger := log.New(os.Stdout, "[migrate] ", log.LstdFlags|log.Lmicroseconds)
	reporter := migrate.NewConsoleReporter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Println("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	engine := migrate.NewEngine(nil, reporter, stateDir, logger)

	if resume {
		migrations, err := findLatestState(stateDir)
		if err != nil {
			logger.Fatalf("Cannot find migration state: %v", err)
		}
		if migrations == "" {
			logger.Fatal("No previous migration found. Cannot resume.")
		}
		logger.Printf("Resuming migration: %s", migrations)
		if err := engine.LoadState(migrations); err != nil {
			logger.Fatalf("Load state failed: %v", err)
		}
		if err := engine.Resume(ctx); err != nil {
			logger.Fatalf("Migration failed: %v", err)
		}
	} else {
		var adapter migrate.SourceAdapter
		switch sourceType {
		case "mediacms":
			adapter = migrate.NewMediaCMSAdapter()
		default:
			logger.Fatalf("Unknown source type: %s", sourceType)
		}
		engine.Source(adapter)

		sourceCfg := &migrate.SourceConfig{
			Type:     sourceType,
			DSN:      sourceDSN,
			MediaDir: sourceMedia,
		}
		targetCfg := &migrate.TargetConfig{
			DSN:       targetDSN,
			Dialect:   targetDialect,
			MediaDir:  targetMedia,
			Overwrite: overwrite,
			DryRun:    dryRun,
		}

		logger.Printf("Starting migration from %s...", sourceType)
		logger.Printf("Source: %s", sourceDSN)
		logger.Printf("Target: %s (%s)", targetDSN, targetDialect)
		if dryRun {
			logger.Println("*** DRY RUN MODE - no data will be written ***")
		}

		if err := engine.Run(ctx, sourceCfg, targetCfg); err != nil {
			logger.Fatalf("Migration failed: %v", err)
		}
	}

	reporter.Summary()
	logger.Println("Migration completed successfully!")
}

func findLatestState(stateDir string) (string, error) {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return "", err
	}
	var latest string
	var latestTime int64
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		statePath := filepath.Join(stateDir, entry.Name(), "state.json")
		info, err := os.Stat(statePath)
		if err != nil {
			continue
		}
		if info.ModTime().UnixNano() > latestTime {
			latestTime = info.ModTime().UnixNano()
			latest = entry.Name()
		}
	}
	return latest, nil
}
