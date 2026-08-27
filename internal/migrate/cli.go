package migrate

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"origadmin/application/origstudio/internal/conf"
)

// RunCLI runs the migration command and returns a process exit code.
//
// Migration is an offline batch job binary (cmd/migrate → origcms-migrate),
// not a development-stage step and not a runtime service. EE is a microservice
// architecture with no `server` monolith; the migration writes across the
// user/media/content domains, so it ships as a standalone one-shot binary whose
// Docker entrypoint is /app/bin/origcms-migrate. This single entry point is
// shared by the standalone cmd/migrate tool (the deployment vehicle) and any
// future embedded use.
func RunCLI(args []string) int {
	var (
		sourceType    string
		sourceDSN     string
		sourceMedia   string
		sourceTypes   string
		targetDSN     string
		targetDialect string
		targetMedia   string
		stateDir      string
		dryRun        bool
		overwrite     bool
		resume        bool
		showHelp      bool
	)

	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.StringVar(&sourceType, "source-type", "mediacms", "Source platform type (see Supported Source Platforms below)")
	fs.StringVar(&sourceDSN, "source-dsn", "", "Source database DSN")
	fs.StringVar(&sourceMedia, "source-media", "", "Source media files directory")
	fs.StringVar(&sourceTypes, "media-types", "video", "Comma-separated source media type whitelist (default: video)")
	fs.StringVar(&targetDSN, "target-dsn", "./data/origstudio.db?_fk=1", "Target database DSN")
	fs.StringVar(&targetDialect, "target-dialect", "sqlite3", "Target dialect (sqlite3|postgres)")
	fs.StringVar(&targetMedia, "target-media", "./data/uploads", "Target media directory")
	fs.StringVar(&stateDir, "state-dir", "./data/migration", "Migration state directory")
	fs.BoolVar(&dryRun, "dry-run", false, "Dry run mode (no writes)")
	fs.BoolVar(&overwrite, "overwrite", false, "Overwrite existing files")
	fs.BoolVar(&resume, "resume", false, "Resume from last checkpoint")
	fs.BoolVar(&showHelp, "help", false, "Show help")

	sources := RegisteredSources()
	sourcesHelp := "Supported Source Platforms:\n"
	for _, s := range sources {
		sourcesHelp += "  - " + s + "\n"
	}

	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `Data Migration Tool for orig-cms-ee

Usage:
  origcms-migrate [options]   # standalone one-shot batch binary (deployment path)

Options:
  -source-type string   Source platform type (default: mediacms)
  -source-dsn string    Source database connection string
  -source-media string  Source media files directory
  -media-types string   Comma-separated source media type whitelist (default: video)
  -target-dsn string    Target database DSN (default: ./data/origstudio.db?_fk=1)
  -target-dialect      Target dialect: sqlite3|postgres (default: sqlite3)
  -target-media string  Target media directory (default: ./data/uploads)
  -state-dir string     Migration state dir (default: ./data/migration)
  -dry-run              Dry run mode, no actual writes (default: false)
  -overwrite            Overwrite existing files (default: false)
  -resume               Resume from last checkpoint (default: false)
  -help                 Show this help

Examples:
  # Migrate from MediaCMS (PostgreSQL) to local SQLite, video only
  migrate -source-dsn "postgres://user:pass@localhost/mediacms" \
          -source-media /var/lib/mediacms/media

  # Migrate only video and audio types
  migrate -source-dsn "..." -source-media "..." -media-types video,audio

  # Migrate from MediaCMS (SQLite) to PostgreSQL target
  migrate -source-dsn "/data/mediacms/db.sqlite3" \
          -target-dsn "postgres://user:pass@localhost/origcms" \
          -target-dialect postgres

  # Migrate from another platform (e.g. PeerTube)
  migrate -source-type peertube -source-dsn "postgres://user:pass@localhost/peertube" \
          -source-media /var/lib/peertube/storage

  # Dry run to preview migration plan
  migrate -dry-run -source-dsn "..." -source-media "..."

  # Resume interrupted migration
  migrate -resume

`+sourcesHelp)
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		fs.Usage()
		return 2
	}

	if showHelp || sourceDSN == "" && !resume {
		fs.Usage()
		return 0
	}

	logger := log.New(os.Stdout, "[migrate] ", log.LstdFlags|log.Lmicroseconds)
	reporter := NewConsoleReporter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Println("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	engine := NewEngine(nil, reporter, stateDir, logger)

	if resume {
		migrations, err := findLatestState(stateDir)
		if err != nil {
			logger.Printf("Cannot find migration state: %v", err)
			return 1
		}
		if migrations == "" {
			logger.Println("No previous migration found. Cannot resume.")
			return 1
		}
		logger.Printf("Resuming migration: %s", migrations)
		if err := engine.LoadState(migrations); err != nil {
			logger.Printf("Load state failed: %v", err)
			return 1
		}
		if err := engine.Resume(ctx); err != nil {
			logger.Printf("Migration failed: %v", err)
			return 1
		}
	} else {
		adapter, err := NewSourceAdapter(sourceType, conf.NewStoragePaths(conf.DefaultStorageConfig().BasePath))
		if err != nil {
			logger.Printf("%v", err)
			return 1
		}
		engine.Source(adapter)

		sourceCfg := &SourceConfig{
			Type:       sourceType,
			DSN:        sourceDSN,
			MediaDir:   sourceMedia,
			MediaTypes: splitMediaTypes(sourceTypes),
		}
		targetCfg := &TargetConfig{
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
			logger.Printf("Migration failed: %v", err)
			return 1
		}
	}

	reporter.Summary()
	logger.Println("Migration completed successfully!")
	return 0
}

// splitMediaTypes splits a comma-separated media type list into a slice.
// An empty input yields nil (meaning "migrate all types").
func splitMediaTypes(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var types []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			types = append(types, t)
		}
	}
	return types
}

// findLatestState returns the most recently modified migration sub-directory
// under stateDir (the resume checkpoint source).
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
