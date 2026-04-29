/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides HTTP handlers for admin operations.
package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-admin/service"
	authbiz "origadmin/application/origcms/internal/svc-auth/biz"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/dto"
	systembiz "origadmin/application/origcms/internal/svc-system/biz"
	userbiz "origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/validation"
)

// AdminHandler handles admin-related routes.
type AdminHandler struct {
	jwt         *auth.Manager
	mediaUC     *mediabiz.MediaUseCase
	channelUC   *contentbiz.PlaylistChannelUseCase
	tagService  *service.TagService
	settingUC   *systembiz.SettingUseCase
	categoryUC  *contentbiz.CategoryTagUseCase
	articleUC   *contentbiz.ArticleUseCase
	userUC      *userbiz.UserUseCase
	permChecker authbiz.PermissionChecker
}

func NewAdminHandler(
	jwt *auth.Manager,
	mediaUC *mediabiz.MediaUseCase,
	channelUC *contentbiz.PlaylistChannelUseCase,
	tagService *service.TagService,
	settingUC *systembiz.SettingUseCase,
	categoryUC *contentbiz.CategoryTagUseCase,
	articleUC *contentbiz.ArticleUseCase,
	userUC *userbiz.UserUseCase,
	permChecker authbiz.PermissionChecker,
) *AdminHandler {
	return &AdminHandler{
		jwt:         jwt,
		mediaUC:     mediaUC,
		channelUC:   channelUC,
		tagService:  tagService,
		settingUC:   settingUC,
		categoryUC:  categoryUC,
		articleUC:   articleUC,
		userUC:      userUC,
		permChecker: permChecker,
	}
}

func (h *AdminHandler) Register(r handler.Router) {
	admin := r.Group("/admin")
	{
		// ================================
		// 1. Stats Panel
		// ================================
		stats := admin.Group("/stats")
		{
			stats.GET("/dashboard", WithAdmin(h.jwt, h.getDashboardStats()))
			stats.GET("/medias", WithAdmin(h.jwt, h.getMediaStats()))
			stats.GET("/users", WithAdmin(h.jwt, h.getUserStats()))
			stats.GET("/traffic", WithAdmin(h.jwt, h.getTrafficStats()))
		}

		// ================================
		// 2. Channel Management (Admin - UUID only) ⭐
		// ================================
		channels := admin.Group("/channels")
		{
			channels.GET("", WithAdmin(h.jwt, h.adminListChannels()))
			channels.GET("/:id", WithAdmin(h.jwt, h.adminGetChannelDetail()))
			channels.PUT("/:id", WithAdminAndPerm(h.jwt, h.permChecker, "media:write", h.adminUpdateChannel()))
			channels.DELETE("/:id", WithAdminAndPerm(h.jwt, h.permChecker, "media:delete", h.adminDeleteChannel()))
		}

		// ================================
		// 3. Encoding Management
		// ================================
		encoding := admin.Group("/encoding")
		{
			encoding.GET("/tasks", WithAdmin(h.jwt, h.getAllEncodingTasks()))
			encoding.GET("/status", WithAdmin(h.jwt, h.getEncodingStatus()))
			encoding.POST("/tasks/:taskId/retry", WithAdmin(h.jwt, h.retryTask()))
			encoding.POST("/retry-failed", WithAdmin(h.jwt, h.retryAllFailedTasks()))

			profiles := encoding.Group("/profiles")
			{
				profiles.GET("", WithAdmin(h.jwt, h.listEncodeProfiles()))
				profiles.POST("", WithAdmin(h.jwt, h.createEncodeProfile()))
				profiles.POST("/preview", WithAdmin(h.jwt, h.previewEncodeCommand()))
				profiles.GET("/:id", WithAdmin(h.jwt, h.getEncodeProfile()))
				profiles.PUT("/:id", WithAdmin(h.jwt, h.updateEncodeProfile()))
				profiles.DELETE("/:id", WithAdmin(h.jwt, h.deleteEncodeProfile()))
			}
		}

		// ================================
		// 3. System Settings
		// ================================
		settings := admin.Group("/settings")
		{
			settings.GET("", WithAdmin(h.jwt, h.getSystemSettings()))
			settings.PUT("", WithAdminAndPerm(h.jwt, h.permChecker, "system:config", h.updateSystemSettings()))
		}

		// ================================
		// 4. Tag Management
		// ================================
		tags := admin.Group("/tags")
		{
			tags.GET("", WithAdmin(h.jwt, h.listTags()))
			tags.GET("/:id", WithAdmin(h.jwt, h.getTag()))
			tags.POST("", WithAdmin(h.jwt, h.createTag()))
			tags.PUT("/:id", WithAdmin(h.jwt, h.updateTag()))
			tags.DELETE("/:id", WithAdmin(h.jwt, h.deleteTag()))
			tags.POST("/bulk", WithAdmin(h.jwt, h.bulkTagOperation()))
			tags.GET("/export", WithAdmin(h.jwt, h.exportTags()))
			tags.POST("/import", WithAdmin(h.jwt, h.importTags()))
		}

		// ================================
		// 5. Playlist Management
		// ================================
		playlists := admin.Group("/playlists")
		{
			playlists.GET("", WithAdmin(h.jwt, h.adminListPlaylists()))
			playlists.GET("/:id", WithAdmin(h.jwt, h.adminGetPlaylistDetail())) // :id = UUID
			playlists.POST("", WithAdmin(h.jwt, h.adminCreatePlaylist()))
			playlists.PUT("/:id", WithAdmin(h.jwt, h.adminUpdatePlaylist()))    // :id = UUID
			playlists.DELETE("/:id", WithAdmin(h.jwt, h.adminDeletePlaylist())) // :id = UUID
		}

		// ================================
		// 6. User Management
		// ================================
		users := admin.Group("/users")
		{
			users.GET("", WithAdmin(h.jwt, h.adminListUsers()))
			users.GET("/:id", WithAdmin(h.jwt, h.adminGetUser()))
			users.PUT("/:id", WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminUpdateUser()))
			users.DELETE("/:id", WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminDeleteUser()))
			users.PATCH("/:id/status", WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminUpdateUserStatus()))
			users.PATCH("/:id/role", WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminUpdateUserRole()))
		}

		// ================================
		// 7. Category Management
		// ================================
		categories := admin.Group("/categories")
		{
			categories.GET("", WithAdmin(h.jwt, h.adminListCategories()))
			categories.GET("/:id", WithAdmin(h.jwt, h.adminGetCategory()))
			categories.POST("", WithAdmin(h.jwt, h.adminCreateCategory()))
			categories.PUT("/:id", WithAdmin(h.jwt, h.adminUpdateCategory()))
			categories.PATCH("/:id", WithAdmin(h.jwt, h.adminPatchCategory()))
			categories.DELETE("/:id", WithAdmin(h.jwt, h.adminDeleteCategory()))
		}

		// ================================
		// 8. Article Management
		// ================================
		articles := admin.Group("/articles")
		{
			articles.GET("", WithAdmin(h.jwt, h.adminListArticles()))
			articles.GET("/:id", WithAdmin(h.jwt, h.adminGetArticle()))
			articles.POST("", WithAdmin(h.jwt, h.adminCreateArticle()))
			articles.PUT("/:id", WithAdmin(h.jwt, h.adminUpdateArticle()))
			articles.DELETE("/:id", WithAdmin(h.jwt, h.adminDeleteArticle()))
			articles.PATCH("/:id/state", WithAdmin(h.jwt, h.adminUpdateArticleState()))
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

		var filterType mediabiz.FilterType
		var specificStatus string

		switch status {
		case "":
			filterType = mediabiz.FilterTypeAll
		case "active":
			filterType = mediabiz.FilterTypeActive
		case "all":
			filterType = mediabiz.FilterTypeAll
		default:
			parsedStatus := enums.ParseEncodingTaskStatus(status)
			if parsedStatus == enums.EncodingTaskStatusUnknown {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter"})
				return
			}
			filterType = mediabiz.FilterTypeSpecific
			specificStatus = status
		}

		filter := &mediabiz.TranscodingStatusFilter{
			FilterType:    filterType,
			Status:        specificStatus,
			Page:          1,
			PageSize:      25,
			OnlyStats:     false,
			ProfileFilter: c.Query("profile"),
			ChunkFilter:   c.Query("chunk"),
			SearchQuery:   c.Query("search"),
		}

		if os := c.Query("only_stats"); os == "true" {
			filter.OnlyStats = true
		}

		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p >= 1 {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "25")); err == nil && ps >= 1 &&
			ps <= 100 {
			filter.PageSize = ps
		}

		var mediaID *string
		if m := c.Query("media_id"); m != "" {
			mediaID = &m
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
		var profile dto.EncodeProfile
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
		var profile dto.EncodeProfile
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

func (h *AdminHandler) previewEncodeCommand() gin.HandlerFunc {
	return func(c *gin.Context) {
		var profile dto.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		preview := h.mediaUC.GenerateCommandPreview(c.Request.Context(), &profile)
		c.JSON(http.StatusOK, gin.H{"command": preview})
	}
}

// --- System Settings Handlers ---

func (h *AdminHandler) getSystemSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		items, err := h.settingUC.ListAll(c.Request.Context())
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		grouped := make(map[string][]*entity.Setting)
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			cat := string(item.Category)
			grouped[cat] = append(grouped[cat], masked)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    grouped,
		})
	}
}

func (h *AdminHandler) updateSystemSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		var req struct {
			Settings []struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value"`
			} `json:"settings" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		var updated []*entity.Setting
		for _, item := range req.Settings {
			existing, err := h.settingUC.GetByKey(c.Request.Context(), item.Key)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"code": ErrNotFound, "message": "setting not found: " + item.Key},
				)
				return
			}

			if err := validateSettingValue(item.Value, existing.Type); err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{
						"code":    ErrBadRequest,
						"message": "invalid value for " + item.Key + ": " + err.Error(),
					},
				)
				return
			}

			s := &entity.Setting{
				Key:           existing.Key,
				Value:         item.Value,
				Type:          existing.Type,
				Category:      existing.Category,
				Description:   existing.Description,
				IsSensitive:   existing.IsSensitive,
				FallbackValue: existing.FallbackValue,
				IsBuiltin:     existing.IsBuiltin,
			}
			result, err := h.settingUC.Upsert(c.Request.Context(), s)
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{"code": ErrInternal, "message": err.Error()},
				)
				return
			}
			updated = append(updated, result)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

// --- Tag Management Handlers ---

func (h *AdminHandler) listTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		search := c.Query("search")
		status := c.Query("status")
		sortBy := c.DefaultQuery("sort_by", "created_at")
		sortOrder := c.DefaultQuery("sort_order", "desc")

		// Get tags
		tags, total, err := h.tagService.List(
			c.Request.Context(),
			page,
			pageSize,
			search,
			status,
			sortBy,
			sortOrder,
		)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": 10000, "message": "Failed to list tags"},
			)
			return
		}

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		// Return response
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":       tags,
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": totalPages,
			},
		})
	}
}

func (h *AdminHandler) getTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		tag, err := h.tagService.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 10001, "message": "Tag not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    tag,
		})
	}
}

func (h *AdminHandler) createTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug" binding:"required"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": "Invalid request"})
			return
		}

		tag := &entity.Tag{
			Title: req.Name,
		}

		createdTag, err := h.tagService.Create(c.Request.Context(), tag)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Tag created successfully",
			"data":    createdTag,
		})
	}
}

func (h *AdminHandler) updateTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": "Invalid request"})
			return
		}

		updates := &entity.Tag{
			Title: req.Name,
		}

		updatedTag, err := h.tagService.Update(c.Request.Context(), id, updates)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Tag updated successfully",
			"data":    updatedTag,
		})
	}
}

func (h *AdminHandler) deleteTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := h.tagService.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Tag deleted successfully",
		})
	}
}

func (h *AdminHandler) bulkTagOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": "Invalid request"})
			return
		}

		if req.Action != "delete" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": "Unsupported action"})
			return
		}

		count, err := h.tagService.BulkDelete(c.Request.Context(), req.IDs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Bulk operation completed",
			"data": gin.H{
				"success": count,
				"failed":  len(req.IDs) - count,
			},
		})
	}
}

func (h *AdminHandler) exportTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement export functionality
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Export functionality not implemented yet",
		})
	}
}

func (h *AdminHandler) importTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement import functionality
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Import functionality not implemented yet",
		})
	}
}

// ================================
// Admin Channel Management Handlers (v3.2 - UUID only)
// ================================

func (h *AdminHandler) adminListChannels() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if pageSize > 100 {
			pageSize = 100
		}

		items, total, err := h.channelUC.ListChannels(c.Request.Context(), page, pageSize)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *AdminHandler) adminGetChannelDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "channel id is required"},
			)
			return
		}

		if !validation.IsValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrBadRequest,
				"message": "invalid_uuid_format",
				"hint":    "Admin channel API requires UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)",
			})
			return
		}

		ch, err := h.channelUC.GetChannel(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    ErrNotFound,
				"message": "channel_not_found",
				"hint":    "No channel exists with the provided UUID",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    ch,
		})
	}
}

func (h *AdminHandler) adminUpdateChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "channel id is required"},
			)
			return
		}

		if !validation.IsValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrBadRequest,
				"message": "invalid_uuid_format",
				"hint":    "Admin channel API requires UUID format",
			})
			return
		}

		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			BannerLogo  string `json:"banner_logo"`
			IsPublic    *bool  `json:"is_public"`
			IsDefault   *bool  `json:"is_default"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		existingCh, err := h.channelUC.GetChannel(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel_not_found"})
			return
		}

		chItem := &contentbiz.Channel{
			ID:          id,
			Title:       input.Title,
			Description: input.Description,
			BannerLogo:  input.BannerLogo,
			IsPublic:    existingCh.IsPublic,
			UserID:      existingCh.UserID,
			CreatedAt:   existingCh.CreatedAt,
		}

		if input.IsPublic != nil {
			chItem.IsPublic = *input.IsPublic
		}
		if input.IsDefault != nil {
			chItem.IsDefault = *input.IsDefault
		}

		updated, err := h.channelUC.UpdateChannel(c.Request.Context(), chItem, "", true)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *AdminHandler) adminDeleteChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "channel id is required"},
			)
			return
		}

		if !validation.IsValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrBadRequest,
				"message": "invalid_uuid_format",
				"hint":    "Admin channel API requires UUID format",
			})
			return
		}

		err := h.channelUC.DeleteChannel(c.Request.Context(), id, "", true)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    gin.H{"message": "channel deleted"},
		})
	}
}

// ================================
// Admin Playlist Management Handlers
// ================================

func (h *AdminHandler) adminListPlaylists() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if pageSize > 100 {
			pageSize = 100
		}

		items, total, err := h.channelUC.ListPlaylists(c.Request.Context(), page, pageSize)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *AdminHandler) adminGetPlaylistDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "playlist id is required"},
			)
			return
		}

		if !validation.IsValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrBadRequest,
				"message": "invalid_uuid_format",
				"hint":    "Admin playlist API requires UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)",
			})
			return
		}

		playlist, err := h.channelUC.GetPlaylist(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    ErrNotFound,
				"message": "playlist_not_found",
				"hint":    "No playlist exists with the provided UUID",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    playlist,
		})
	}
}

func (h *AdminHandler) adminCreatePlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			UserID      string `json:"user_id" binding:"required"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		isPublic := true
		if input.IsPublic != nil {
			isPublic = *input.IsPublic
		}

		playlist := &contentbiz.Playlist{
			Title:       input.Title,
			Description: input.Description,
			UserID:      input.UserID,
			IsPublic:    isPublic,
		}

		created, err := h.channelUC.CreatePlaylist(c.Request.Context(), playlist)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    created,
		})
	}
}

func (h *AdminHandler) adminUpdatePlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "playlist id is required"},
			)
			return
		}

		if !validation.IsValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrBadRequest,
				"message": "invalid_uuid_format",
				"hint":    "Admin playlist API requires UUID format",
			})
			return
		}

		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		existingPlaylist, err := h.channelUC.GetPlaylist(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "playlist_not_found"})
			return
		}

		playlistItem := &contentbiz.Playlist{
			ID:          id,
			Title:       input.Title,
			Description: input.Description,
			UserID:      existingPlaylist.UserID,
			IsPublic:    existingPlaylist.IsPublic,
			ShortToken:  existingPlaylist.ShortToken,
		}

		if input.IsPublic != nil {
			playlistItem.IsPublic = *input.IsPublic
		}

		updated, err := h.channelUC.UpdatePlaylist(c.Request.Context(), playlistItem, "", true)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *AdminHandler) adminDeletePlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "playlist id is required"},
			)
			return
		}

		if !validation.IsValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrBadRequest,
				"message": "invalid_uuid_format",
				"hint":    "Admin playlist API requires UUID format",
			})
			return
		}

		err := h.channelUC.DeletePlaylist(c.Request.Context(), id, "", true)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"code": ErrInternal, "message": err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    gin.H{"message": "playlist deleted"},
		})
	}
}

// ================================
// Admin User Management Handlers
// ================================

func (h *AdminHandler) adminListUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if pageSize > 100 {
			pageSize = 100
		}

		users, total, err := h.userUC.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		// Paginate the results manually since ListUsers doesn't support pagination
		start := (page - 1) * pageSize
		end := start + pageSize
		totalInt := int(total)
		if start > totalInt {
			start = totalInt
		}
		if end > totalInt {
			end = totalInt
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     users[start:end],
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *AdminHandler) adminGetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "user id is required"})
			return
		}

		user, err := h.userUC.GetUser(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    user,
		})
	}
}

func (h *AdminHandler) adminUpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "user id is required"})
			return
		}

		existing, err := h.userUC.GetUser(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "user not found"})
			return
		}

		var input struct {
			Nickname string `json:"nickname"`
			Email    string `json:"email"`
			Avatar   string `json:"avatar"`
			Phone    string `json:"phone"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		if input.Nickname != "" {
			existing.Nickname = input.Nickname
		}
		if input.Email != "" {
			existing.Email = input.Email
		}
		if input.Avatar != "" {
			existing.Avatar = input.Avatar
		}
		if input.Phone != "" {
			existing.Phone = input.Phone
		}

		updated, err := h.userUC.UpdateUser(c.Request.Context(), existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *AdminHandler) adminDeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "user id is required"})
			return
		}

		if err := h.userUC.DeleteUser(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    gin.H{"message": "deleted"},
		})
	}
}

func (h *AdminHandler) adminUpdateUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "user id is required"})
			return
		}

		var input struct {
			Status int32 `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		if err := h.userUC.UpdateUserStatus(c.Request.Context(), id, int8(input.Status)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
		})
	}
}

// ================================
// Admin Article Management Handlers
// ================================

func (h *AdminHandler) adminListArticles() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if pageSize > 100 {
			pageSize = 100
		}

		filters := make(map[string]interface{})
		if v := c.Query("status"); v != "" {
			filters["status"] = v
		}
		if v := c.Query("category_id"); v != "" {
			filters["category_id"] = v
		}
		if v := c.Query("keyword"); v != "" {
			filters["keyword"] = v
		}

		articles, total, err := h.articleUC.List(c.Request.Context(), page, pageSize, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     articles,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *AdminHandler) adminGetArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "article id is required"})
			return
		}

		article, err := h.articleUC.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "article not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    article,
		})
	}
}

func (h *AdminHandler) adminCreateArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title      string `json:"title" binding:"required"`
			Slug       string `json:"slug" binding:"required"`
			Content    string `json:"content"`
			Summary    string `json:"summary"`
			CategoryID int64  `json:"category_id"`
			State      string `json:"state"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		article := &contentbiz.Article{
			Title:      input.Title,
			Slug:       input.Slug,
			Content:    input.Content,
			Summary:    input.Summary,
			CategoryID: input.CategoryID,
			State:      input.State,
		}

		created, err := h.articleUC.Create(c.Request.Context(), article)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    created,
		})
	}
}

func (h *AdminHandler) adminUpdateArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "article id is required"})
			return
		}

		var input struct {
			Title      string `json:"title"`
			Slug       string `json:"slug"`
			Content    string `json:"content"`
			Summary    string `json:"summary"`
			CategoryID int64  `json:"category_id"`
			State      string `json:"state"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		existing, err := h.articleUC.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "article not found"})
			return
		}

		if input.Title != "" {
			existing.Title = input.Title
		}
		if input.Slug != "" {
			existing.Slug = input.Slug
		}
		if input.Content != "" {
			existing.Content = input.Content
		}
		if input.Summary != "" {
			existing.Summary = input.Summary
		}
		if input.CategoryID != 0 {
			existing.CategoryID = input.CategoryID
		}
		if input.State != "" {
			existing.State = input.State
		}

		updated, err := h.articleUC.Update(c.Request.Context(), existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *AdminHandler) adminDeleteArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "article id is required"})
			return
		}

		if err := h.articleUC.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
		})
	}
}

func (h *AdminHandler) adminUpdateArticleState() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "article id is required"})
			return
		}

		var input struct {
			State string `json:"state" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		if err := h.articleUC.UpdateState(c.Request.Context(), id, input.State); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
		})
	}
}

// ================================
// Admin Category Management Handlers
// ================================

func (h *AdminHandler) adminListCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := h.categoryUC.ListCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if pageSize > 100 {
			pageSize = 100
		}

		total := len(categories)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     categories[start:end],
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *AdminHandler) adminGetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
			return
		}

		cat, err := h.categoryUC.GetCategory(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "category not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    cat,
		})
	}
}

func (h *AdminHandler) adminCreateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug" binding:"required"`
			Description string `json:"description"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		cat := &contentbiz.Category{
			Name:        input.Name,
			Slug:        input.Slug,
			Description: input.Description,
		}

		created, err := h.categoryUC.CreateCategory(c.Request.Context(), cat)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    created,
		})
	}
}

func (h *AdminHandler) adminUpdateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
			return
		}

		var input contentbiz.UpdateCategoryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(c.Request.Context(), id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *AdminHandler) adminPatchCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
			return
		}

		var input contentbiz.UpdateCategoryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(c.Request.Context(), id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *AdminHandler) adminDeleteCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
			return
		}

		if err := h.categoryUC.DeleteCategory(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
		})
	}
}

func (h *AdminHandler) adminUpdateUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "user id is required"})
			return
		}

		var input struct {
			Role string `json:"role" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		validRoles := map[string]bool{"user": true, "admin": true, "moderator": true}
		if !validRoles[input.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid role, must be one of: user, admin, moderator"})
			return
		}

		if err := h.userUC.SetUserRole(c.Request.Context(), id, input.Role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
		})
	}
}
