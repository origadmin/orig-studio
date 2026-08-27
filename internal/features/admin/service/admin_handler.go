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

	"google.golang.org/protobuf/types/known/emptypb"

	pb "origadmin/application/origstudio/api/gen/v1/user"
	mediapb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
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
	notificationUC *contentbiz.NotificationUseCase
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
	notificationUC *contentbiz.NotificationUseCase,
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
		notificationUC: notificationUC,
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
			stats.GET("/dashboard", server.WithAdminCtx(h.jwt, h.getDashboardStats()))
			stats.GET("/medias", server.WithAdminCtx(h.jwt, h.getMediaStats()))
			stats.GET("/users", server.WithAdminCtx(h.jwt, h.getUserStats()))
			stats.GET("/traffic", server.WithAdminCtx(h.jwt, h.getTrafficStats()))
			stats.GET("/revenue", server.WithAdminCtx(h.jwt, h.getRevenueStats()))
		}

		// ================================
		// 1.5 Media Management (Admin - UUID only)
		// ================================
		medias := admin.Group("/medias")
		{
			medias.GET("", server.WithAdminCtx(h.jwt, h.adminListMedias()))
			medias.GET("/:id", server.WithAdminCtx(h.jwt, h.adminGetMedia()))
			medias.PUT("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", h.adminUpdateMedia()))
			medias.DELETE("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:delete", h.adminDeleteMedia()))
			medias.GET("/:id/stats", server.WithAdminCtx(h.jwt, h.adminGetMediaStats()))
			medias.GET("/:id/variants", server.WithAdminCtx(h.jwt, h.adminGetMediaVariants()))
			medias.PUT("/:id/state", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", h.adminUpdateMediaState()))
			medias.GET("/:id/tasks", server.WithAdminCtx(h.jwt, h.adminGetMediaTasks()))
			medias.POST("/:id/tasks/:taskId/retry", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", h.adminRetryMediaTask()))
			// TODO(BUG-087后续): 完整性校验 - 后端 API 尚未实现,前端调用会 404
			// 设计文档: docs/ee/modules/media/integrity/02-API_ENDPOINTS.md
			// 待实现:
			//   medias.POST("/:id/integrity-check", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", h.adminCheckMediaIntegrity()))
			//   medias.POST("/:id/repair", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", h.adminRepairMedia()))
		}

		// ================================
		// 2. Channel Management (Admin - UUID only)
		// ================================
		channels := admin.Group("/channels")
		{
			channels.GET("", server.WithAdminCtx(h.jwt, h.adminListChannels()))
			channels.GET("/:id", server.WithAdminCtx(h.jwt, h.adminGetChannelDetail()))
			channels.PUT("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:write", h.adminUpdateChannel()))
			channels.DELETE("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "media:delete", h.adminDeleteChannel()))
			channels.POST("", server.WithAdminCtx(h.jwt, h.adminCreateChannel()))
		}

		// ================================
		// 3. Encoding Management
		// ================================
		encoding := admin.Group("/encoding")
		{
			encoding.GET("/tasks", server.WithAdminCtx(h.jwt, h.getAllEncodingTasks()))
			encoding.GET("/status", server.WithAdminCtx(h.jwt, h.getEncodingStatus()))
			encoding.POST("/tasks/:taskId/retry", server.WithAdminCtx(h.jwt, h.retryTask()))
			encoding.POST("/retry-failed", server.WithAdminCtx(h.jwt, h.retryAllFailedTasks()))

			profiles := encoding.Group("/profiles")
			{
				profiles.GET("", server.WithAdminCtx(h.jwt, h.listEncodeProfiles()))
				profiles.POST("", server.WithAdminCtx(h.jwt, h.createEncodeProfile()))
				profiles.POST("/preview", server.WithAdminCtx(h.jwt, h.previewEncodeCommand()))
				profiles.GET("/:id", server.WithAdminCtx(h.jwt, h.getEncodeProfile()))
				profiles.PUT("/:id", server.WithAdminCtx(h.jwt, h.updateEncodeProfile()))
				profiles.DELETE("/:id", server.WithAdminCtx(h.jwt, h.deleteEncodeProfile()))
			}
		}

		// ================================
		// 3.5 SSE for transcoding progress (admin only)
		// ================================
		// Uses query parameter ?token=<jwt> for authentication because
		// EventSource API does not support custom headers.
		admin.GET("/medias/transcoding/events", server.WithAdminCtx(h.jwt, h.sseTranscodingEvents()))

		// ================================
		// 3. System Settings
		// ================================
		settings := admin.Group("/settings")
		{
			settings.GET("", server.WithAdminCtx(h.jwt, h.getSystemSettings()))
			settings.GET("/info", server.WithAdminCtx(h.jwt, h.getSystemInfo()))
			settings.PUT("", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "system:config", h.updateSystemSettings()))
			settings.GET("/:key", server.WithAdminCtx(h.jwt, h.getSystemSettingByKey()))
			settings.PUT("/:key", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "system:config", h.updateSystemSettingByKey()))
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
			playlists.GET("", server.WithAdminCtx(h.jwt, h.adminListPlaylists()))
			playlists.GET("/:id", server.WithAdminCtx(h.jwt, h.adminGetPlaylistDetail())) // :id = UUID
			playlists.POST("", server.WithAdminCtx(h.jwt, h.adminCreatePlaylist()))
			playlists.PUT("/:id", server.WithAdminCtx(h.jwt, h.adminUpdatePlaylist()))    // :id = UUID
			playlists.DELETE("/:id", server.WithAdminCtx(h.jwt, h.adminDeletePlaylist())) // :id = UUID
		}

		// ================================
		// 6. User Management
		// ================================
		users := admin.Group("/users")
		{
			users.GET("", server.WithAdminCtx(h.jwt, h.adminListUsers()))
			users.POST("", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", h.adminCreateUser()))
			users.GET("/:id", server.WithAdminCtx(h.jwt, h.adminGetUser()))
			users.PUT("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", h.adminUpdateUser()))
			users.DELETE("/:id", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", h.adminDeleteUser()))
			users.PATCH("/:id/status", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", h.adminUpdateUserStatus()))
			users.PATCH("/:id/role", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "user:manage", h.adminUpdateUserRole()))
		}

		// ================================
		// 7. Category Management
		// ================================
		categories := admin.Group("/categories")
		{
			categories.GET("", server.WithAdminCtx(h.jwt, h.adminListCategories()))
			categories.GET("/:id", server.WithAdminCtx(h.jwt, h.adminGetCategory()))
			categories.POST("", server.WithAdminCtx(h.jwt, h.adminCreateCategory()))
			categories.PUT("/:id", server.WithAdminCtx(h.jwt, h.adminUpdateCategory()))
			categories.PATCH("/:id", server.WithAdminCtx(h.jwt, h.adminPatchCategory()))
			categories.DELETE("/:id", server.WithAdminCtx(h.jwt, h.adminDeleteCategory()))
		}

		// ================================
		// 8. Article Management
		// ================================
		articles := admin.Group("/articles")
		{
			articles.GET("", server.WithAdminCtx(h.jwt, h.adminListArticles()))
			articles.GET("/:id", server.WithAdminCtx(h.jwt, h.adminGetArticle()))
			articles.POST("", server.WithAdminCtx(h.jwt, h.adminCreateArticle()))
			articles.PUT("/:id", server.WithAdminCtx(h.jwt, h.adminUpdateArticle()))
			articles.DELETE("/:id", server.WithAdminCtx(h.jwt, h.adminDeleteArticle()))
			articles.PATCH("/:id/state", server.WithAdminCtx(h.jwt, h.adminUpdateArticleState()))
		}

		// ================================
		// 9. Notification Management (Admin)
		// ================================
		notifications := admin.Group("/notifications")
		{
			notifications.GET("", server.WithAdminCtx(h.jwt, h.adminListNotifications()))
			notifications.POST("/batch", server.WithAdminAndPermCtx(h.jwt, h.permChecker, "system:config", h.adminBatchCreateNotifications()))
		}
	}
}

// --- Stats Panel Handlers ---

func (h *AdminHandler) getDashboardStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.statsRepo == nil {
			return server.FailCtx(ctx, server.ErrInternal, "stats service not available")
		}

		stats, err := h.statsRepo.GetExtendedDashboardStats(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"total_users":           stats.TotalUsers,
			"total_media":           stats.TotalMedia,
			"total_views":           stats.TotalViews,
			"total_comments":        stats.TotalComments,
			"total_subscribers":     stats.TotalSubscribers,
			"total_revenue":         0,
			"active_users":          h.statsRepo.GetActiveUsersToday(ctx),
			"new_users_today":       stats.NewUsersToday,
			"new_media_today":       stats.NewMediaToday,
			"new_views_today":       0,
			"new_comments_today":    stats.NewCommentsToday,
			"new_subscribers_today": stats.NewSubsToday,
			"media_by_type":         stats.MediaByType,
			"users_by_role":         stats.UsersByRole,
			"views_by_date":         h.statsRepo.GetViewsByDate(ctx, 7),
			"media_by_date":         h.statsRepo.GetMediaByDate(ctx, 7),
			"top_categories":        h.statsRepo.GetTopCategories(ctx, 5),
			"top_creators":          h.statsRepo.GetTopCreators(ctx, 5),
			"top_media":             h.statsRepo.GetTopMedia(ctx, 5),
		})
	}
}

func (h *AdminHandler) getMediaStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.statsRepo == nil {
			return server.FailCtx(ctx, server.ErrInternal, "stats service not available")
		}

		stats, err := h.statsRepo.GetMediaStats(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"total_media":      stats.Total,
			"video_count":      stats.VideoCount,
			"image_count":      stats.ImageCount,
			"audio_count":      stats.AudioCount,
			"public_count":     stats.PublicCount,
			"private_count":    stats.PrivateCount,
			"pending_encoding": stats.EncodingPending,
			"encoding_failed":  stats.EncodingFailed,
		})
	}
}

func (h *AdminHandler) getUserStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.statsRepo == nil {
			return server.FailCtx(ctx, server.ErrInternal, "stats service not available")
		}

		stats, err := h.statsRepo.GetUserStats(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"total_users":     stats.Total,
			"active_users":    stats.ActiveToday,
			"new_users_today": stats.NewToday,
			"admin_count":     stats.AdminCount,
			"editor_count":    stats.EditorCount,
			"regular_count":   stats.RegularCount,
		})
	}
}

func (h *AdminHandler) getTrafficStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.statsRepo == nil {
			return server.FailCtx(ctx, server.ErrInternal, "stats service not available")
		}

		stats, err := h.statsRepo.GetDashboardStats(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"total_views":  stats.TotalViews,
			"today_views":  stats.ViewsToday,
			"new_media":    stats.NewMediaToday,
			"new_users":    stats.NewUsersToday,
			"encoding":     map[string]any{"pending": stats.EncodingPending, "failed": stats.EncodingFailed},
		})
	}
}

func (h *AdminHandler) getRevenueStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		// TODO: Implement revenue tracking
		return server.OKCtx(ctx, map[string]any{
			"total_revenue":  0,
			"monthly_revenue": 0,
			"revenue_by_date": []interface{}{},
			"message":        "Revenue tracking is not yet implemented",
		})
	}
}

// --- Encoding Management Handlers ---

func (h *AdminHandler) getAllEncodingTasks() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		status := ctx.QueryVar("status")

		var filterType mediabiz.FilterType
		var specificStatus string

		switch status {
		case "":
			filterType = mediabiz.FilterTypeAll
		case "active":
			filterType = mediabiz.FilterTypeActive
		case "all":
			filterType = mediabiz.FilterTypeAll
		case "queued":
			filterType = mediabiz.FilterTypeSpecific
			specificStatus = "pending"
		case "completed":
			filterType = mediabiz.FilterTypeSpecific
			specificStatus = "success"
		default:
			parsedStatus := enums.ParseEncodingTaskStatus(status)
			if parsedStatus == enums.EncodingTaskStatusUnknown {
				return server.FailCtx(ctx, server.ErrBadRequest, "Invalid status parameter")
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
			ProfileFilter: ctx.QueryVar("profile"),
			ChunkFilter:   ctx.QueryVar("chunk"),
			SearchQuery:   ctx.QueryVar("search"),
		}

		if os := ctx.QueryVar("only_stats"); os == "true" {
			filter.OnlyStats = true
		}

		if p, err := strconv.Atoi(ctx.QueryVarDefault("page", "1")); err == nil {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(ctx.QueryVarDefault("page_size", "25")); err == nil {
			filter.PageSize = ps
		}
		// Normalize pagination parameters
		page, pageSize := repotypes.NormalizeHTTPPagination(filter.Page, filter.PageSize)
		filter.Page = page
		filter.PageSize = pageSize

		var mediaID *string
		if m := ctx.QueryVar("media_id"); m != "" {
			mediaID = &m
		}

		result, err := h.mediaUC.ListEncodingTasksFlat(ctx, filter, mediaID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		totalPages := 0
		if pageSize > 0 {
			totalPages = (result.Total + pageSize - 1) / pageSize
		}

		return server.OKCtx(ctx, map[string]any{
			"items":            result.Items,
			"total":            result.Total,
			"page":             result.Page,
			"page_size":        result.PageSize,
			"total_pages":      totalPages,
			"processing_count": result.ProcessingCount,
			"pending_count":    result.PendingCount,
			"queued_count":     result.PendingCount,
			"partial_count":    result.PartialCount,
			"failed_count":     result.FailedCount,
			"success_count":    result.SuccessCount,
		})
	}
}

func (h *AdminHandler) getEncodingStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		// TODO: Implement encoding status
		return server.OKCtx(ctx, map[string]any{
			"total_tasks":  100,
			"pending":      5,
			"processing":   3,
			"completed":    80,
			"failed":       12,
			"success_rate": "80%",
		})
	}
}

func (h *AdminHandler) retryTask() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		taskIDStr := ctx.Var("taskId")
		if taskIDStr == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid task ID")
		}

		task, err := h.mediaUC.RetryTask(ctx, taskIDStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, task)
	}
}

func (h *AdminHandler) retryAllFailedTasks() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		mediaIDStr := ctx.QueryVar("media_id")

		count, err := h.mediaUC.RetryAllFailedTasks(ctx, mediaIDStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"retried_count": count})
	}
}

// --- Encode Profile Handlers ---

func (h *AdminHandler) listEncodeProfiles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		profiles, err := h.mediaUC.ListEncodeProfiles(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, map[string]any{"profiles": profiles})
	}
}

func (h *AdminHandler) createEncodeProfile() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var profile mediadto.EncodeProfile
		if err := ctx.BindJSON(&profile); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}
		p, err := h.mediaUC.CreateEncodeProfile(ctx, &profile)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.CreatedCtx(ctx, map[string]any{"profile": p})
	}
}

func (h *AdminHandler) getEncodeProfile() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id, err := strconv.Atoi(ctx.Var("id"))
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid Profile ID")
		}
		p, err := h.mediaUC.GetEncodeProfile(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "Profile not found")
		}
		return server.OKCtx(ctx, map[string]any{"profile": p})
	}
}

func (h *AdminHandler) updateEncodeProfile() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id, err := strconv.Atoi(ctx.Var("id"))
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid Profile ID")
		}
		var profile mediadto.EncodeProfile
		if err := ctx.BindJSON(&profile); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}
		profile.Id = id
		p, err := h.mediaUC.UpdateEncodeProfile(ctx, &profile)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, map[string]any{"profile": p})
	}
}

func (h *AdminHandler) deleteEncodeProfile() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id, err := strconv.Atoi(ctx.Var("id"))
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid Profile ID")
		}
		if err := h.mediaUC.DeleteEncodeProfile(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

func (h *AdminHandler) previewEncodeCommand() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var profile mediadto.EncodeProfile
		if err := ctx.BindJSON(&profile); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}
		preview := h.mediaUC.GenerateCommandPreview(ctx, &profile)
		return server.OKCtx(ctx, map[string]any{"command": preview})
	}
}

// --- System Settings Handlers ---

func (h *AdminHandler) getSystemSettings() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		items, err := h.settingUC.ListAll(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		// BUG-232: return the authoritative flat map<string,string> contract
		// (matches proto + frontend live module system.ts). Sensitive values are
		// masked; the frontend keeps the masked placeholder on save (see
		// updateSystemSettings) so the real value is never overwritten.
		settings := make(map[string]string, len(items))
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			settings[masked.Key] = masked.Value
		}

		return server.OKCtx(ctx, map[string]any{"settings": settings})
	}
}

// getSystemSettingByKey returns a single setting, or — when key matches a
// category — all settings in that category. Frontend getByCategory calls
// GET /admin/settings/:category, so we accept either a category or a key.
func (h *AdminHandler) getSystemSettingByKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}
		key := ctx.Var("key")
		items, err := h.settingUC.ListAll(ctx.Request().Context())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		var matched []*systemdto.SettingDTO
		for _, item := range items {
			if string(item.Category) == key || item.Key == key {
				matched = append(matched, h.settingUC.MaskSensitive(item))
			}
		}
		if len(matched) == 0 {
			return server.FailCtx(ctx, server.ErrNotFound, "setting not found: "+key)
		}
		return server.OKCtx(ctx, map[string]any{
			"category": key,
			"settings": matched,
		})
	}
}

// updateSystemSettingByKey updates a single setting by key. Frontend
// updateOne sends {value}; we upsert that one key and return the result.
func (h *AdminHandler) updateSystemSettingByKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}
		key := ctx.Var("key")
		var req struct {
			Value string `json:"value"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}
		existing, err := h.settingUC.GetByKey(ctx.Request().Context(), key)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "setting not found: "+key)
		}
		if err := systemservice.ValidateSettingValue(req.Value, existing.Type); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid value for "+key+": "+err.Error())
		}
		s := &systemdto.SettingDTO{
			Key:           existing.Key,
			Value:         req.Value,
			Type:          existing.Type,
			Category:      existing.Category,
			Description:   existing.Description,
			IsSensitive:   existing.IsSensitive,
			FallbackValue: existing.FallbackValue,
			IsBuiltin:     existing.IsBuiltin,
		}
		result, err := h.settingUC.Upsert(ctx.Request().Context(), s)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, result)
	}
}

// getSystemInfo returns runtime system information for the admin settings page.
// Response format matches the frontend SystemInfo interface.
//
// Future extensibility (multi-instance): The response can be extended to include
// an "instances" array for distributed deployments, while keeping backward
// compatibility by retaining the top-level fields as the "local" instance's data.
func (h *AdminHandler) getSystemInfo() http2.HandlerFunc {
	return func(ctx http2.Context) error {
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

		info := map[string]any{
			"version":      h.appVersion,
			"goVersion":    runtime.Version(),
			"database":     dbName,
			"os":           runtime.GOOS + "/" + runtime.GOARCH,
			"uptime":       uptimeStr,
			"totalMemory":  formatBytes(totalMem),
			"usedMemory":   formatBytes(usedMem),
			"cpuUsage":     float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()),
			"memoryUsage":  memUsagePercent,
			"numCPU":       runtime.NumCPU(),
			"numGoroutine": runtime.NumGoroutine(),
		}

		return server.OKCtx(ctx, info)
	}
}

// maskedSensitiveValue is what MaskSensitive substitutes for sensitive values;
// the frontend echoes it back unchanged on save, so we must NOT persist it.
const maskedSensitiveValue = "******"

func (h *AdminHandler) updateSystemSettings() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		// BUG-232: authoritative flat map<string,string> contract (proto +
		// frontend system.ts). Replaces the legacy categories/array shape.
		var req struct {
			Settings map[string]string `json:"settings" binding:"required"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		updated := make(map[string]string, len(req.Settings))
		for key, value := range req.Settings {
			existing, err := h.settingUC.GetByKey(ctx, key)
			if err != nil {
				defaults := systemdal.DefaultSettings()
				var defaultSetting *systemdto.SettingDTO
				for _, d := range defaults {
					if d.Key == key {
						defaultSetting = d
						break
					}
				}
				if defaultSetting == nil {
					return server.FailCtx(ctx, server.ErrBadRequest, "unknown setting key: "+key)
				}
				existing = defaultSetting
			}

			// Sensitive fields come back from GET masked as "******". If the
			// value is unchanged (still the placeholder), keep the stored value
			// instead of overwriting the real secret with the mask.
			if existing.IsSensitive && value == maskedSensitiveValue {
				updated[key] = existing.Value
				continue
			}

			if err := systemservice.ValidateSettingValue(value, existing.Type); err != nil {
				return server.FailCtx(ctx, server.ErrBadRequest, "invalid value for "+key+": "+err.Error())
			}

			s := &systemdto.SettingDTO{
				Key:           existing.Key,
				Value:         value,
				Type:          existing.Type,
				Category:      existing.Category,
				Description:   existing.Description,
				IsSensitive:   existing.IsSensitive,
				FallbackValue: existing.FallbackValue,
				IsBuiltin:     existing.IsBuiltin,
			}
			result, err := h.settingUC.Upsert(ctx, s)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
			updated[key] = result.Value
		}

		return server.OKCtx(ctx, map[string]any{"settings": updated})
	}
}

// --- Tag Management Handlers ---
// B087-R2 Fix: These handlers also use TagResponse DTO for frontend-compatible field names.
// Note: Tag routes are handled by AdminTagHandler, but these methods exist
// for backward compatibility if called from AdminHandler.

func (h *AdminHandler) listTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		// Parse query parameters
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))

		// B087-R2 Fix: Support both "search" and "keyword" parameters
		search := ctx.QueryVar("search")
		if search == "" {
			search = ctx.QueryVar("keyword")
		}

		status := ctx.QueryVar("status")
		sortBy := ctx.QueryVarDefault("sort_by", "create_time")
		sortOrder := ctx.QueryVarDefault("sort_order", "desc")

		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		// Get tags
		tags, total, err := h.tagService.List(
			ctx,
			page,
			pageSize,
			search,
			status,
			sortBy,
			sortOrder,
		)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, "Failed to list tags")
		}

		// B087-R2 Fix: Convert entity.Tag to TagResponse DTO
		tagResponses := ToTagResponseList(tags)

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		return server.OKCtx(ctx, map[string]any{
			"items":       tagResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		})
	}
}

func (h *AdminHandler) getTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")

		tag, err := h.tagService.Get(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "Tag not found")
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		return server.OKCtx(ctx, ToTagResponse(tag))
	}
}

func (h *AdminHandler) createTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug"` // Optional: auto-generated from name when empty
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid request")
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

		createdTag, err := h.tagService.Create(ctx, tag)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		return server.OKCtx(ctx, ToTagResponse(createdTag))
	}
}

func (h *AdminHandler) updateTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")

		var req struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid request")
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

		updatedTag, err := h.tagService.Update(ctx, id, updates)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		return server.OKCtx(ctx, ToTagResponse(updatedTag))
	}
}

func (h *AdminHandler) deleteTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")

		if err := h.tagService.Delete(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		return server.OKCtx(ctx, nil)
	}
}

func (h *AdminHandler) bulkTagOperation() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid request")
		}

		if req.Action != "delete" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Unsupported action")
		}

		count, err := h.tagService.BulkDelete(ctx, req.IDs)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"success": count,
			"failed":  len(req.IDs) - count,
		})
	}
}

func (h *AdminHandler) exportTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.tagService == nil {
			return server.FailCtx(ctx, server.ErrInternal, "tag service not available")
		}

		status := ctx.QueryVar("status")
		tags, _, err := h.tagService.List(ctx, 1, 10000, "", status, "create_time", "desc")
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().Header().Set("Content-Disposition", "attachment; filename=tags-export.json")
		return ctx.JSON(http.StatusOK, tags)
	}
}

func (h *AdminHandler) importTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.tagService == nil {
			return server.FailCtx(ctx, server.ErrInternal, "tag service not available")
		}

		var input []struct {
			Title       string `json:"title" binding:"required"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		var imported, failed int
		for _, item := range input {
			tag := &dto.TagDTO{
				Title:       item.Title,
				Slug:        item.Slug,
				Description: item.Description,
				Color:       item.Color,
				Status:      dto.TagStatusType(item.Status),
			}
			if tag.Status == "" {
				tag.Status = dto.TagStatusActive
			}
			_, err := h.tagService.Create(ctx, tag)
			if err != nil {
				failed++
				continue
			}
			imported++
		}

		return server.OKCtx(ctx, map[string]any{
			"imported": imported,
			"failed":   failed,
			"total":    len(input),
		})
	}
}

// ================================
// Admin Channel Management Handlers (v3.2 - UUID only)
// ================================

func (h *AdminHandler) adminListChannels() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.channelUC.ListChannels(ctx, page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		server.PageCtx(ctx, items, int64(total), page, pageSize)
		return nil
	}
}

func (h *AdminHandler) adminGetChannelDetail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "channel id is required")
		}

		if !validation.IsValidUUID(id) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
		}

		ch, err := h.channelUC.GetChannel(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found")
		}

		return server.OKCtx(ctx, ch)
	}
}

func (h *AdminHandler) adminUpdateChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "channel id is required")
		}

		if !validation.IsValidUUID(id) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
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
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existingCh, err := h.channelUC.GetChannel(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found")
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

		updated, err := h.channelUC.UpdateChannel(ctx, chItem, "", true)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, updated)
	}
}

func (h *AdminHandler) adminDeleteChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "channel id is required")
		}

		if !validation.IsValidUUID(id) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_uuid_format: Admin channel API requires UUID format")
		}

		err := h.channelUC.DeleteChannel(ctx, id, "", true)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, nil)
	}
}

func (h *AdminHandler) adminCreateChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Name        string `json:"name" binding:"required,min=3,max=150"`
			Description string `json:"description"`
			ShortToken  string `json:"short_token"`
			Status      string `json:"status"`
			UserID      string `json:"user_id" binding:"required"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if input.Status == "" {
			input.Status = "active"
		}

		slug := strings.ToLower(input.Name)
		slug = strings.ReplaceAll(slug, " ", "-")
		for strings.Contains(slug, "--") {
			slug = strings.ReplaceAll(slug, "--", "-")
		}

		ch := &contentbiz.Channel{
			Name:       input.Name,
			Title:      input.Name,
			Slug:       slug,
			Handle:     slug,
			Description: input.Description,
			ShortToken: input.ShortToken,
			Status:     input.Status,
			UserID:     input.UserID,
		}

		created, err := h.channelUC.CreateChannel(ctx, ch)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"channel": created})
	}
}

// ================================
// Admin Playlist Management Handlers
// ================================

func (h *AdminHandler) adminListPlaylists() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.channelUC.ListPlaylists(ctx, page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		server.PageCtx(ctx, items, int64(total), page, pageSize)
		return nil
	}
}

func (h *AdminHandler) adminGetPlaylistDetail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "playlist id is required")
		}

		if !validation.IsValidUUID(id) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
		}

		playlist, err := h.channelUC.GetPlaylist(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "playlist_not_found")
		}

		return server.OKCtx(ctx, map[string]any{"playlist": playlist})
	}
}

func (h *AdminHandler) adminCreatePlaylist() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			UserID      string `json:"user_id" binding:"required"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
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

		created, err := h.channelUC.CreatePlaylist(ctx, playlist)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"playlist": created})
	}
}

func (h *AdminHandler) adminUpdatePlaylist() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "playlist id is required")
		}

		if !validation.IsValidUUID(id) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
		}

		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			IsPublic    *bool  `json:"is_public"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existingPlaylist, err := h.channelUC.GetPlaylist(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "playlist_not_found")
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

		updated, err := h.channelUC.UpdatePlaylist(ctx, playlistItem, "", true)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"playlist": updated})
	}
}

func (h *AdminHandler) adminDeletePlaylist() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "playlist id is required")
		}

		if !validation.IsValidUUID(id) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_uuid_format: Admin playlist API requires UUID format")
		}

		err := h.channelUC.DeletePlaylist(ctx, id, "", true)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, nil)
	}
}

// ================================
// Admin User Management Handlers
// ================================

func (h *AdminHandler) adminListUsers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		keyword := ctx.QueryVar("keyword")

		opts := &userdto.UserQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  keyword,
			},
		}

		// Filter by role if specified
		if role := ctx.QueryVar("role"); role != "" {
			opts.Role = role
		}

		// Filter by status if specified
		if statusStr := ctx.QueryVar("status"); statusStr != "" {
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

		users, total, err := h.userUC.ListUsers(ctx, opts)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.ListUsersResponse{
			Items:    users,
			Total:    total,
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
	}
}

func (h *AdminHandler) adminCreateUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password"`                           // optional: admin can create user without password
			Nickname string `json:"nickname"`
			Role     string `json:"role"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		// If password is provided, validate minimum length
		if input.Password != "" && len(input.Password) < 6 {
			return server.FailCtx(ctx, server.ErrBadRequest, "password must be at least 6 characters")
		}

		// Hash password if provided; otherwise generate a random one
		var hashedPassword string
		var err error
		if input.Password != "" {
			hashedPassword, err = h.userUC.HashPassword(input.Password)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, "failed to hash password")
			}
		} else {
			// Generate a random password so the account is not left without one
			randomPwd := generateRandomPassword(12)
			hashedPassword, err = h.userUC.HashPassword(randomPwd)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, "failed to hash password")
			}
		}

		user := &types.User{
			Username: input.Username,
			Email:    input.Email,
			Nickname: input.Nickname,
		}

		created, err := h.userUC.CreateUser(ctx, user, hashedPassword)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		// Set role if specified (default is "user")
		role := input.Role
		if role == "" {
			role = "user"
		}
		if role != "user" {
			if err := h.userUC.SetUserRole(ctx, created.Id, role); err != nil {
				return server.FailCtx(ctx, server.ErrInternal, "failed to set role: "+err.Error())
			}
		}

		return server.CreatedCtx(ctx, &pb.CreateUserResponse{User: created})
	}
}

func (h *AdminHandler) adminGetUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "user id is required")
		}

		user, err := h.userUC.GetUser(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "user not found")
		}

		return server.OKCtx(ctx, &pb.GetUserResponse{User: user})
	}
}

func (h *AdminHandler) adminUpdateUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "user id is required")
		}

		existing, err := h.userUC.GetUser(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "user not found")
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

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
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
			if err := h.userUC.SetUserRole(ctx, id, input.Role); err != nil {
				return server.FailCtx(ctx, server.ErrInternal, "failed to update role: "+err.Error())
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
				if err := h.userUC.UpdateUserStatus(ctx, id, int8(statusCode)); err != nil {
					return server.FailCtx(ctx, server.ErrInternal, "failed to update status: "+err.Error())
				}
			}
		}

		updated, err := h.userUC.UpdateUser(ctx, existing)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateUserResponse{User: updated})
	}
}

func (h *AdminHandler) adminDeleteUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "user id is required")
		}

		if err := h.userUC.DeleteUser(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.DeleteUserResponse{Empty: &emptypb.Empty{}})
	}
}

func (h *AdminHandler) adminUpdateUserStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "user id is required")
		}

		var input struct {
			Status int32 `json:"status" binding:"required"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if err := h.userUC.UpdateUserStatus(ctx, id, int8(input.Status)); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &emptypb.Empty{})
	}
}

// --- Admin Notification Handlers ---

func (h *AdminHandler) adminListNotifications() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.notificationUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "notification service not available")
		}

		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.notificationUC.ListAllNotifications(ctx, page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		server.PageCtx(ctx, items, int64(total), page, pageSize)
		return nil
	}
}

func (h *AdminHandler) adminBatchCreateNotifications() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.notificationUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "notification service not available")
		}

		var input struct {
			UserIDs []string `json:"user_ids" binding:"required,min=1"`
			Action  string   `json:"action" binding:"required,max=30"`
			Notify  bool     `json:"notify"`
			Method  string   `json:"method"`
			Title   string   `json:"title" binding:"required,max=200"`
			Body    string   `json:"body" binding:"required"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		n := &contentbiz.Notification{
			Action: input.Action,
			Notify: input.Notify,
			Method: input.Method,
			Title:  input.Title,
			Body:   input.Body,
		}

		created, err := h.notificationUC.BatchCreateNotifications(ctx, input.UserIDs, n)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"created": len(created), "total": len(input.UserIDs)})
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
func (h *AdminHandler) sseTranscodingEvents() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.mediaService != nil {
			h.mediaService.SSEHandler(ctx.Response(), ctx.Request())
			return nil
		}
		ctx.Response().WriteHeader(http.StatusNotFound)
		return nil
	}
}

// ==================== Media Management Handlers ====================

// adminListMedias handles GET /admin/medias
func (h *AdminHandler) adminListMedias() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		page, pageSize = repotypes.NormalizePagination(page, pageSize)

		opts := &mediadto.MediaQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  ctx.QueryVar("keyword"),
			},
			AdminMode: true, // Admin sees all medias regardless of state
		}

		if state := ctx.QueryVar("state"); state != "" {
			opts.State = state
		}
		// Media type filtering: default to video for admin pages to prevent
		// images/audios from mixing into video list. Use "all" to see all types.
		if mediaType := ctx.QueryVar("type"); mediaType != "" && mediaType != "all" {
			opts.MediaType = mediaType
		} else if mediaType == "" {
			// No type specified: default to video for admin media page
			opts.MediaType = "video"
		}
		if tagsStr := ctx.QueryVar("tags"); tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			opts.Tags = tags
		}

		items, total, err := h.mediaUC.ListMedias(ctx, opts)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		totalPages := int32(0)
		if pageSize > 0 {
			totalPages = (total + int32(pageSize) - 1) / int32(pageSize)
		}

		return server.ProtoOKCtx(ctx, &mediapb.ListMediasResponse{
			Total:      total,
			Items:      items,
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalPages: totalPages,
		})
	}
}

// adminGetMedia handles GET /admin/medias/:id
func (h *AdminHandler) adminGetMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		media, err := h.mediaUC.GetByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media_not_found")
		}

		return server.ProtoOKCtx(ctx, media)
	}
}

// adminUpdateMedia handles PUT /admin/medias/:id
// Supports partial updates: only fields present in the request body are updated.
func (h *AdminHandler) adminUpdateMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		// Fetch existing media first for partial update
		existing, err := h.mediaUC.GetByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media_not_found")
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
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
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

		updated, err := h.mediaUC.UpdateMedia(ctx, existing)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.ProtoOKCtx(ctx, updated)
	}
}

// adminDeleteMedia handles DELETE /admin/medias/:id
func (h *AdminHandler) adminDeleteMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		if err := h.mediaUC.DeleteMedia(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &emptypb.Empty{})
	}
}

// adminGetMediaStats handles GET /admin/medias/:id/stats
func (h *AdminHandler) adminGetMediaStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		media, err := h.mediaUC.GetByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media_not_found")
		}

		return server.OKCtx(ctx, map[string]any{
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
func (h *AdminHandler) adminGetMediaVariants() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		summary, err := h.mediaUC.GetMediaVariantsByUUID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, summary)
	}
}

// adminUpdateMediaState handles PUT /admin/medias/:id/state
func (h *AdminHandler) adminUpdateMediaState() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		var input struct {
			State   string `json:"state" binding:"required"`
			Comment string `json:"comment"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if err := h.mediaUC.UpdateMediaState(ctx, id, input.State); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		// Fetch updated media to return current state
		updated, err := h.mediaUC.GetByID(ctx, id)
		if err != nil {
			return server.OKCtx(ctx, map[string]any{"id": id, "state": input.State})
		}

		return server.OKCtx(ctx, map[string]any{
			"id":          updated.Id,
			"state":       updated.State,
			"update_time": updated.UpdateTime,
		})
	}
}

// adminGetMediaTasks handles GET /admin/medias/:id/tasks
func (h *AdminHandler) adminGetMediaTasks() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media id is required")
		}

		tasks, err := h.mediaUC.ListEncodingTasks(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"items": tasks,
			"total": len(tasks),
		})
	}
}

// adminRetryMediaTask handles POST /admin/medias/:id/tasks/:taskId/retry
func (h *AdminHandler) adminRetryMediaTask() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		taskID := ctx.Var("taskId")
		if taskID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "task id is required")
		}

		_, err := h.mediaUC.RetryTask(ctx, taskID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"message": "retry initiated"})
	}
}

// ================================
// Admin Article Management Handlers
// ================================

func (h *AdminHandler) adminListArticles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))

		req := &types.ListArticlesRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		}

		if v := ctx.QueryVar("state"); v != "" {
			req.State = v
		}
		if v := ctx.QueryVar("category_id"); v != "" {
			if catID, err := strconv.ParseInt(v, 10, 64); err == nil {
				req.CategoryId = catID
			}
		}
		if v := ctx.QueryVar("keyword"); v != "" {
			req.Keyword = v
		}

		resp, err := h.articleUC.List(ctx, req)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, resp)
	}
}

func (h *AdminHandler) adminGetArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "article id is required")
		}

		article, err := h.articleUC.Get(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "article not found")
		}

		return server.OKCtx(ctx, article)
	}
}

func (h *AdminHandler) adminCreateArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
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

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		userID := ""
		if claims, exists := ctx.Get("claims"); exists {
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

		created, err := h.articleUC.Create(ctx, article)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, created)
	}
}

func (h *AdminHandler) adminUpdateArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "article id is required")
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

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existing, err := h.articleUC.Get(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "article not found")
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

		updated, err := h.articleUC.Update(ctx, existing)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, updated)
	}
}

func (h *AdminHandler) adminDeleteArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "article id is required")
		}

		if err := h.articleUC.Delete(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, nil)
	}
}

func (h *AdminHandler) adminUpdateArticleState() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "article id is required")
		}

		var input struct {
			State string `json:"state" binding:"required"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if err := h.articleUC.UpdateState(ctx, id, input.State); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, nil)
	}
}

// ================================
// Admin Category Management Handlers
// ================================

func (h *AdminHandler) adminListCategories() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		categories, err := h.categoryUC.ListCategories(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
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

		server.PageCtx(ctx, categories[start:end], int64(total), page, pageSize)
		return nil
	}
}

func (h *AdminHandler) adminGetCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		cat, err := h.categoryUC.GetCategory(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "category not found")
		}

		return server.OKCtx(ctx, cat)
	}
}

func (h *AdminHandler) adminCreateCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug" binding:"required"`
			Description string `json:"description"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		cat := &contentbiz.Category{
			Name:        input.Name,
			Slug:        input.Slug,
			Description: input.Description,
		}

		created, err := h.categoryUC.CreateCategory(ctx, cat)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, created)
	}
}

func (h *AdminHandler) adminUpdateCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		var input contentbiz.UpdateCategoryInput
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(ctx, id, &input)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, updated)
	}
}

func (h *AdminHandler) adminPatchCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		var input contentbiz.UpdateCategoryInput
		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		updated, err := h.categoryUC.UpdateCategoryPartial(ctx, id, &input)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, updated)
	}
}

func (h *AdminHandler) adminDeleteCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		if err := h.categoryUC.DeleteCategory(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, nil)
	}
}

func (h *AdminHandler) adminUpdateUserRole() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "user id is required")
		}

		var input struct {
			Role string `json:"role" binding:"required"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		validRoles := map[string]bool{"user": true, "admin": true, "editor": true}
		if !validRoles[input.Role] {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid role, must be one of: user, admin, editor")
		}

		if err := h.userUC.SetUserRole(ctx, id, input.Role); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &emptypb.Empty{})
	}
}
