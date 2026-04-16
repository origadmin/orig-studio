/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * System module - handles stats and settings
 */

package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
	systemData "origadmin/application/origcms/internal/svc-system/data"
)

// SystemHandler handles system-related routes
type SystemHandler struct {
	jwtMgr    *auth.Manager
	statsRepo *systemData.StatsRepo
}

// NewSystemHandler creates a new SystemHandler
func NewSystemHandler(jwtMgr *auth.Manager, statsRepo *systemData.StatsRepo) *SystemHandler {
	return &SystemHandler{
		jwtMgr:    jwtMgr,
		statsRepo: statsRepo,
	}
}

// Register registers all system routes
func (h *SystemHandler) Register(r handler.Router) {
	system := r.Group("/system")
	// All system routes require auth
	{
		// ========== 1. Stats sub-module ==========
		h.registerStats(system)

		// ========== 2. Settings sub-module ==========
		h.registerSettings(system)
	}
}

// registerStats handles all statistics routes
func (h *SystemHandler) registerStats(g handler.Router) {
	stats := g.Group("/stats")
	{
		// Static routes first (alphabetical order)
		stats.GET("/dashboard", GinHandlerToHTTP(h.getDashboardStats()))
		stats.GET("/media", GinHandlerToHTTP(h.getMediaStats()))
		stats.GET("/traffic", GinHandlerToHTTP(h.getTrafficStats()))
		stats.GET("/users", GinHandlerToHTTP(h.getUserStats()))
	}
}

// registerSettings handles all settings routes
func (h *SystemHandler) registerSettings(g handler.Router) {
	settings := g.Group("/settings")
	{
		// Collection routes
		settings.GET("", GinHandlerToHTTP(h.getSettings()))
		settings.PUT("", GinHandlerToHTTP(h.updateSettings()))
	}
}

// ==================== Stats Handlers ====================

func (h *SystemHandler) getDashboardStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.statsRepo == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stats service not available"})
			return
		}

		stats, err := h.statsRepo.GetDashboardStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (h *SystemHandler) getMediaStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.statsRepo == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stats service not available"})
			return
		}

		stats, err := h.statsRepo.GetMediaStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (h *SystemHandler) getUserStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.statsRepo == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stats service not available"})
			return
		}

		stats, err := h.statsRepo.GetUserStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (h *SystemHandler) getTrafficStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		// TODO: Implement real traffic stats
		_ = page
		_ = pageSize

		c.JSON(http.StatusOK, gin.H{
			"list":      []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// ==================== Settings Handlers ====================

func (h *SystemHandler) getSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get settings
		c.JSON(http.StatusOK, gin.H{
			"site_name":        "OrigCMS",
			"site_description": "A modern media content management system",
			"allow_register":   true,
			"allow_upload":     true,
			"max_upload_size":  1073741824, // 1GB
		})
	}
}

func (h *SystemHandler) updateSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement update settings
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
