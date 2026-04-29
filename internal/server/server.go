/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server handles server startup and routing for the application.
package server

import (
	"mime"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/frontend"
	"origadmin/application/origcms/internal/handler"
)

// Server represents the application server.
type Server struct {
	modules         []handler.Module
	entityClient    *entity.Client
	jwtMgr          *auth.Manager
	storageBasePath string // base directory for static file serving (resolved to absolute path)
}

// NewServer creates a new server instance.
func NewServer(
	modules []handler.Module,
	entityClient *entity.Client,
	jwtMgr *auth.Manager,
	storageBasePath string,
) *Server {
	return &Server{
		modules:         modules,
		entityClient:    entityClient,
		jwtMgr:          jwtMgr,
		storageBasePath: storageBasePath,
	}
}

// Start starts the server.
func (s *Server) Start(addr string) error {
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Register HLS-related MIME types so that .m3u8 and .ts files
	// are served with the correct Content-Type header.
	mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	mime.AddExtensionType(".ts", "video/mp2t")

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

	// Resolve storage base path to absolute path to avoid working directory dependency.
	// When the server is started from a different directory (e.g., framework root
	// instead of project root), relative paths like "./data/uploads/hls" would
	// resolve incorrectly, causing 404 errors for static file requests.
	storageBase := s.storageBasePath
	if storageBase == "" {
		storageBase = "./data/uploads"
	}
	absStorageBase, err := filepath.Abs(storageBase)
	if err != nil {
		log.Warnf("failed to resolve storage base path %q: %v, using as-is", storageBase, err)
		absStorageBase = storageBase
	}
	log.Infof("static file storage base: %s (resolved to %s)", storageBase, absStorageBase)

	// Static files for media uploads — use absolute paths
	r.Static("/uploads", filepath.Join(absStorageBase, "uploads"))
	r.Static("/thumbnails", filepath.Join(absStorageBase, "thumbnails"))
	r.Static("/hls", filepath.Join(absStorageBase, "hls"))

	// Register all module routes
	s.RegisterRoutes(r)

	// Register frontend SPA routes (auto-detect: serves embedded dist if present)
	frontend.RegisterRoutes(r)

	log.Infof("origcms server starting, addr: %s", addr)
	return r.Run(addr)
}

// RegisterRoutes registers all routes for the server.
func (s *Server) RegisterRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", HealthHandler)

	// API v1 routes
	apiV1 := r.Group("/api/v1")

	// Register all handler modules
	for _, mod := range s.modules {
		mod.RegisterRoutes(apiV1)
	}
}

// getEnv gets an environment variable or returns the default value.
func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
