/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// main is the M1 monolith entry point for origcms.
// Wires up ent (SQLite/PostgreSQL), svc-user biz/data, and a Gin HTTP server.
// Run: go run ./cmd/server (zero-config with SQLite by default)
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	confhelper "origadmin/application/origcms/internal/helpers/conf"
	"origadmin/application/origcms/internal/server"
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
		slog.Error(
			"Could not find configuration file. Searched -conf flag, executable path, and development path.",
		)
		os.Exit(1)
	}

	slog.Info("Loading configuration from: " + confPath)

	// Load config
	cfg := loadConfig()

	// ── 1. Database ──────────────────────────────────────────────────────────
	db, err := openDB(cfg.DBSource, cfg.DatabaseType)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	// AutoMigrate all schemas in the unified entity
	if err := db.Schema.Create(ctx); err != nil {
		slog.Error("ent AutoMigrate failed", "err", err)
		os.Exit(1)
	}
	slog.Info("database migration complete")

	// ── 2. Dependency injection (manual wire) ────────────────────────────────
	hasher, err := hash.NewCrypto(hashtypes.BCRYPT)
	if err != nil {
		slog.Error("failed to create hasher", "err", err)
		os.Exit(1)
	}

	logger := log.NewStdLogger(os.Stderr)
	userRepo := data.NewUserRepo(db)
	userUC := biz.NewUserUseCase(userRepo, hasher, logger)

	jwtManager := auth.NewManager(
		cfg.JWTSecret,
		time.Duration(cfg.JWTExpireHour)*time.Hour,
	)

	authHandler := server.NewAuthHandler(userUC, jwtManager)

	// ── 3. Gin router ─────────────────────────────────────────────────────────
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Static files for media uploads
	r.Static("/uploads", "./data/uploads")

	// Register all module routes (user, media, content, etc.)
	server.RegisterRoutes(r, db)

	// Auth routes (public) - these supplement the standard CRUD routes
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/signin", authHandler.Login)
		authGroup.POST("/signup", authHandler.Register)
		authGroup.POST("/signout", authHandler.Logout)
	}

	// Protected auth routes
	protected := r.Group("/api/v1/auth")
	protected.Use(server.JWTMiddleware(jwtManager))
	{
		protected.GET("/me", authHandler.Me)
	}

	// ── 4. Start server ───────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	slog.Info("origcms server starting", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

// config holds all runtime configuration from environment/config file.
type config struct {
	DBSource      string
	ServerPort    int
	JWTSecret     string
	JWTExpireHour int
	DatabaseType  string // sqlite3 or postgres
}

func loadConfig() config {
	// Default to SQLite for zero-config testing
	return config{
		DBSource:      getEnv("DB_SOURCE", "./data/origcms.db"),
		ServerPort:    getEnvInt("SERVER_PORT", 9090),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpireHour: getEnvInt("JWT_EXPIRE_HOUR", 24),
		DatabaseType:  getEnv("DATABASE_TYPE", "sqlite3"),
	}
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
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("create database %s: %w", dbName, err)
		}
		slog.Info("Created database: " + dbName)
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

func getEnvInt(key string, defaultVal int) int {
	if v, ok := os.LookupEnv(key); ok {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return defaultVal
}
