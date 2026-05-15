/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server handles server startup and routing for the application.
package server

import (
	"context"
	"encoding/json"
	"mime"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/origadmin/runtime/log"
	http2 "origadmin/application/origstudio/internal/helpers/http"
	ginadapter "origadmin/application/origstudio/internal/helpers/http/gin"
	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/web"
)

// Module defines the interface for route registration.
// All feature-module handlers must implement this interface so that the
// server can iterate over them uniformly.
// It is a type alias for http2.Module (i.e., RegisterRoutes(r http2.Router)).
type Module = http2.Module

// Server represents the application server.
type Server struct {
	modules      []Module
	entityClient *entity.Client
	jwtMgr       *auth.Manager
	paths        *conf.StoragePaths
	settingUC    SettingProvider
	rateLimiter  interface{ Middleware() gin.HandlerFunc; Stop() }
}

type SettingProvider interface {
	Get(ctx context.Context, key string) string
}

func NewServer(
	modules []Module,
	entityClient *entity.Client,
	jwtMgr *auth.Manager,
	paths *conf.StoragePaths,
) *Server {
	return &Server{
		modules:      modules,
		entityClient: entityClient,
		jwtMgr:       jwtMgr,
		paths:        paths,
	}
}

func (s *Server) SetSettingProvider(uc SettingProvider) {
	s.settingUC = uc
}

func (s *Server) SetRateLimiter(rl interface{ Middleware() gin.HandlerFunc; Stop() }) {
	s.rateLimiter = rl
}

func (s *Server) Stop() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
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
		origin := c.GetHeader("Origin")
		allowedOrigin := "*"

		if s.settingUC != nil && origin != "" {
			urls := s.settingUC.Get(c.Request.Context(), "base_urls")
			if urls != "" {
				var allowed []string
				if err := json.Unmarshal([]byte(urls), &allowed); err == nil {
					for _, u := range allowed {
						if u != "" && u == origin {
							allowedOrigin = origin
							break
						}
					}
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Range")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Rate limiting (applied to all routes)
	if s.rateLimiter != nil {
		r.Use(s.rateLimiter.Middleware())
	}

	// Register static file routes using StoragePaths
	absBase := s.paths.BasePath()
	log.Infof("static file storage base: %s", absBase)

	for urlPrefix, fsDir := range s.paths.StaticRouteMap() {
		r.Static(urlPrefix, fsDir)
	}

	// Register all module routes
	s.RegisterRoutes(r)

	// Register frontend SPA routes (auto-detect: serves embedded dist if present)
	web.RegisterRoutes(r)

	// Print access URLs
	displayAddr := addr
	if len(displayAddr) > 0 && displayAddr[0] == ':' {
		displayAddr = "localhost" + displayAddr
	}

	log.Infof("=")
	log.Infof("  OrigStudio Server Started")
	log.Infof("=")

	if !web.IsDistEmpty() {
		log.Infof("  -> Web UI (Embedded): http://%s", displayAddr)
	} else {
		log.Infof("  -> Backend API Only")
		log.Infof("  (Frontend not embedded - run dev server separately)")
	}

	log.Infof("  -> API Base: http://%s/api/v1", displayAddr)
	log.Infof("  -> Health Check: http://%s/health", displayAddr)
	log.Infof("=")
	
	log.Infof("listening on: %s", addr)
	return r.Run(addr)
}

// RegisterRoutes registers all routes for the server.
func (s *Server) RegisterRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", HealthHandler)

	// API v1 routes — adapt *gin.RouterGroup to http2.Router
	apiV1 := r.Group("/api/v1")
	router := ginadapter.NewRouterAdapter(apiV1)

	// Register all handler modules
	for _, mod := range s.modules {
		mod.RegisterRoutes(router)
	}
}

// getEnv gets an environment variable or returns the default value.
func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
