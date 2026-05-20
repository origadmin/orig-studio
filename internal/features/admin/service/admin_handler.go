/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides HTTP handlers for admin operations.
package service

import (
	"crypto/rand"
	"net/http"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "origadmin/application/origstudio/api/gen/v1/user"
	mediapb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/dal/enums"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/server/validation"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	"origadmin/application/origstudio/internal/features/admin/dto"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	mediabiz "origadmin/application/origstudio/internal/features/media/biz"
	mediadto "origadmin/application/origstudio/internal/features/media/dto"
	mediaservice "origadmin/application/origstudio/internal/features/media/service"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemdal "origadmin/application/origstudio/internal/features/system/dal"
	systemdto "origadmin/application/origstudio/internal/features/system/dto"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
	userdto "origadmin/application/origstudio/internal/features/user/dto"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
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
	statsRepo      *systemdal.StatsRepo
	appVersion     string
	dbDialect      string
	startTime      time.Time
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
	statsRepo *systemdal.StatsRepo,
	adminCfg *AdminConfig,
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
		statsRepo:      statsRepo,
		appVersion:     adminCfg.AppVersion,
		dbDialect:      adminCfg.DBDialect,
		startTime:      time.Now(),
	}
}

func (h *AdminHandler) RegisterRoutes(r http2.Router) {
	admin := r.Group("/admin")
	{
		// ================================
		// 1. Stats Panel
		// ================================
		stats := admin.Group("/stats")
		{
			stats.GET("/dashboard", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getDashboardStats())))
			stats.GET("/medias", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getMediaStats())))
			stats.GET("/users", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getUserStats())))
			stats.GET("/traffic", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getTrafficStats())))
		}

		// ================================
		// 1.5 Media Management (Admin - UUID only)
		// ================================
		medias := admin.Group("/medias")
		{
			medias.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminListMedias())))
			medias.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetMedia())))
			medias.PUT("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", server.HTTPToHandlerFunc(h.adminUpdateMedia())))
			medias.DELETE("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:delete", server.HTTPToHandlerFunc(h.adminDeleteMedia())))
			medias.GET("/:id/stats", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetMediaStats())))
			medias.GET("/:id/variants", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetMediaVariants())))
			medias.PUT("/:id/state", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", server.HTTPToHandlerFunc(h.adminUpdateMediaState())))
			medias.GET("/:id/tasks", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetMediaTasks())))
			medias.POST("/:id/tasks/:taskId/retry", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", server.HTTPToHandlerFunc(h.adminRetryMediaTask())))
		}

		// ================================
		// 2. Channel Management (Admin - UUID only)
		// ================================
		channels := admin.Group("/channels")
		{
			channels.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminListChannels())))
			channels.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetChannelDetail())))
			channels.PUT("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", server.HTTPToHandlerFunc(h.adminUpdateChannel())))
			channels.DELETE("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:delete", server.HTTPToHandlerFunc(h.adminDeleteChannel())))
		}

		// ================================
		// 3. Encoding Management
		// ================================
		encoding := admin.Group("/encoding")
		{
			encoding.GET("/tasks", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getAllEncodingTasks())))
			encoding.GET("/status", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getEncodingStatus())))
			encoding.POST("/tasks/:taskId/retry", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.retryTask())))
			encoding.POST("/retry-failed", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.retryAllFailedTasks())))

			profiles := encoding.Group("/profiles")
			{
				profiles.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.listEncodeProfiles())))
				profiles.POST("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.createEncodeProfile())))
				profiles.POST("/preview", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.previewEncodeCommand())))
				profiles.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getEncodeProfile())))
				profiles.PUT("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.updateEncodeProfile())))
				profiles.DELETE("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.deleteEncodeProfile())))
			}
		}

		// ================================
		// 3.5 SSE for transcoding progress (admin only)
		// ================================
		// Uses query parameter ?token=<jwt> for authentication because
		// EventSource API does not support custom headers.
		admin.GET("/medias/transcoding/events", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.sseTranscodingEvents())))

		// ================================
		// 3. System Settings
		// ================================
		settings := admin.Group("/settings")
		{
			settings.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getSystemSettings())))
			settings.GET("/info", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.getSystemInfo())))
			settings.PUT("", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "system:config", server.HTTPToHandlerFunc(h.updateSystemSettings())))
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
			playlists.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminListPlaylists())))
			playlists.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetPlaylistDetail()))) // :id = UUID
			playlists.POST("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminCreatePlaylist())))
			playlists.PUT("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminUpdatePlaylist())))    // :id = UUID
			playlists.DELETE("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminDeletePlaylist()))) // :id = UUID
		}

		// ================================
		// 6. User Management
		// ================================
		users := admin.Group("/users")
		{
			users.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminListUsers())))
			users.POST("", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", server.HTTPToHandlerFunc(h.adminCreateUser())))
			users.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetUser())))
			users.PUT("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", server.HTTPToHandlerFunc(h.adminUpdateUser())))
			users.DELETE("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", server.HTTPToHandlerFunc(h.adminDeleteUser())))
			users.PATCH("/:id/status", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", server.HTTPToHandlerFunc(h.adminUpdateUserStatus())))
			users.PATCH("/:id/role", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", server.HTTPToHandlerFunc(h.adminUpdateUserRole())))
		}

		// ================================
		// 7. Category Management
		// ================================
		categories := admin.Group("/categories")
		{
			categories.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminListCategories())))
			categories.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetCategory())))
			categories.POST("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminCreateCategory())))
			categories.PUT("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminUpdateCategory())))
			categories.PATCH("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminPatchCategory())))
			categories.DELETE("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminDeleteCategory())))
		}

		// ================================
		// 8. Article Management
		// ================================
		articles := admin.Group("/articles")
		{
			articles.GET("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminListArticles())))
			articles.GET("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminGetArticle())))
			articles.POST("", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminCreateArticle())))
			articles.PUT("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminUpdateArticle())))
			articles.DELETE("/:id", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminDeleteArticle())))
			articles.PATCH("/:id/state", server.WithAdminCtx(h.jwt, server.HTTPToHandlerFunc(h.adminUpdateArticleState())))
		}
	}
}

// --- Stats Panel Handlers ---

func (h *AdminHandler) getDashboardStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)

		if h.statsRepo == nil {
			server.Fail(gc, server.ErrInternal, "stats service not available")
			return
		}

		stats, err := h.statsRepo.GetExtendedDashboardStats(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"total_users":           stats.TotalUsers,
			"total_media":           stats.TotalMedia,
			"total_views":           stats.TotalViews,
			"total_comments":        stats.TotalComments,
			"total_subscribers":     stats.TotalSubscribers,
			"total_revenue":         0,
			"active_users":          0,
			"new_users_today":       stats.NewUsersToday,
			"new_media_today":       stats.NewMediaToday,
			"new_views_today":       0,
			"new_comments_today":    stats.NewCommentsToday,
			"new_subscribers_today": stats.NewSubsToday,
			"media_by_type":         stats.MediaByType,
			"users_by_role":         stats.UsersByRole,
			"views_by_date":         []interface{}{},
			"media_by_date":         []interface{}{},
			"top_categories":        []interface{}{},
			"top_creators":          []interface{}{},
			"top_media":             []interface{}{},
		})
	}
}

func (h *AdminHandler) getMediaStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement media stats
		server.OK(gc, gin.H{
			"total_media":      100,
			"total_storage":    "10GB",
			"video_count":      60,
			"image_count":      30,
			"audio_count":      10,
			"pending_encoding": 5,
		})
	}
}

func (h *AdminHandler) getUserStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement user stats
		server.OK(gc, gin.H{
			"total_users":     50,
			"active_users":    30,
			"new_users_today": 2,
			"admin_count":     2,
			"user_growth":     "10%",
		})
	}
}

func (h *AdminHandler) getTrafficStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement traffic stats
		server.OK(gc, gin.H{
			"total_views":     10000,
			"today_views":     1000,
			"total_bandwidth": "100GB",
			"today_bandwidth": "10GB",
			"top_media":       []interface{}{},
		})
	}
}

// --- Encoding Management Handlers ---

func (h *AdminHandler) getAllEncodingTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		status := gc.Query("status")

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
				server.Fail(gc, server.ErrBadRequest, "Invalid status parameter")
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
			ProfileFilter: gc.Query("profile"),
			ChunkFilter:   gc.Query("chunk"),
			SearchQuery:   gc.Query("search"),
		}

		if os := gc.Query("only_stats"); os == "true" {
			filter.OnlyStats = true
		}

		if p, err := strconv.Atoi(gc.DefaultQuery("page", "1")); err == nil {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(gc.DefaultQuery("page_size", "25")); err == nil {
			filter.PageSize = ps
		}
		// Normalize pagination parameters
		page, pageSize := repotypes.NormalizeHTTPPagination(filter.Page, filter.PageSize)
		filter.Page = page
		filter.PageSize = pageSize

		var mediaID *string
		if m := gc.Query("media_id"); m != "" {
			mediaID = &m
		}

		result, err := h.mediaUC.ListEncodingTasksFlat(r.Context(), filter, mediaID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, result)
	}
}

func (h *AdminHandler) getEncodingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement encoding status
		server.OK(gc, gin.H{
			"total_tasks":  100,
			"pending":      5,
			"processing":   3,
			"completed":    80,
			"failed":       12,
			"success_rate": "80%",
		})
	}
}

func (h *AdminHandler) retryTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		taskIDStr := gc.Param("taskId")
		if taskIDStr == "" {
			server.Fail(gc, server.ErrBadRequest, "Invalid task ID")
			return
		}

		task, err := h.mediaUC.RetryTask(r.Context(), taskIDStr)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, task)
	}
}

func (h *AdminHandler) retryAllFailedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		mediaIDStr := gc.Query("media_id")

		count, err := h.mediaUC.RetryAllFailedTasks(r.Context(), mediaIDStr)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"retried_count": count})
	}
}

// --- Encode Profile Handlers ---

func (h *AdminHandler) listEncodeProfiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		profiles, err := h.mediaUC.ListEncodeProfiles(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, gin.H{"profiles": profiles})
	}
}

func (h *AdminHandler) createEncodeProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var profile mediadto.EncodeProfile
		if err := gc.ShouldBindJSON(&profile); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}
		p, err := h.mediaUC.CreateEncodeProfile(r.Context(), &profile)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.Created(gc, gin.H{"profile": p})
	}
}

func (h *AdminHandler) getEncodeProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "Invalid Profile ID")
			return
		}
		p, err := h.mediaUC.GetEncodeProfile(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "Profile not found")
			return
		}
		server.OK(gc, gin.H{"profile": p})
	}
}

func (h *AdminHandler) updateEncodeProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "Invalid Profile ID")
			return
		}
		var profile mediadto.EncodeProfile
		if err := gc.ShouldBindJSON(&profile); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}
		profile.Id = id
		p, err := h.mediaUC.UpdateEncodeProfile(r.Context(), &profile)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, gin.H{"profile": p})
	}
}

func (h *AdminHandler) deleteEncodeProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "Invalid Profile ID")
			return
		}
		if err := h.mediaUC.DeleteEncodeProfile(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, nil)
	}
}

func (h *AdminHandler) previewEncodeCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var profile mediadto.EncodeProfile
		if err := gc.ShouldBindJSON(&profile); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}
		preview := h.mediaUC.GenerateCommandPreview(r.Context(), &profile)
		server.OK(gc, gin.H{"command": preview})
	}
}

// --- System Settings Handlers ---

func (h *AdminHandler) getSystemSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		items, err := h.settingUC.ListAll(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		grouped := make(map[string][]*systemdto.SettingDTO)
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			cat := string(item.Category)
			grouped[cat] = append(grouped[cat], masked)
		}

		server.OK(gc, grouped)
	}
}

// getSystemInfo returns runtime system information for the admin settings page.
// Response format matches the frontend SystemInfo interface.
//
// Future extensibility (multi-instance): The response can be extended to include
// an "instances" array for distributed deployments, while keeping backward
// compatibility by retaining the top-level fields as the "local" instance's data.
func (h *AdminHandler) getSystemInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Uptime calculation
		uptime := time.Since(h.startTime)
		uptimeStr := formatDuration(uptime)

		// Database display name
		dbName := h.dbDialect
		switch dbName {
		case "sqlite3":
			dbName = "SQLite"
		case "postgres":
			dbName = "PostgreSQL"
		}

		// Memory info
		totalMem := m.Sys
		usedMem := m.Alloc

		// Memory usage percentage
		memUsagePercent := float64(0)
		if totalMem > 0 {
			memUsagePercent = float64(usedMem) / float64(totalMem) * 100
		}

		info := gin.H{
			"version":      h.appVersion,
			"goVersion":    runtime.Version(),
			"database":     dbName,
			"os":           runtime.GOOS + "/" + runtime.GOARCH,
			"uptime":       uptimeStr,
			"totalMemory":  formatBytes(totalMem),
			"usedMemory":   formatBytes(usedMem),
			"cpuUsage":     "-", // CPU usage requires cgroup/OS-specific APIs; placeholder for now
			"memoryUsage":  memUsagePercent,
			"numCPU":       runtime.NumCPU(),
			"numGoroutine": runtime.NumGoroutine(),
		}

		server.OK(gc, info)
	}
}

func (h *AdminHandler) updateSystemSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		var req struct {
			Settings []struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value"`
			} `json:"settings" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		var updated []*systemdto.SettingDTO
		for _, item := range req.Settings {
			existing, err := h.settingUC.GetByKey(r.Context(), item.Key)
			if err != nil {
				defaults := systemdal.DefaultSettings()
				var defaultSetting *systemdto.SettingDTO
				for _, d := range defaults {
					if d.Key == item.Key {
						defaultSetting = d
						break
					}
				}
				if defaultSetting == nil {
					server.Fail(gc, server.ErrBadRequest, "unknown setting key: "+item.Key)
					return
				}
				existing = defaultSetting
			}

			if err := systemservice.ValidateSettingValue(item.Value, existing.Type); err != nil {
				server.Fail(gc, server.ErrBadRequest, "invalid value for "+item.Key+": "+err.Error())
				return
			}

			s := &systemdto.SettingDTO{
				Key:           existing.Key,
				Value:         item.Value,
				Type:          existing.Type,
				Category:      existing.Category,
				Description:   existing.Description,
				IsSensitive:   existing.IsSensitive,
				FallbackValue: existing.FallbackValue,
				IsBuiltin:     existing.IsBuiltin,
			}
			result, err := h.settingUC.Upsert(r.Context(), s)
			if err != nil {
				server.Fail(gc, server.ErrInternal, err.Error())
				return
			}
			updated = append(updated, result)
		}

		server.OK(gc, updated)
	}
}

// --- Tag Management Handlers ---
// B087-R2 Fix: These handlers also use TagResponse DTO for frontend-compatible field names.
// Note: Tag routes are handled by AdminTagHandler, but these methods exist
// for backward compatibility if called from AdminHandler.

func (h *AdminHandler) listTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// Parse query parameters
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))

		// B087-R2 Fix: Support both "search" and "keyword" parameters
		search := gc.Query("search")
		if search == "" {
			search = gc.Query("keyword")
		}

		status := gc.Query("status")
		sortBy := gc.DefaultQuery("sort_by", "create_time")
		sortOrder := gc.DefaultQuery("sort_order", "desc")

		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		// Get tags
		tags, total, err := h.tagService.List(
			r.Context(),
			page,
			pageSize,
			search,
			status,
			sortBy,
			sortOrder,
		)
		if err != nil {
			server.Fail(gc, server.ErrInternal, "Failed to list tags")
			return
		}

		// B087-R2 Fix: Convert entity.Tag to TagResponse DTO
		tagResponses := ToTagResponseList(tags)

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		server.OK(gc, gin.H{
			"items":       tagResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		})
	}
}

func (h *AdminHandler) getTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		tag, err := h.tagService.Get(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "Tag not found")
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(gc, ToTagResponse(tag))
	}
}

func (h *AdminHandler) createTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug"` // Optional: auto-generated from name when empty
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, "Invalid request")
			return
		}

		tag := &dto.TagDTO{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			Status:      ParseTagStatus(req.Status),
		}

		// Auto-generate slug from name when not provided
		if req.Slug != "" {
			tag.Slug = req.Slug
		} else {
			tag.Slug = hashtag.GenerateTagSlug(req.Name)
		}

		createdTag, err := h.tagService.Create(r.Context(), tag)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(gc, ToTagResponse(createdTag))
	}
}

func (h *AdminHandler) updateTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		var req struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, "Invalid request")
			return
		}

		updates := &dto.TagDTO{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			Status:      ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			updates.Slug = req.Slug
		}

		updatedTag, err := h.tagService.Update(r.Context(), id, updates)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(gc, ToTagResponse(updatedTag))
	}
}

func (h *AdminHandler) deleteTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		if err := h.tagService.Delete(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(gc, nil)
	}
}

func (h *AdminHandler) bulkTagOperation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, "Invalid request")
			return
		}

		if req.Action != "delete" {
			server.Fail(gc, server.ErrBadRequest, "Unsupported action")
			return
		}

		count, err := h.tagService.BulkDelete(r.Context(), req.IDs)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"success": count,
			"failed":  len(req.IDs) - count,
		})
	}
}

func (h *AdminHandler) exportTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement export functionality
		server.OK(gc, gin.H{"message": "Export functionality not implemented yet"})
	}
}

func (h *AdminHandler) importTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement import functionality
		server.OK(gc, gin.H{"message": "Import functionality not implemented yet"})
	}
}

// ================================
// Admin Channel Management Handlers (v3.2 - UUID only)
// ================================

func (h *AdminHandler) adminListChannels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.channelUC.ListChannels(r.Context(), page, pageSize)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.Page(gc, items, int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetChannelDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "channel id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(gc, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
			return
		}

		ch, err := h.channelUC.GetChannel(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "channel_not_found")
			return
		}

		server.OK(gc, ch)
	}
}

func (h *AdminHandler) adminUpdateChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "channel id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(gc, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
			return
		}

		var input struct {
			Name        *string  `json:"name"`
			Title       *string  `json:"title"`
			Description *string  `json:"description"`
			Avatar      *string  `json:"avatar"`
			Banner      *string  `json:"banner"`
			BannerLogo  *string  `json:"banner_logo"`
			Privacy     *string  `json:"privacy"`
			Status      *string  `json:"status"`
			Tags        []string `json:"tags"`
			CategoryID  *int64   `json:"category_id"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		existingCh, err := h.channelUC.GetChannel(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "channel_not_found")
			return
		}

		chItem := &contentbiz.Channel{
			ID:              id,
			Name:            existingCh.Name,
			Title:           existingCh.Title,
			Slug:            existingCh.Slug,
			Handle:          existingCh.Handle,
			Description:     existingCh.Description,
			Avatar:          existingCh.Avatar,
			Banner:          existingCh.Banner,
			BannerLogo:      existingCh.BannerLogo,
			ShortToken:      existingCh.ShortToken,
			Status:          existingCh.Status,
			Privacy:         existingCh.Privacy,
			IsVerified:      existingCh.IsVerified,
			Tags:            existingCh.Tags,
			CategoryID:      existingCh.CategoryID,
			SubscriberCount: existingCh.SubscriberCount,
			MediaCount:      existingCh.MediaCount,
			ArticleCount:    existingCh.ArticleCount,
			TotalViews:      existingCh.TotalViews,
			Links:           existingCh.Links,
			UserID:          existingCh.UserID,
			CreateTime:      existingCh.CreateTime,
			UpdateTime:      existingCh.UpdateTime,
		}

		if input.Name != nil {
			chItem.Name = *input.Name
		}
		if input.Title != nil {
			chItem.Title = *input.Title
		}
		if input.Description != nil {
			chItem.Description = *input.Description
		}
		if input.Avatar != nil {
			chItem.Avatar = *input.Avatar
		}
		if input.Banner != nil {
			chItem.Banner = *input.Banner
		}
		if input.BannerLogo != nil {
			chItem.BannerLogo = *input.BannerLogo
		}
		if input.Privacy != nil {
			chItem.Privacy = *input.Privacy
		}
		if input.Status != nil {
			chItem.Status = *input.Status
		}
		if input.Tags != nil {
			chItem.Tags = input.Tags
		}
		if input.CategoryID != nil {
			chItem.CategoryID = input.CategoryID
		}

		updated, err := h.channelUC.UpdateChannel(r.Context(), chItem, "", true)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, updated)
	}
}

func (h *AdminHandler) adminDeleteChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "channel id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(gc, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
			return
		}

		err := h.channelUC.DeleteChannel(r.Context(), id, "", true)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, nil)
	}
}

// ================================
// Admin Playlist Management Handlers
// ================================

func (h *AdminHandler) adminListPlaylists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.channelUC.ListPlaylists(r.Context(), page, pageSize)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.Page(gc, items, int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetPlaylistDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "playlist id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(gc, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
			return
		}

		playlist, err := h.channelUC.GetPlaylist(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "playlist_not_found")
			return
		}

		server.OK(gc, gin.H{"playlist": playlist})
	}
}

func (h *AdminHandler) adminCreatePlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var input struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			UserID      string `json:"user_id" binding:"required"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
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

		created, err := h.channelUC.CreatePlaylist(r.Context(), playlist)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"playlist": created})
	}
}

func (h *AdminHandler) adminUpdatePlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "playlist id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(gc, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
			return
		}

		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		existingPlaylist, err := h.channelUC.GetPlaylist(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "playlist_not_found")
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

		updated, err := h.channelUC.UpdatePlaylist(r.Context(), playlistItem, "", true)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"playlist": updated})
	}
}

func (h *AdminHandler) adminDeletePlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "playlist id is required")
			return
		}

		if !validation.IsValidUUID(id) {
			server.Fail(gc, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
			return
		}

		err := h.channelUC.DeletePlaylist(r.Context(), id, "", true)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, nil)
	}
}

// ================================
// Admin User Management Handlers
// ================================

func (h *AdminHandler) adminListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		keyword := gc.Query("keyword")

		opts := &userdto.UserQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  keyword,
			},
		}

		// Filter by role if specified
		if role := gc.Query("role"); role != "" {
			opts.Role = role
		}

		// Filter by status if specified
		if statusStr := gc.Query("status"); statusStr != "" {
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

		users, total, err := h.userUC.ListUsers(r.Context(), opts)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &pb.ListUsersResponse{
			Items:    users,
			Total:    total,
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
	}
}

func (h *AdminHandler) adminCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var input struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password"`                           // optional: admin can create user without password
			Nickname string `json:"nickname"`
			Role     string `json:"role"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		// If password is provided, validate minimum length
		if input.Password != "" && len(input.Password) < 6 {
			server.Fail(gc, server.ErrBadRequest, "password must be at least 6 characters")
			return
		}

		// Hash password if provided; otherwise generate a random one
		var hashedPassword string
		var err error
		if input.Password != "" {
			hashedPassword, err = h.userUC.HashPassword(input.Password)
			if err != nil {
				server.Fail(gc, server.ErrInternal, "failed to hash password")
				return
			}
		} else {
			// Generate a random password so the account is not left without one
			randomPwd := generateRandomPassword(12)
			hashedPassword, err = h.userUC.HashPassword(randomPwd)
			if err != nil {
				server.Fail(gc, server.ErrInternal, "failed to hash password")
				return
			}
		}

		user := &types.User{
			Username: input.Username,
			Email:    input.Email,
			Nickname: input.Nickname,
		}

		created, err := h.userUC.CreateUser(r.Context(), user, hashedPassword)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		// Set role if specified (default is "user")
		role := input.Role
		if role == "" {
			role = "user"
		}
		if role != "user" {
			if err := h.userUC.SetUserRole(r.Context(), created.Id, role); err != nil {
				server.Fail(gc, server.ErrInternal, "failed to set role: "+err.Error())
				return
			}
		}

		server.Created(gc, &pb.CreateUserResponse{User: created})
	}
}

func (h *AdminHandler) adminGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "user id is required")
			return
		}

		user, err := h.userUC.GetUser(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "user not found")
			return
		}

		server.OK(gc, &pb.GetUserResponse{User: user})
	}
}

func (h *AdminHandler) adminUpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "user id is required")
			return
		}

		existing, err := h.userUC.GetUser(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "user not found")
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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
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
			if err := h.userUC.SetUserRole(r.Context(), id, input.Role); err != nil {
				server.Fail(gc, server.ErrInternal, "failed to update role: "+err.Error())
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
				if err := h.userUC.UpdateUserStatus(r.Context(), id, int8(statusCode)); err != nil {
					server.Fail(gc, server.ErrInternal, "failed to update status: "+err.Error())
					return
				}
			}
		}

		updated, err := h.userUC.UpdateUser(r.Context(), existing)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &pb.UpdateUserResponse{User: updated})
	}
}

func (h *AdminHandler) adminDeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "user id is required")
			return
		}

		if err := h.userUC.DeleteUser(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &pb.DeleteUserResponse{Empty: &emptypb.Empty{}})
	}
}

func (h *AdminHandler) adminUpdateUserStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "user id is required")
			return
		}

		var input struct {
			Status int32 `json:"status" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.userUC.UpdateUserStatus(r.Context(), id, int8(input.Status)); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &emptypb.Empty{})
	}
}

// formatDuration formats a time.Duration into a human-readable string.
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}
	return strings.Join(parts, " ")
}

// formatBytes formats a byte count into a human-readable string.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
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
func (h *AdminHandler) sseTranscodingEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.mediaService != nil {
			h.mediaService.SSEHandler(gc.Writer, gc.Request)
			return
		}
		gc.Status(404)
	}
}

// ==================== Media Management Handlers ====================

// adminListMedias handles GET /admin/medias
func (h *AdminHandler) adminListMedias() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		page, pageSize = repotypes.NormalizePagination(page, pageSize)

		opts := &mediadto.MediaQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  gc.Query("keyword"),
			},
			AdminMode: true, // Admin sees all medias regardless of state
		}

		if state := gc.Query("state"); state != "" {
			opts.State = state
		}
		if mediaType := gc.Query("type"); mediaType != "" {
			opts.MediaType = mediaType
		}
		if tagsStr := gc.Query("tags"); tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			opts.Tags = tags
		}

		items, total, err := h.mediaUC.ListMedias(r.Context(), opts)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		totalPages := int32(0)
		if pageSize > 0 {
			totalPages = (total + int32(pageSize) - 1) / int32(pageSize)
		}

		server.ProtoOK(gc, &mediapb.ListMediasResponse{
			Total:      total,
			Items:      items,
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalPages: totalPages,
		})
	}
}

// adminGetMedia handles GET /admin/medias/:id
func (h *AdminHandler) adminGetMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		media, err := h.mediaUC.GetByID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "media_not_found")
			return
		}

		server.ProtoOK(gc, media)
	}
}

// adminUpdateMedia handles PUT /admin/medias/:id
// Supports partial updates: only fields present in the request body are updated.
func (h *AdminHandler) adminUpdateMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		// Fetch existing media first for partial update
		existing, err := h.mediaUC.GetByID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "media_not_found")
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
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
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

		updated, err := h.mediaUC.UpdateMedia(r.Context(), existing)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.ProtoOK(gc, updated)
	}
}

// adminDeleteMedia handles DELETE /admin/medias/:id
func (h *AdminHandler) adminDeleteMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		if err := h.mediaUC.DeleteMedia(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &emptypb.Empty{})
	}
}

// adminGetMediaStats handles GET /admin/medias/:id/stats
func (h *AdminHandler) adminGetMediaStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		media, err := h.mediaUC.GetByID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "media_not_found")
			return
		}

		server.OK(gc, gin.H{
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
func (h *AdminHandler) adminGetMediaVariants() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		summary, err := h.mediaUC.GetMediaVariantsByUUID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, summary)
	}
}

// adminUpdateMediaState handles PUT /admin/medias/:id/state
func (h *AdminHandler) adminUpdateMediaState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		var input struct {
			State   string `json:"state" binding:"required"`
			Comment string `json:"comment"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.mediaUC.UpdateMediaState(r.Context(), id, input.State); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		// Fetch updated media to return current state
		updated, err := h.mediaUC.GetByID(r.Context(), id)
		if err != nil {
			server.OK(gc, gin.H{"id": id, "state": input.State})
			return
		}

		server.OK(gc, gin.H{
			"id":          updated.Id,
			"state":       updated.State,
			"update_time": updated.UpdateTime,
		})
	}
}

// adminGetMediaTasks handles GET /admin/medias/:id/tasks
func (h *AdminHandler) adminGetMediaTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media id is required")
			return
		}

		tasks, err := h.mediaUC.ListEncodingTasks(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"items": tasks,
			"total": len(tasks),
		})
	}
}

// adminRetryMediaTask handles POST /admin/medias/:id/tasks/:taskId/retry
func (h *AdminHandler) adminRetryMediaTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		taskID := gc.Param("taskId")
		if taskID == "" {
			server.Fail(gc, server.ErrBadRequest, "task id is required")
			return
		}

		_, err := h.mediaUC.RetryTask(r.Context(), taskID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"message": "retry initiated"})
	}
}

// ================================
// Admin Article Management Handlers
// ================================

func (h *AdminHandler) adminListArticles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))

		req := &types.ListArticlesRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		}

		if v := gc.Query("state"); v != "" {
			req.State = v
		}
		if v := gc.Query("category_id"); v != "" {
			if catID, err := strconv.ParseInt(v, 10, 64); err == nil {
				req.CategoryId = catID
			}
		}
		if v := gc.Query("keyword"); v != "" {
			req.Keyword = v
		}

		resp, err := h.articleUC.List(r.Context(), req)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, resp)
	}
}

func (h *AdminHandler) adminGetArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "article id is required")
			return
		}

		article, err := h.articleUC.Get(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "article not found")
			return
		}

		server.OK(gc, article)
	}
}

func (h *AdminHandler) adminCreateArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		userID := ""
		if claims, exists := gc.Get("claims"); exists {
			if cl, ok := claims.(*auth.Claims); ok {
				userID = cl.GetUserID()
			}
		}

		state := input.State
		if state == "" {
			state = "draft"
		}

		slug := input.Slug
		if slug == "" {
			slug = hashtag.GenerateTagSlug(input.Title)
		}

		article := &types.Article{
			Title:      input.Title,
			Slug:       slug,
			Content:    input.Content,
			Summary:    input.Summary,
			UserId:     userID,
			CategoryId: input.CategoryID,
			MediaId:    input.MediaID,
			Thumbnail:  input.Thumbnail,
			State:      state,
			Tags:       input.Tags,
			Featured:   input.Featured,
		}

		created, err := h.articleUC.Create(r.Context(), article)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, created)
	}
}

func (h *AdminHandler) adminUpdateArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "article id is required")
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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		existing, err := h.articleUC.Get(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "article not found")
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
			existing.CategoryId = input.CategoryID
		}
		if input.MediaID != "" {
			existing.MediaId = input.MediaID
		}
		existing.Thumbnail = input.Thumbnail // Allow empty string to clear
		if input.State != "" {
			existing.State = input.State
		}
		if input.Tags != nil {
			existing.Tags = input.Tags
		}
		existing.Featured = input.Featured

		updated, err := h.articleUC.Update(r.Context(), existing)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, updated)
	}
}

func (h *AdminHandler) adminDeleteArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "article id is required")
			return
		}

		if err := h.articleUC.Delete(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, nil)
	}
}

func (h *AdminHandler) adminUpdateArticleState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "article id is required")
			return
		}

		var input struct {
			State string `json:"state" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.articleUC.UpdateState(r.Context(), id, input.State); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, nil)
	}
}

// ================================
// Admin Category Management Handlers
// ================================

func (h *AdminHandler) adminListCategories() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		categories, err := h.categoryUC.ListCategories(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		total := len(categories)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		server.Page(gc, categories[start:end], int64(total), page, pageSize)
	}
}

func (h *AdminHandler) adminGetCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		idStr := gc.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid category id")
			return
		}

		cat, err := h.categoryUC.GetCategory(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "category not found")
			return
		}

		server.OK(gc, cat)
	}
}

func (h *AdminHandler) adminCreateCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var input struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug" binding:"required"`
			Description string `json:"description"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		cat := &contentbiz.Category{
			Name:        input.Name,
			Slug:        input.Slug,
			Description: input.Description,
		}

		created, err := h.categoryUC.CreateCategory(r.Context(), cat)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, created)
	}
}

func (h *AdminHandler) adminUpdateCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		idStr := gc.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid category id")
			return
		}

		var input contentbiz.UpdateCategoryInput
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(r.Context(), id, &input)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, updated)
	}
}

func (h *AdminHandler) adminPatchCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		idStr := gc.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid category id")
			return
		}

		var input contentbiz.UpdateCategoryInput
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(r.Context(), id, &input)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, updated)
	}
}

func (h *AdminHandler) adminDeleteCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		idStr := gc.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid category id")
			return
		}

		if err := h.categoryUC.DeleteCategory(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, nil)
	}
}

func (h *AdminHandler) adminUpdateUserRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "user id is required")
			return
		}

		var input struct {
			Role string `json:"role" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		validRoles := map[string]bool{"user": true, "admin": true, "editor": true}
		if !validRoles[input.Role] {
			server.Fail(gc, server.ErrBadRequest, "invalid role, must be one of: user, admin, editor")
			return
		}

		if err := h.userUC.SetUserRole(r.Context(), id, input.Role); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &emptypb.Empty{})
	}
}
