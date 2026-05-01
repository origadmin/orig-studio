/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides HTTP handlers for admin operations.
package service

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "origadmin/application/origcms/api/gen/v1/user"
	mediapb "origadmin/application/origcms/api/gen/v1/media"
	types "origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/helpers/hashtag"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/validation"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	mediabiz "origadmin/application/origcms/internal/features/media/biz"
	mediadto "origadmin/application/origcms/internal/features/media/dto"
	mediaservice "origadmin/application/origcms/internal/features/media/service"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemservice "origadmin/application/origcms/internal/features/system/service"
	userdto "origadmin/application/origcms/internal/features/user/dto"
	userbiz "origadmin/application/origcms/internal/features/user/biz"
)

// AdminHandler handles admin-related routes.
type AdminHandler struct {
	jwt            *auth.Manager
	mediaUC        *mediabiz.MediaUseCase
	mediaService   *mediaservice.MediaService
	channelUC      *contentbiz.PlaylistChannelUseCase
	tagService     *TagService
	settingUC      *systembiz.SettingUseCase
	categoryUC     *contentbiz.CategoryTagUseCase
	articleUC      *contentbiz.ArticleUseCase
	userUC         *userbiz.UserUseCase
	permChecker    authbiz.PermissionChecker
}

func NewAdminHandler(
	jwt *auth.Manager,
	mediaUC *mediabiz.MediaUseCase,
	mediaService *mediaservice.MediaService,
	channelUC *contentbiz.PlaylistChannelUseCase,
	tagService *TagService,
	settingUC *systembiz.SettingUseCase,
	categoryUC *contentbiz.CategoryTagUseCase,
	articleUC *contentbiz.ArticleUseCase,
	userUC *userbiz.UserUseCase,
	permChecker authbiz.PermissionChecker,
) *AdminHandler {
	return &AdminHandler{
		jwt:            jwt,
		mediaUC:        mediaUC,
		mediaService:   mediaService,
		channelUC:      channelUC,
		tagService:     tagService,
		settingUC:      settingUC,
		categoryUC:     categoryUC,
		articleUC:      articleUC,
		userUC:         userUC,
		permChecker:    permChecker,
	}
}

func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	admin := r.Group("/admin")
	{
		// ================================
		// 1. Stats Panel
		// ================================
		stats := admin.Group("/stats")
		{
			stats.GET("/dashboard", server.WithAdmin(h.jwt, h.getDashboardStats()))
			stats.GET("/medias", server.WithAdmin(h.jwt, h.getMediaStats()))
			stats.GET("/users", server.WithAdmin(h.jwt, h.getUserStats()))
			stats.GET("/traffic", server.WithAdmin(h.jwt, h.getTrafficStats()))
		}

		// ================================
		// 1.5 Media Management (Admin - UUID only)
		// ================================
		medias := admin.Group("/medias")
		{
			medias.GET("", server.WithAdmin(h.jwt, h.adminListMedias()))
			medias.GET("/:id", server.WithAdmin(h.jwt, h.adminGetMedia()))
			medias.PUT("/:id", server.WithAdminAndPerm(h.jwt, h.permChecker, "media:write", h.adminUpdateMedia()))
			medias.DELETE("/:id", server.WithAdminAndPerm(h.jwt, h.permChecker, "media:delete", h.adminDeleteMedia()))
			medias.GET("/:id/stats", server.WithAdmin(h.jwt, h.adminGetMediaStats()))
			medias.GET("/:id/variants", server.WithAdmin(h.jwt, h.adminGetMediaVariants()))
			medias.PUT("/:id/state", server.WithAdminAndPerm(h.jwt, h.permChecker, "media:write", h.adminUpdateMediaState()))
			medias.GET("/:id/tasks", server.WithAdmin(h.jwt, h.adminGetMediaTasks()))
			medias.POST("/:id/tasks/:taskId/retry", server.WithAdminAndPerm(h.jwt, h.permChecker, "media:write", h.adminRetryMediaTask()))
		}

		// ================================
		// 2. Channel Management (Admin - UUID only)
		// ================================
		channels := admin.Group("/channels")
		{
			channels.GET("", server.WithAdmin(h.jwt, h.adminListChannels()))
			channels.GET("/:id", server.WithAdmin(h.jwt, h.adminGetChannelDetail()))
			channels.PUT("/:id", server.WithAdminAndPerm(h.jwt, h.permChecker, "media:write", h.adminUpdateChannel()))
			channels.DELETE("/:id", server.WithAdminAndPerm(h.jwt, h.permChecker, "media:delete", h.adminDeleteChannel()))
		}

		// ================================
		// 3. Encoding Management
		// ================================
		encoding := admin.Group("/encoding")
		{
			encoding.GET("/tasks", server.WithAdmin(h.jwt, h.getAllEncodingTasks()))
			encoding.GET("/status", server.WithAdmin(h.jwt, h.getEncodingStatus()))
			encoding.POST("/tasks/:taskId/retry", server.WithAdmin(h.jwt, h.retryTask()))
			encoding.POST("/retry-failed", server.WithAdmin(h.jwt, h.retryAllFailedTasks()))

			profiles := encoding.Group("/profiles")
			{
				profiles.GET("", server.WithAdmin(h.jwt, h.listEncodeProfiles()))
				profiles.POST("", server.WithAdmin(h.jwt, h.createEncodeProfile()))
				profiles.POST("/preview", server.WithAdmin(h.jwt, h.previewEncodeCommand()))
				profiles.GET("/:id", server.WithAdmin(h.jwt, h.getEncodeProfile()))
				profiles.PUT("/:id", server.WithAdmin(h.jwt, h.updateEncodeProfile()))
				profiles.DELETE("/:id", server.WithAdmin(h.jwt, h.deleteEncodeProfile()))
			}
		}

		// ================================
		// 3.5 SSE for transcoding progress (admin only)
		// ================================
		// Uses query parameter ?token=<jwt> for authentication because
		// EventSource API does not support custom headers.
		admin.GET("/medias/transcoding/events", server.WithAdmin(h.jwt, h.sseTranscodingEvents()))

		// ================================
		// 3. System Settings
		// ================================
		settings := admin.Group("/settings")
		{
			settings.GET("", server.WithAdmin(h.jwt, h.getSystemSettings()))
			settings.PUT("", server.WithAdminAndPerm(h.jwt, h.permChecker, "system:config", h.updateSystemSettings()))
		}

		// ================================
		// 4. Tag Management
		// NOTE: Tag routes are handled by AdminTagHandler to avoid duplicate registration.
		// Do NOT re-register /admin/tags here.
		// ================================

		// ================================
		// 5. Playlist Management
		// ================================
		playlists := admin.Group("/playlists")
		{
			playlists.GET("", server.WithAdmin(h.jwt, h.adminListPlaylists()))
			playlists.GET("/:id", server.WithAdmin(h.jwt, h.adminGetPlaylistDetail())) // :id = UUID
			playlists.POST("", server.WithAdmin(h.jwt, h.adminCreatePlaylist()))
			playlists.PUT("/:id", server.WithAdmin(h.jwt, h.adminUpdatePlaylist()))    // :id = UUID
			playlists.DELETE("/:id", server.WithAdmin(h.jwt, h.adminDeletePlaylist())) // :id = UUID
		}

		// ================================
		// 6. User Management
		// ================================
		users := admin.Group("/users")
		{
			users.GET("", server.WithAdmin(h.jwt, h.adminListUsers()))
			users.POST("", server.WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminCreateUser()))
			users.GET("/:id", server.WithAdmin(h.jwt, h.adminGetUser()))
			users.PUT("/:id", server.WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminUpdateUser()))
			users.DELETE("/:id", server.WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminDeleteUser()))
			users.PATCH("/:id/status", server.WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminUpdateUserStatus()))
			users.PATCH("/:id/role", server.WithAdminAndPerm(h.jwt, h.permChecker, "user:manage", h.adminUpdateUserRole()))
		}

		// ================================
		// 7. Category Management
		// ================================
		categories := admin.Group("/categories")
		{
			categories.GET("", server.WithAdmin(h.jwt, h.adminListCategories()))
			categories.GET("/:id", server.WithAdmin(h.jwt, h.adminGetCategory()))
			categories.POST("", server.WithAdmin(h.jwt, h.adminCreateCategory()))
			categories.PUT("/:id", server.WithAdmin(h.jwt, h.adminUpdateCategory()))
			categories.PATCH("/:id", server.WithAdmin(h.jwt, h.adminPatchCategory()))
			categories.DELETE("/:id", server.WithAdmin(h.jwt, h.adminDeleteCategory()))
		}

		// ================================
		// 8. Article Management
		// ================================
		articles := admin.Group("/articles")
		{
			articles.GET("", server.WithAdmin(h.jwt, h.adminListArticles()))
			articles.GET("/:id", server.WithAdmin(h.jwt, h.adminGetArticle()))
			articles.POST("", server.WithAdmin(h.jwt, h.adminCreateArticle()))
			articles.PUT("/:id", server.WithAdmin(h.jwt, h.adminUpdateArticle()))
			articles.DELETE("/:id", server.WithAdmin(h.jwt, h.adminDeleteArticle()))
			articles.PATCH("/:id/state", server.WithAdmin(h.jwt, h.adminUpdateArticleState()))
		}
	}
}

// --- Stats Panel Handlers ---

func (h *AdminHandler) getDashboardStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement dashboard stats
		server.OK(c, gin.H{
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
		server.OK(c, gin.H{
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
		server.OK(c, gin.H{
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
		server.OK(c, gin.H{
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
				server.Fail(c, server.ErrBadRequest, "Invalid status parameter")
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

		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "25")); err == nil {
			filter.PageSize = ps
		}
		// Normalize pagination parameters
		page, pageSize := repo.NormalizeHTTPPagination(filter.Page, filter.PageSize)
		filter.Page = page
		filter.PageSize = pageSize

		var mediaID *string
		if m := c.Query("media_id"); m != "" {
			mediaID = &m
		}

		result, err := h.mediaUC.ListEncodingTasksFlat(c.Request.Context(), filter, mediaID)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, result)
	}
}

func (h *AdminHandler) getEncodingStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement encoding status
		server.OK(c, gin.H{
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
			server.Fail(c, server.ErrBadRequest, "Invalid task ID")
			return
		}

		task, err := h.mediaUC.RetryTask(c.Request.Context(), taskIDStr)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, task)
	}
}

func (h *AdminHandler) retryAllFailedTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		mediaIDStr := c.Query("media_id")

		count, err := h.mediaUC.RetryAllFailedTasks(c.Request.Context(), mediaIDStr)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, gin.H{"retried_count": count})
	}
}

// --- Encode Profile Handlers ---

func (h *AdminHandler) listEncodeProfiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		profiles, err := h.mediaUC.ListEncodeProfiles(c.Request.Context())
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}
		server.OK(c, profiles)
	}
}

func (h *AdminHandler) createEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var profile mediadto.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}
		p, err := h.mediaUC.CreateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}
		server.Created(c, p)
	}
}

func (h *AdminHandler) getEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "Invalid Profile ID")
			return
		}
		p, err := h.mediaUC.GetEncodeProfile(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "Profile not found")
			return
		}
		server.OK(c, p)
	}
}

func (h *AdminHandler) updateEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "Invalid Profile ID")
			return
		}
		var profile mediadto.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}
		profile.Id = id
		p, err := h.mediaUC.UpdateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}
		server.OK(c, p)
	}
}

func (h *AdminHandler) deleteEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "Invalid Profile ID")
			return
		}
		if err := h.mediaUC.DeleteEncodeProfile(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}
		server.OK(c, nil)
	}
}

func (h *AdminHandler) previewEncodeCommand() gin.HandlerFunc {
	return func(c *gin.Context) {
		var profile mediadto.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}
		preview := h.mediaUC.GenerateCommandPreview(c.Request.Context(), &profile)
		server.OK(c, gin.H{"command": preview})
	}
}

// --- System Settings Handlers ---

func (h *AdminHandler) getSystemSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			server.Fail(c, server.ErrInternal, "settings service not available")
			return
		}

		items, err := h.settingUC.ListAll(c.Request.Context())
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		grouped := make(map[string][]*entity.Setting)
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			cat := string(item.Category)
			grouped[cat] = append(grouped[cat], masked)
		}

		server.OK(c, grouped)
	}
}

func (h *AdminHandler) updateSystemSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			server.Fail(c, server.ErrInternal, "settings service not available")
			return
		}

		var req struct {
			Settings []struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value"`
			} `json:"settings" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		var updated []*entity.Setting
		for _, item := range req.Settings {
			existing, err := h.settingUC.GetByKey(c.Request.Context(), item.Key)
			if err != nil {
				server.Fail(c, server.ErrNotFound, "setting not found: "+item.Key)
				return
			}

			if err := systemservice.ValidateSettingValue(item.Value, existing.Type); err != nil {
				server.Fail(c, server.ErrBadRequest, "invalid value for "+item.Key+": "+err.Error())
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
				server.Fail(c, server.ErrInternal, err.Error())
				return
			}
			updated = append(updated, result)
		}

		server.OK(c, updated)
	}
}

// --- Tag Management Handlers ---
// B087-R2 Fix: These handlers also use TagResponse DTO for frontend-compatible field names.
// Note: Tag routes are handled by AdminTagHandler, but these methods exist
// for backward compatibility if called from AdminHandler.

func (h *AdminHandler) listTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		// B087-R2 Fix: Support both "search" and "keyword" parameters
		search := c.Query("search")
		if search == "" {
			search = c.Query("keyword")
		}

		status := c.Query("status")
		sortBy := c.DefaultQuery("sort_by", "create_time")
		sortOrder := c.DefaultQuery("sort_order", "desc")

		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

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
			server.Fail(c, server.ErrInternal, "Failed to list tags")
			return
		}

		// B087-R2 Fix: Convert entity.Tag to TagResponse DTO
		tagResponses := ToTagResponseList(tags)

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		server.OK(c, gin.H{
			"items":       tagResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		})
	}
}

func (h *AdminHandler) getTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		tag, err := h.tagService.Get(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "Tag not found")
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(c, ToTagResponse(tag))
	}
}

func (h *AdminHandler) createTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug"` // Optional: auto-generated from name when empty
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, server.ErrBadRequest, "Invalid request")
			return
		}

		tag := &entity.Tag{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			// B087-R2 Fix: Parse frontend status string to DB enum
			Status: ParseTagStatus(req.Status),
		}

		// Auto-generate slug from name when not provided
		if req.Slug != "" {
			tag.Slug = req.Slug
		} else {
			tag.Slug = hashtag.GenerateTagSlug(req.Name)
		}

		createdTag, err := h.tagService.Create(c.Request.Context(), tag)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(c, ToTagResponse(createdTag))
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
			server.Fail(c, server.ErrBadRequest, "Invalid request")
			return
		}

		updates := &entity.Tag{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			// B087-R2 Fix: Parse frontend status string to DB enum
			Status: ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			updates.Slug = req.Slug
		}

		updatedTag, err := h.tagService.Update(c.Request.Context(), id, updates)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(c, ToTagResponse(updatedTag))
	}
}

func (h *AdminHandler) deleteTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := h.tagService.Delete(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

func (h *AdminHandler) bulkTagOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, server.ErrBadRequest, "Invalid request")
			return
		}

		if req.Action != "delete" {
			server.Fail(c, server.ErrBadRequest, "Unsupported action")
			return
		}

		count, err := h.tagService.BulkDelete(c.Request.Context(), req.IDs)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(c, gin.H{
			"success": count,
			"failed":  len(req.IDs) - count,
		})
	}
}

func (h *AdminHandler) exportTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement export functionality
		server.OK(c, gin.H{"message": "Export functionality not implemented yet"})
	}
}

func (h *AdminHandler) importTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement import functionality
		server.OK(c, gin.H{"message": "Import functionality not implemented yet"})
	}
}

// ================================
// Admin Channel Management Handlers (v3.2 - UUID only)
// ================================

func (h *AdminHandler) adminListChannels() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.channelUC.ListChannels(c.Request.Context(), page, pageSize)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.Page(c, items, int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetChannelDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "channel id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(c, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
			return
		}

		ch, err := h.channelUC.GetChannel(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "channel_not_found")
			return
		}

		server.OK(c, ch)
	}
}

func (h *AdminHandler) adminUpdateChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "channel id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(c, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
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
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		existingCh, err := h.channelUC.GetChannel(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "channel_not_found")
			return
		}

		chItem := &contentbiz.Channel{
			ID:          id,
			Title:       input.Title,
			Description: input.Description,
			BannerLogo:  input.BannerLogo,
			IsPublic:    existingCh.IsPublic,
			UserID:      existingCh.UserID,
			CreateTime:  existingCh.CreateTime,
		}

		if input.IsPublic != nil {
			chItem.IsPublic = *input.IsPublic
		}
		if input.IsDefault != nil {
			chItem.IsDefault = *input.IsDefault
		}

		updated, err := h.channelUC.UpdateChannel(c.Request.Context(), chItem, "", true)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, updated)
	}
}

func (h *AdminHandler) adminDeleteChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "channel id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(c, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
			return
		}

		err := h.channelUC.DeleteChannel(c.Request.Context(), id, "", true)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

// ================================
// Admin Playlist Management Handlers
// ================================

func (h *AdminHandler) adminListPlaylists() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.channelUC.ListPlaylists(c.Request.Context(), page, pageSize)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.Page(c, items, int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetPlaylistDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "playlist id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(c, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
			return
		}

		playlist, err := h.channelUC.GetPlaylist(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "playlist_not_found")
			return
		}

		server.OK(c, playlist)
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
			server.Fail(c, server.ErrBadRequest, err.Error())
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
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, created)
	}
}

func (h *AdminHandler) adminUpdatePlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "playlist id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(c, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
			return
		}

		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		existingPlaylist, err := h.channelUC.GetPlaylist(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "playlist_not_found")
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
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, updated)
	}
}

func (h *AdminHandler) adminDeletePlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "playlist id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(c, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
			return
		}

		err := h.channelUC.DeletePlaylist(c.Request.Context(), id, "", true)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

// ================================
// Admin User Management Handlers
// ================================

func (h *AdminHandler) adminListUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		keyword := c.Query("keyword")

		opts := &userdto.UserQueryOption{
			QueryOption: repo.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  keyword,
			},
		}

		// Filter by role if specified
		if role := c.Query("role"); role != "" {
			opts.Role = role
		}

		// Filter by status if specified
		if statusStr := c.Query("status"); statusStr != "" {
			statusMap := map[string]int32{
				"pending":   1,
				"active":    2,
				"inactive":  3,
				"suspended": 4,
				"rejected":  5,
			}
			if s, ok := statusMap[statusStr]; ok {
				opts.Status = &s
			}
		}

		users, total, err := h.userUC.ListUsers(c.Request.Context(), opts)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, &pb.ListUsersResponse{
			Items:    users,
			Total:    total,
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
	}
}

func (h *AdminHandler) adminCreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password"`                           // optional: admin can create user without password
			Nickname string `json:"nickname"`
			Role     string `json:"role"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		// If password is provided, validate minimum length
		if input.Password != "" && len(input.Password) < 6 {
			server.Fail(c, server.ErrBadRequest, "password must be at least 6 characters")
			return
		}

		// Hash password if provided; otherwise generate a random one
		var hashedPassword string
		var err error
		if input.Password != "" {
			hashedPassword, err = h.userUC.HashPassword(input.Password)
			if err != nil {
				server.Fail(c, server.ErrInternal, "failed to hash password")
				return
			}
		} else {
			// Generate a random password so the account is not left without one
			randomPwd := generateRandomPassword(12)
			hashedPassword, err = h.userUC.HashPassword(randomPwd)
			if err != nil {
				server.Fail(c, server.ErrInternal, "failed to hash password")
				return
			}
		}

		user := &types.User{
			Username: input.Username,
			Email:    input.Email,
			Nickname: input.Nickname,
		}

		created, err := h.userUC.CreateUser(c.Request.Context(), user, hashedPassword)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		// Set role if specified (default is "user")
		role := input.Role
		if role == "" {
			role = "user"
		}
		if role != "user" {
			if err := h.userUC.SetUserRole(c.Request.Context(), created.Id, role); err != nil {
				server.Fail(c, server.ErrInternal, "failed to set role: "+err.Error())
				return
			}
		}

		server.Created(c, &pb.CreateUserResponse{User: created})
	}
}

func (h *AdminHandler) adminGetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "user id is required")
			return
		}

		user, err := h.userUC.GetUser(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "user not found")
			return
		}

		server.OK(c, &pb.GetUserResponse{User: user})
	}
}

func (h *AdminHandler) adminUpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "user id is required")
			return
		}

		existing, err := h.userUC.GetUser(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "user not found")
			return
		}

		var input struct {
			Username string `json:"username"`
			Nickname string `json:"nickname"`
			Email    string `json:"email"`
			Avatar   string `json:"avatar"`
			Phone    string `json:"phone"`
			Role     string `json:"role"`
			Status   string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
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

		// Update role if specified
		if input.Role != "" {
			if err := h.userUC.SetUserRole(c.Request.Context(), id, input.Role); err != nil {
				server.Fail(c, server.ErrInternal, "failed to update role: "+err.Error())
				return
			}
		}

		// Update status if specified (convert string status to enum)
		if input.Status != "" {
			statusMap := map[string]int32{
				"pending":   1,
				"active":    2,
				"inactive":  3,
				"suspended": 4,
				"rejected":  5,
			}
			if statusCode, ok := statusMap[input.Status]; ok {
				if err := h.userUC.UpdateUserStatus(c.Request.Context(), id, int8(statusCode)); err != nil {
					server.Fail(c, server.ErrInternal, "failed to update status: "+err.Error())
					return
				}
			}
		}

		updated, err := h.userUC.UpdateUser(c.Request.Context(), existing)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, &pb.UpdateUserResponse{User: updated})
	}
}

func (h *AdminHandler) adminDeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "user id is required")
			return
		}

		if err := h.userUC.DeleteUser(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, &pb.DeleteUserResponse{Empty: &emptypb.Empty{}})
	}
}

func (h *AdminHandler) adminUpdateUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "user id is required")
			return
		}

		var input struct {
			Status int32 `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.userUC.UpdateUserStatus(c.Request.Context(), id, int8(input.Status)); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, &emptypb.Empty{})
	}
}

// mergeTags merges existing tags with new parsed tags, deduplicating
// case-insensitively. Existing tags are preserved; new tags are appended
// only if their lowercase form is not already present.
func mergeTags(existing []string, parsed []string) []string {
	seen := make(map[string]bool)
	for _, t := range existing {
		seen[strings.ToLower(t)] = true
	}
	result := make([]string, len(existing))
	copy(result, existing)
	for _, t := range parsed {
		if !seen[strings.ToLower(t)] {
			seen[strings.ToLower(t)] = true
			result = append(result, t)
		}
	}
	return result
}

// generateRandomPassword generates a cryptographically secure random password
// of the specified length using alphanumeric characters.
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback: should never happen in practice
			result[i] = charset[i%len(charset)]
			continue
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

// sseTranscodingEvents handles GET /admin/medias/transcoding/events
// SSE endpoint for real-time transcoding progress updates.
// Requires admin authentication (JWT via query parameter ?token=<jwt>).
func (h *AdminHandler) sseTranscodingEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.mediaService != nil {
			h.mediaService.SSEHandler(c.Writer, c.Request)
			return
		}
		c.Status(404)
	}
}

// ==================== Media Management Handlers ====================

// adminListMedias handles GET /admin/medias
func (h *AdminHandler) adminListMedias() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		page, pageSize = repo.NormalizePagination(page, pageSize)

		opts := &mediadto.MediaQueryOption{
			QueryOption: repo.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  c.Query("keyword"),
			},
			AdminMode: true, // Admin sees all medias regardless of state
		}

		if state := c.Query("state"); state != "" {
			opts.State = state
		}
		if mediaType := c.Query("type"); mediaType != "" {
			opts.MediaType = mediaType
		}

		items, total, err := h.mediaUC.ListMedias(c.Request.Context(), opts)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		totalPages := int32(0)
		if pageSize > 0 {
			totalPages = (total + int32(pageSize) - 1) / int32(pageSize)
		}

		server.ProtoOK(c, &mediapb.ListMediasResponse{
			Total:      total,
			Items:      items,
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalPages: totalPages,
		})
	}
}

// adminGetMedia handles GET /admin/medias/:id
func (h *AdminHandler) adminGetMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		media, err := h.mediaUC.GetByID(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "media_not_found")
			return
		}

		server.ProtoOK(c, media)
	}
}

// adminUpdateMedia handles PUT /admin/medias/:id
// Supports partial updates: only fields present in the request body are updated.
func (h *AdminHandler) adminUpdateMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		// Fetch existing media first for partial update
		existing, err := h.mediaUC.GetByID(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "media_not_found")
			return
		}

		var input struct {
			Title          *string  `json:"title"`
			Description    *string  `json:"description"`
			Thumbnail      *string  `json:"thumbnail"`
			CategoryID     *int64   `json:"category_id"`
			Tags           []string `json:"tags"`
			State          *string  `json:"state"`
			Privacy        *int32   `json:"privacy"`
			Featured       *bool    `json:"featured"`
			EnableComments *bool    `json:"enable_comments"`
			AllowDownload  *bool    `json:"allow_download"`
			Listable       *bool    `json:"listable"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		// Merge: only overwrite fields that were explicitly provided
		titleChanged := false
		descChanged := false
		if input.Title != nil {
			existing.Title = *input.Title
			titleChanged = true
		}
		if input.Description != nil {
			existing.Description = *input.Description
			descChanged = true
		}
		if input.Thumbnail != nil {
			existing.Thumbnail = *input.Thumbnail
		}
		if input.CategoryID != nil {
			existing.CategoryId = *input.CategoryID
		}
		if input.Tags != nil {
			existing.Tags = input.Tags
		}
		if input.State != nil {
			existing.State = *input.State
		}
		if input.Privacy != nil {
			existing.Privacy = types.Privacy(*input.Privacy)
		}
		if input.Featured != nil {
			existing.Featured = *input.Featured
		}
		if input.EnableComments != nil {
			existing.EnableComments = *input.EnableComments
		}
		if input.AllowDownload != nil {
			existing.AllowDownload = *input.AllowDownload
		}
		if input.Listable != nil {
			existing.Listable = *input.Listable
		}

		// Parse #hashtags from title and description when either changes.
		// Merges parsed hashtag names into existing tags (deduped).
		if titleChanged || descChanged {
			parsedTags := hashtag.ParseHashtags(existing.Title + " " + existing.Description)
			if len(parsedTags) > 0 {
				// Merge parsed tags into existing tags (case-insensitive dedup)
				existingTags := existing.Tags
				if existingTags == nil {
					existingTags = []string{}
				}
				merged := mergeTags(existingTags, parsedTags)
				existing.Tags = merged
			}
		}

		updated, err := h.mediaUC.UpdateMedia(c.Request.Context(), existing)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.ProtoOK(c, updated)
	}
}

// adminDeleteMedia handles DELETE /admin/medias/:id
func (h *AdminHandler) adminDeleteMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		if err := h.mediaUC.DeleteMedia(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, &emptypb.Empty{})
	}
}

// adminGetMediaStats handles GET /admin/medias/:id/stats
func (h *AdminHandler) adminGetMediaStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		media, err := h.mediaUC.GetByID(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "media_not_found")
			return
		}

		server.OK(c, gin.H{
			"id":              media.Id,
			"view_count":      media.ViewCount,
			"like_count":      media.LikeCount,
			"dislike_count":   media.DislikeCount,
			"comment_count":   media.CommentCount,
			"favorite_count":  media.FavoriteCount,
			"encoding_status": media.EncodingStatus,
		})
	}
}

// adminGetMediaVariants handles GET /admin/medias/:id/variants
func (h *AdminHandler) adminGetMediaVariants() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		summary, err := h.mediaUC.GetMediaVariantsByUUID(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, summary)
	}
}

// adminUpdateMediaState handles PUT /admin/medias/:id/state
func (h *AdminHandler) adminUpdateMediaState() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		var input struct {
			State   string `json:"state" binding:"required"`
			Comment string `json:"comment"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.mediaUC.UpdateMediaState(c.Request.Context(), id, input.State); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		// Fetch updated media to return current state
		updated, err := h.mediaUC.GetByID(c.Request.Context(), id)
		if err != nil {
			server.OK(c, gin.H{"id": id, "state": input.State})
			return
		}

		server.OK(c, gin.H{
			"id":          updated.Id,
			"state":       updated.State,
			"update_time": updated.UpdateTime,
		})
	}
}

// adminGetMediaTasks handles GET /admin/medias/:id/tasks
func (h *AdminHandler) adminGetMediaTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "media id is required")
			return
		}

		tasks, err := h.mediaUC.ListEncodingTasks(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, gin.H{
			"items": tasks,
			"total": len(tasks),
		})
	}
}

// adminRetryMediaTask handles POST /admin/medias/:id/tasks/:taskId/retry
func (h *AdminHandler) adminRetryMediaTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("taskId")
		if taskID == "" {
			server.Fail(c, server.ErrBadRequest, "task id is required")
			return
		}

		_, err := h.mediaUC.RetryTask(c.Request.Context(), taskID)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, gin.H{"message": "retry initiated"})
	}
}

// ================================
// Admin Article Management Handlers
// ================================

func (h *AdminHandler) adminListArticles() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

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
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.Page(c, articles, int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		article, err := h.articleUC.Get(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "article not found")
			return
		}

		server.OK(c, article)
	}
}

func (h *AdminHandler) adminCreateArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title      string   `json:"title" binding:"required"`
			Slug       string   `json:"slug"`
			Content    string   `json:"content"`
			Summary    string   `json:"summary"`
			CategoryID int64    `json:"category_id"`
			MediaID    string   `json:"media_id"`
			Thumbnail  string   `json:"thumbnail"`
			State      string   `json:"state"`
			Tags       []string `json:"tags"`
			Featured   bool     `json:"featured"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		userID := ""
		if claims, exists := c.Get("claims"); exists {
			if cl, ok := claims.(*auth.Claims); ok {
				userID = cl.GetUserID()
			}
		}

		state := input.State
		if state == "" {
			state = "draft"
		}

		article := &contentbiz.Article{
			Title:      input.Title,
			Slug:       input.Slug,
			Content:    input.Content,
			Summary:    input.Summary,
			UserID:     userID,
			CategoryID: input.CategoryID,
			MediaID:    input.MediaID,
			Thumbnail:  input.Thumbnail,
			State:      state,
			Tags:       input.Tags,
			Featured:   input.Featured,
		}

		created, err := h.articleUC.Create(c.Request.Context(), article)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, created)
	}
}

func (h *AdminHandler) adminUpdateArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		var input struct {
			Title      string   `json:"title"`
			Slug       string   `json:"slug"`
			Content    string   `json:"content"`
			Summary    string   `json:"summary"`
			CategoryID int64    `json:"category_id"`
			MediaID    string   `json:"media_id"`
			Thumbnail  string   `json:"thumbnail"`
			State      string   `json:"state"`
			Tags       []string `json:"tags"`
			Featured   bool     `json:"featured"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		existing, err := h.articleUC.Get(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "article not found")
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
		if input.MediaID != "" {
			existing.MediaID = input.MediaID
		}
		existing.Thumbnail = input.Thumbnail // Allow empty string to clear
		if input.State != "" {
			existing.State = input.State
		}
		if input.Tags != nil {
			existing.Tags = input.Tags
		}
		existing.Featured = input.Featured

		updated, err := h.articleUC.Update(c.Request.Context(), existing)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, updated)
	}
}

func (h *AdminHandler) adminDeleteArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		if err := h.articleUC.Delete(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

func (h *AdminHandler) adminUpdateArticleState() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		var input struct {
			State string `json:"state" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.articleUC.UpdateState(c.Request.Context(), id, input.State); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

// ================================
// Admin Category Management Handlers
// ================================

func (h *AdminHandler) adminListCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := h.categoryUC.ListCategories(c.Request.Context())
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		total := len(categories)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		server.Page(c, categories[start:end], int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "invalid category id")
			return
		}

		cat, err := h.categoryUC.GetCategory(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "category not found")
			return
		}

		server.OK(c, cat)
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
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		cat := &contentbiz.Category{
			Name:        input.Name,
			Slug:        input.Slug,
			Description: input.Description,
		}

		created, err := h.categoryUC.CreateCategory(c.Request.Context(), cat)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, created)
	}
}

func (h *AdminHandler) adminUpdateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "invalid category id")
			return
		}

		var input contentbiz.UpdateCategoryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(c.Request.Context(), id, &input)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, updated)
	}
}

func (h *AdminHandler) adminPatchCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "invalid category id")
			return
		}

		var input contentbiz.UpdateCategoryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(c.Request.Context(), id, &input)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, updated)
	}
}

func (h *AdminHandler) adminDeleteCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, "invalid category id")
			return
		}

		if err := h.categoryUC.DeleteCategory(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

func (h *AdminHandler) adminUpdateUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "user id is required")
			return
		}

		var input struct {
			Role string `json:"role" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		validRoles := map[string]bool{"user": true, "admin": true, "editor": true}
		if !validRoles[input.Role] {
			server.Fail(c, server.ErrBadRequest, "invalid role, must be one of: user, admin, editor")
			return
		}

		if err := h.userUC.SetUserRole(c.Request.Context(), id, input.Role); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, &emptypb.Empty{})
	}
}
