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

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/origadmin/runtime"
	"github.com/origadmin/runtime/log"
	_ "github.com/origadmin/runtime/config/envsource"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"
	_ "github.com/sqlite3ent/sqlite3"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	confhelper "origadmin/application/origcms/internal/helpers/conf"
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
	if err := rt.Load(confPath); err != nil {
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
	db, err := openDB(cfg.Data.Database.Source, cfg.Data.Database.Driver)
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

	jwtSecret := cfg.Security.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	jwtExpire := cfg.Security.JWT.ExpireHour
	if jwtExpire == 0 {
		jwtExpire = 24
	}
	jwtManager := auth.NewManager(
		jwtSecret,
		time.Duration(jwtExpire)*time.Hour,
	)

	// svc-media initialization
	mediaRepo := mediadata.NewMediaRepo(db)
	profileRepo := mediadata.NewEncodeProfileRepo(db)
	taskRepo := mediadata.NewEncodingTaskRepo(db)

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
type Config struct {
	Data struct {
		Database struct {
			Driver string `yaml:"driver"`
			Source string `yaml:"source"`
		} `yaml:"database"`
	} `yaml:"data"`
	Server struct {
		HTTP struct {
			Network string `yaml:"network"`
			Addr    string `yaml:"addr"`
		} `yaml:"http"`
		GRPC struct {
			Network string `yaml:"network"`
			Addr    string `yaml:"addr"`
		} `yaml:"grpc"`
	} `yaml:"server"`
	Security struct {
		JWT struct {
			Secret     string `yaml:"secret"`
			ExpireHour int    `yaml:"expire_hour"`
		} `yaml:"jwt"`
	} `yaml:"security"`
}

// openDB opens an ent client using entity.Open
func openDB(dsn, dbType string) (*entity.Client, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		// Ensure database exists before connecting
		if err := ensurePostgresDB(dsn); err != nil {
			return nil, fmt.Errorf("ensure database: %w", err)
		}
		// Add sslmode if not present
		if !contains(dsn, "sslmode") {
			dsn = dsn + " sslmode=disable"
		}
	}
	return entity.Open(driverName, dsn)
}

// ensurePostgresDB creates the database if it doesn't exist
func ensurePostgresDB(dsn string) error {
	// Parse DSN to extract connection info
	connStr, dbName := parsePostgresDSN(dsn)
	if dbName == "" {
		return nil
	}

	// Connect to default 'postgres' database to create our DB
	defaultDSN := connStr + " dbname=postgres sslmode=disable"
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

// parsePostgresDSN extracts connection string and database name from DSN
func parsePostgresDSN(dsn string) (connStr, dbName string) {
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
	return
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAny(s, substr))
}

func containsAny(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
