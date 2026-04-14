/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides HTTP handlers for admin operations.
package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/svc-media/biz"
)

// AdminHandler handles admin-related routes.
type AdminHandler struct {
	jwt     *auth.Manager
	mediaUC *biz.MediaUseCase
	// Add other use cases as needed
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(jwt *auth.Manager, mediaUC *biz.MediaUseCase) *AdminHandler {
	return &AdminHandler{jwt: jwt, mediaUC: mediaUC}
}

func (h *AdminHandler) Register(group *gin.RouterGroup) {
	admin := group.Group("/admin")
	admin.Use(JWTMiddleware(h.jwt))
	admin.Use(func(c *gin.Context) {
		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok || !claims.IsStaff {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	})

	{
		// ================================
		// 1. Stats Panel
		// ================================
		stats := admin.Group("/stats")
		{
			stats.GET("/dashboard", h.getDashboardStats())
			stats.GET("/medias", h.getMediaStats())
			stats.GET("/users", h.getUserStats())
			stats.GET("/traffic", h.getTrafficStats())
		}

		// ================================
		// 2. Encoding Management
		// ================================
		encoding := admin.Group("/encoding")
		{
			// Tasks
			encoding.GET("/tasks", h.getAllEncodingTasks())
			encoding.GET("/status", h.getEncodingStatus())
			encoding.POST("/tasks/:taskId/retry", h.retryTask())
			encoding.POST("/retry-failed", h.retryAllFailedTasks())

			// Profiles
			profiles := encoding.Group("/profiles")
			{
				profiles.GET("", h.listEncodeProfiles())
				profiles.POST("", h.createEncodeProfile())
				profiles.GET("/:id", h.getEncodeProfile())
				profiles.PUT("/:id", h.updateEncodeProfile())
				profiles.DELETE("/:id", h.deleteEncodeProfile())
			}
		}

		// ================================
		// 3. System Settings
		// ================================
		settings := admin.Group("/settings")
		{
			settings.GET("", h.getSystemSettings())
			settings.PUT("", h.updateSystemSettings())
		}
	}
}

// --- Stats Panel Handlers ---

func (h *AdminHandler) getDashboardStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement dashboard stats
		c.JSON(http.StatusOK, gin.H{
			"total_medias":   100,
			"total_users":    50,
			"total_channels": 20,
			"total_comments": 300,
			"today_uploads":  5,
			"today_views":    1000,
		})
	}
}

func (h *AdminHandler) getMediaStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement media stats
		c.JSON(http.StatusOK, gin.H{
			"total_media":      100,
			"total_storage":    "10GB",
			"video_count":      60,
			"image_count":      30,
			"audio_count":      10,
			"pending_encoding": 5,
		})
	}
}

func (h *AdminHandler) getUserStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement user stats
		c.JSON(http.StatusOK, gin.H{
			"total_users":     50,
			"active_users":    30,
			"new_users_today": 2,
			"admin_count":     2,
			"user_growth":     "10%",
		})
	}
}

func (h *AdminHandler) getTrafficStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement traffic stats
		c.JSON(http.StatusOK, gin.H{
			"total_views":     10000,
			"today_views":     1000,
			"total_bandwidth": "100GB",
			"today_bandwidth": "10GB",
			"top_media":       []interface{}{},
		})
	}
}

// --- Encoding Management Handlers ---

func (h *AdminHandler) getAllEncodingTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")

		// Validate status parameter
		if status != "" {
			if status == "active" || status == "all" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter"})
				return
			}
			parsedStatus := enums.ParseEncodingTaskStatus(status)
			if parsedStatus == enums.EncodingTaskStatusUnknown {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter"})
				return
			}
		}

		filter := &biz.TranscodingStatusFilter{
			Status:   status,
			Page:     1,
			PageSize: 25,
		}

		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p >= 1 {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "25")); err == nil && ps >= 1 &&
			ps <= 100 {
			filter.PageSize = ps
		}

		var mediaID *int64
		if m := c.Query("media_id"); m != "" {
			var id int64
			if _, err := fmt.Sscanf(m, "%d", &id); err == nil && id > 0 {
				mediaID = &id
			}
		}

		result, err := h.mediaUC.ListEncodingTasksFlat(c.Request.Context(), filter, mediaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (h *AdminHandler) getEncodingStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement encoding status
		c.JSON(http.StatusOK, gin.H{
			"total_tasks":  100,
			"pending":      5,
			"processing":   3,
			"completed":    80,
			"failed":       12,
			"success_rate": "80%",
		})
	}
}

func (h *AdminHandler) retryTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskIDStr := c.Param("taskId")
		if taskIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}

		task, err := h.mediaUC.RetryTask(c.Request.Context(), taskIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Retry initiated", "task": task})
	}
}

func (h *AdminHandler) retryAllFailedTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		mediaIDStr := c.Query("media_id")

		count, err := h.mediaUC.RetryAllFailedTasks(c.Request.Context(), mediaIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Retry initiated", "retried_count": count})
	}
}

// --- Encode Profile Handlers ---

func (h *AdminHandler) listEncodeProfiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		profiles, err := h.mediaUC.ListEncodeProfiles(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profiles": profiles})
	}
}

func (h *AdminHandler) createEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var profile biz.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p, err := h.mediaUC.CreateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"profile": p})
	}
}

func (h *AdminHandler) getEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
			return
		}
		p, err := h.mediaUC.GetEncodeProfile(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": p})
	}
}

func (h *AdminHandler) updateEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
			return
		}
		var profile biz.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		profile.Id = id
		p, err := h.mediaUC.UpdateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": p})
	}
}

func (h *AdminHandler) deleteEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
			return
		}
		if err := h.mediaUC.DeleteEncodeProfile(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// --- System Settings Handlers ---

func (h *AdminHandler) getSystemSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get system settings
		c.JSON(http.StatusOK, gin.H{
			"site_name":           "OrigCMS",
			"max_upload_size":     "5GB",
			"enable_registration": true,
			"default_privacy":     "public",
			"encoding_queue_size": 10,
		})
	}
}

func (h *AdminHandler) updateSystemSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement update system settings
		var settings struct {
			SiteName           string `json:"site_name"`
			MaxUploadSize      string `json:"max_upload_size"`
			EnableRegistration bool   `json:"enable_registration"`
			DefaultPrivacy     string `json:"default_privacy"`
			EncodingQueueSize  int    `json:"encoding_queue_size"`
		}
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// TODO: Save settings to database
		c.JSON(http.StatusOK, gin.H{"message": "Settings updated", "settings": settings})
	}
}
