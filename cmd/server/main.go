/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// main is the M1 monolith entry point for origcms.
// Uses runtime for config loading and logger initialization.
// Wires up ent (SQLite/PostgreSQL), svc-user biz/data, and a Gin HTTP server.
// Run: go run ./cmd/server -conf configs/bootstrap.yaml
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/origadmin/runtime"
	"github.com/origadmin/runtime/engine/bootstrap"
	"github.com/origadmin/runtime/log"
	_ "github.com/origadmin/runtime/config/envsource"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"
	_ "github.com/sqlite3ent/sqlite3"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	confhelper "origadmin/application/origcms/internal/helpers/conf"
	"origadmin/application/origcms/internal/pubsub"
	"origadmin/application/origcms/internal/server"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	mediadata "origadmin/application/origcms/internal/svc-media/data"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/data"
)

var (
	// envName is the environment name suffix for .env file
	envName = ".server"
	// flagconf is the config flag
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "", "config path, eg: -conf bootstrap.yaml")
}

func main() {
	flag.Parse()

	// Initialize environment variables and find configuration file
	confPath := confhelper.InitEnvAndConf(envName, flagconf)
	if confPath == "" {
		log.Fatalf(
			"Could not find configuration file. Searched -conf flag, executable path, and development path.",
		)
	}

	log.Infof("Loading configuration from: %s\n", confPath)

	// ── 1. Runtime: config loading + logger initialization ────────────
	rt := runtime.New("origcms.server", "v1.0.0")
	if err := rt.Load(confPath, bootstrap.WithDirectly(true), bootstrap.WithEnvSource()); err != nil {
		log.Fatalf("failed to load runtime: %v", err)
	}
	defer func() {
		_ = rt.Decoder().Close()
	}()

	// Set runtime logger as global logger
	log.SetLogger(rt.Logger())

	rt.ShowAppInfo()

	// Decode business config from YAML
	cfg := &Config{}
	if rt.Decoder() != nil {
		if err := rt.Decoder().Scan(cfg); err != nil {
			log.Fatalf("failed to scan config: %v", err)
		}
	}

	// ── 2. Database ──────────────────────────────────────────────────
	dbDialect, dbSource := cfg.GetDefaultDB()
	db, err := openDB(dbSource, dbDialect)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	// AutoMigrate all schemas in the unified entity
	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("ent AutoMigrate failed: %v", err)
	}
	log.Info("database migration complete")

	// --- Initialize Seed Data ---
	if err := mediadata.SeedEncodeProfiles(ctx, db); err != nil {
		log.Fatalf("failed to seed encode profiles: %v", err)
	}
	log.Info("encode profiles seeded successfully")

	// ── 3. Dependency injection (manual wire) ────────────────────────
	hasher, err := hash.NewCrypto(hashtypes.BCRYPT)
	if err != nil {
		log.Fatalf("failed to create hasher: %v", err)
	}

	logger := rt.Logger()
	userRepo := data.NewUserRepo(db)
	userUC := biz.NewUserUseCase(userRepo, hasher, logger)

	jwtSigningKey, _, jwtTTL := cfg.GetJWTConfig()
	jwtExpire := parseDuration(jwtTTL, 3600*time.Second)
	jwtManager := auth.NewManager(
		jwtSigningKey,
		jwtExpire,
	)

	// svc-media initialization
	mediaRepo := mediadata.NewMediaRepo(db)
	profileRepo := mediadata.NewEncodeProfileRepo(db)
	taskRepo := mediadata.NewEncodingTaskRepo(db)

	// Reset stale processing media (recovery from service restart)
	resetCount, err := mediaRepo.ResetStaleProcessing(ctx)
	if err != nil {
		log.Warnf("failed to reset stale processing media: %v", err)
	} else if resetCount > 0 {
		log.Infof("reset %d media items from 'processing' back to 'pending' (service restart recovery)", resetCount)
	}

	mediaUC := mediabiz.NewMediaUseCase(mediaRepo, profileRepo, taskRepo, logger)

	uploadRepo := mediadata.NewUploadRepo(db, logger)
	storage := mediadata.NewLocalStorage("./data/uploads", logger)
	uploadUC := mediabiz.NewUploadUseCase(
		uploadRepo,
		mediaRepo,
		profileRepo,
		taskRepo,
		mediaUC,
		storage,
		logger,
	)

	// ── 3b. Watermill GoChannel + Transcode Pipeline ─────────────────
	wmLogger := watermill.NewStdLogger(true, true)
	ps := pubsub.NewGoChannel(64, wmLogger)

	maxWorkers := int32(envInt("TRANSCODE_MAX_WORKERS", 3))
	worker := mediabiz.NewGoroutineWorker(maxWorkers, log.NewHelper(log.With(logger, "module", "transcode.worker")))

	transcodeHandler := mediabiz.NewTranscodeHandler(
		mediaUC,
		profileRepo,
		taskRepo,
		mediaRepo,
		worker,
		ps.Pub,
		logger,
		"./data/uploads",
	)

	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		log.Fatalf("failed to create Watermill router: %v", err)
	}
	router.AddHandler(
		"media_transcode",
		pubsub.MediaEncodeRequestTopic,
		ps.Sub,
		"",  // no output topic (handler publishes directly)
		nil, // no output publisher needed
		func(msg *message.Message) ([]*message.Message, error) {
			return nil, transcodeHandler.Handle(msg)
		},
	)
	go func() {
		if err := router.Run(context.Background()); err != nil {
			log.Fatalf("Watermill router error: %v", err)
		}
	}()
	log.Infof("transcode pipeline started (maxWorkers=%d)", maxWorkers)

	// Inject publisher into UploadUseCase for async encoding requests
	uploadUC.SetPublisher(ps.Pub)

	// --- 4. Handlers (Monolith) ---
	authHandler := server.NewAuthHandler(userUC, jwtManager)
	userHandler := server.NewUserHandler(db)
	mediaHandler := server.NewMediaHandler(db, jwtManager, mediaUC, uploadUC)
	uploadHandler := server.NewUploadHandler(uploadUC, jwtManager)
	categoryHandler := server.NewCategoryHandler(db)
	tagHandler := server.NewTagHandler(db)
	commentHandler := server.NewCommentHandler(db)
	playlistHandler := server.NewPlaylistHandler(db)
	feedHandler := server.NewFeedHandler(db)

	// ── 5. Gin router ───────────────────────────────────────────────
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Range")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Static files for media uploads
	r.Static("/uploads", "./data/uploads/uploads")
	r.Static("/thumbnails", "./data/uploads/thumbnails")
	r.Static("/hls", "./data/uploads/hls")

	// Register all module routes using the new interface-based approach
	server.RegisterRoutes(r,
		authHandler,
		userHandler,
		mediaHandler,
		uploadHandler,
		categoryHandler,
		tagHandler,
		commentHandler,
		playlistHandler,
		feedHandler,
	)

	// ── 6. Start server ─────────────────────────────────────────────
	addr := cfg.Server.HTTP.Addr
	if addr == "" {
		addr = ":9090"
	}
	log.Infof("origcms server starting, addr: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// Config holds all runtime configuration parsed from bootstrap.yaml.
// Fields use YAML tags matching the config file structure.
// Structure aligns with backend (projects/backend/resources/configs/).
type Config struct {
	Data struct {
		Databases map[string]struct {
			Name    string `yaml:"name"`
			Dialect string `yaml:"dialect"`
			Source  string `yaml:"source"`
		} `yaml:"databases"`
	} `yaml:"data"`
	Server struct {
		HTTP struct {
			Network string `yaml:"network"`
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"http"`
	} `yaml:"server"`
	Security struct {
		Authn struct {
			Configs []struct {
				Type string `yaml:"type"`
				JWT  struct {
					SigningKey      string `yaml:"signing_key"`
					SigningMethod   string `yaml:"signing_method"`
					AccessTokenTTL  string `yaml:"access_token_ttl"`
					RefreshTokenTTL string `yaml:"refresh_token_ttl"`
				} `yaml:"jwt"`
			} `yaml:"configs"`
		} `yaml:"authn"`
	} `yaml:"security"`
}

// GetDefaultDB returns the "default" database config for convenience.
func (c *Config) GetDefaultDB() (dialect, source string) {
	if c.Data.Databases != nil {
		if db, ok := c.Data.Databases["default"]; ok {
			return db.Dialect, db.Source
		}
	}
	return "postgres", ""
}

// GetJWTConfig returns the first JWT authn config found.
func (c *Config) GetJWTConfig() (signingKey, signingMethod, accessTokenTTL string) {
	for _, cfg := range c.Security.Authn.Configs {
		if cfg.Type == "jwt" {
			return cfg.JWT.SigningKey, cfg.JWT.SigningMethod, cfg.JWT.AccessTokenTTL
		}
	}
	return "change-me-in-production", "HS256", "3600s"
}

// openDB opens an ent client using entity.Open.
// Supports both URI DSN (postgres://user:pass@host:5432/db?sslmode=disable)
// and key=value DSN (host=x user=y dbname=z).
func openDB(dsn, dbType string) (*entity.Client, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		// Ensure database exists before connecting
		if err := ensurePostgresDB(dsn); err != nil {
			return nil, fmt.Errorf("ensure database: %w", err)
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
	}
	return entity.Open(driverName, dsn)
}

// ensurePostgresDB creates the database if it doesn't exist
func ensurePostgresDB(dsn string) error {
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
		log.Info("Created database: " + dbName)
	}
	return nil
}

// replaceDBNameInURI replaces the database name in a URI-style PostgreSQL DSN.
// e.g. postgres://user:pass@host:5432/mydb?sslmode=disable
//   -> postgres://user:pass@host:5432/postgres?sslmode=disable
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

// parsePostgresDSN extracts the connection string and database name from DSN.
// Supports both URI format (postgres://user:pass@host/db?opts) and
// key=value format (host=x dbname=y).
func parsePostgresDSN(dsn string) (connStr, dbName string) {
	// URI format: postgres://user:pass@host:port/dbname?options
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return parsePostgresURIDSN(dsn)
	}
	// key=value format: host=x dbname=y
	return parsePostgresKVDSN(dsn)
}

// parsePostgresURIDSN parses a URI-style PostgreSQL DSN.
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

// parsePostgresKVDSN parses a key=value style PostgreSQL DSN.
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

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return defaultVal
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
