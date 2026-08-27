/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Stub handler for routes that frontend calls but backend has not implemented yet.
 * Returns empty/mock data to prevent 404 errors.
 */

package service

import (
	"net/http"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/features/content/dal"
	"origadmin/application/origstudio/internal/features/media/biz"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

// StubHandler registers all missing routes as stubs returning empty/mock data.
type StubHandler struct {
	jwt     *auth.Manager
	mediaUC *biz.MediaUseCase
	paths   *conf.StoragePaths
	logger  *log.Helper
	// subtitle holds the real BUG-186 subtitle endpoints (replaces the stubs).
	subtitle *SubtitleHandler
}

// NewStubHandler creates a new StubHandler.
func NewStubHandler(jwt *auth.Manager, mediaUC *biz.MediaUseCase, paths *conf.StoragePaths, logger log.Logger, data *dal.Data) *StubHandler {
	return &StubHandler{
		jwt:      jwt,
		mediaUC:  mediaUC,
		paths:    paths,
		logger:   log.NewHelper(log.With(logger, "module", "service.stub")),
		subtitle: NewSubtitleHandler(jwt, data, paths, logger),
	}
}

// RegisterRoutes registers all stub routes.
func (h *StubHandler) RegisterRoutes(r http2.Router) {
	// ================================
	// 1. Admin Media Management — MOVED to AdminHandler (B079)
	// ================================

	// ================================
	// 2. Review Module
	// ================================
	adminReview := r.Group("/admin/medias/review")
	adminReview.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminReview.GET("/pending", h.stubReviewPending())
		adminReview.GET("/history", h.stubReviewHistory())
		adminReview.POST("/batch", h.stubReviewBatch())
	}
	adminMediaReview := r.Group("/admin/medias/:id")
	adminMediaReview.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminMediaReview.PUT("/review", h.stubReviewMedia())
		adminMediaReview.GET("/review-logs", h.stubReviewLogs())
	}

	// ================================
	// 3. Portal / Config — MOVED to SystemHandler
	// ================================
	portal := r.Group("/portal")
	{
		portal.GET("/home", h.stubPortalHome())
		portal.GET("/trending", h.stubPortalTrending())
		portal.GET("/subscriptions", h.stubPortalSubscriptions())
	}

	// ================================
	// 4. Admin Nav Items — MOVED to PortalHandler (F031)
	// ================================

	// ================================
	// 5. Admin Banners — MOVED to PortalHandler (F031)
	// ================================

	// ================================
	// 6. Media Metadata / Sprite / Subtitle / Download/Stream/Thumbnail / Likes/Favorites/Shares / Update/Delete
	// ================================
	medias := r.Group("/medias")
	{
		medias.GET("/:token/metadata", h.stubMediaMetadata())
		medias.POST("/:token/metadata/mining", server.WithJWTCtx(h.jwt, h.stubMediaMetadataMining()))
		medias.GET("/:token/metadata/status", h.stubMediaMetadataStatus())
		medias.GET("/:token/metadata/key-frames", h.stubMediaMetadataKeyFrames())
		medias.GET("/:token/metadata/audio-waveform", h.stubMediaMetadataAudioWaveform())
		medias.GET("/:token/metadata/text-content", h.stubMediaMetadataTextContent())
		medias.GET("/:token/metadata/scene-changes", h.stubMediaMetadataSceneChanges())

		medias.GET("/:token/subtitles", h.subtitle.handleSubtitleList)
		medias.POST("/:token/subtitles", server.WithJWTCtx(h.jwt, h.subtitle.handleSubtitleCreate))

		medias.GET("/:token/download", h.stubMediaDownload())
		medias.GET("/:token/stream", h.stubMediaStream())
		medias.GET("/:token/thumbnail", h.stubMediaThumbnail())

		medias.PUT("/:token", server.WithJWTCtx(h.jwt, h.stubMediaUpdate()))
		medias.DELETE("/:token", server.WithJWTCtx(h.jwt, h.stubMediaDelete()))

		medias.POST("/upload", server.WithJWTCtx(h.jwt, h.stubMediaUpload()))

		medias.GET("/:token/tasks", h.stubMediaTasks())
		medias.POST("/:token/tasks/:taskId/retry", server.WithJWTCtx(h.jwt, h.stubMediaTaskRetry()))
	}

	// ================================
	// 7. Subtitle (root level) — BUG-186 real endpoints.
	// ================================
	subtitles := r.Group("/subtitles")
	{
		subtitles.DELETE("/:id", server.WithJWTCtx(h.jwt, h.subtitle.handleSubtitleDelete))
		subtitles.GET("/languages", h.subtitle.handleSubtitleLanguages)
	}

	// ================================
	// 8. Admin Sprite/Thumbnail regeneration (replaced by SpriteHandler)
	// ================================

	// ================================
	// 9. Admin Stats Revenue — MOVED to AdminHandler.getRevenueStats (duplicate route removed)
	// ================================

	// ================================
	// 10. Admin Settings — MOVED to AdminHandler
	// ================================
	adminSettings := r.Group("/admin/settings")
	adminSettings.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminSettings.DELETE("/:key", h.stubAdminSettingsDelete())
	}

	// ================================
	// 11. Admin Channels POST — MOVED to AdminHandler.adminCreateChannel (duplicate route removed)
	// ================================

	// ================================
	// 12. Admin Comments DELETE — MOVED to CommentModerationHandler (B087)
	// ================================

	// ================================
	// 13. Notification DELETE — MOVED to NotificationHandler
	// ================================

	// ================================
	// 14. User subscription/update/status routes
	// ================================
	users := r.Group("/users")
	{
		users.GET("/:slug/subscription", h.stubUserSubscription())
		users.POST("/:slug/subscribe", server.WithJWTCtx(h.jwt, h.stubUserSubscribe()))
		users.DELETE("/:slug/subscribe", server.WithJWTCtx(h.jwt, h.stubUserUnsubscribe()))
		users.PUT("/:slug", server.WithJWTCtx(h.jwt, h.stubUserUpdate()))
		users.PATCH("/:slug/status", server.WithJWTCtx(h.jwt, h.stubUserStatusUpdate()))
	}

	// ================================
	// 15. Playlist — MOVED to PlaylistHandler + MeHandler + AdminHandler
	// ================================
	mePlaylists := r.Group("/me/playlists")
	mePlaylists.Use(server.JWTMiddlewareCtx(h.jwt))
	{
		mePlaylists.PATCH("/:id/media/reorder", h.stubPlaylistMediaReorder())
	}

	// ================================
	// 16. Encoding status/events (public aliases)
	// ================================
	encoding := r.Group("/encoding")
	{
		encoding.GET("/status", h.stubEncodingStatus())
		encoding.GET("/events", h.stubEncodingEvents())
	}
}

// ==================== Admin Media Stubs ====================

func (h *StubHandler) stubAdminMediaList() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		return server.OKCtx(ctx, map[string]any{
			"code":    0,
			"message": "ok",
			"data": map[string]any{
				"items":     []interface{}{},
				"total":     0,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *StubHandler) stubAdminMediaGet() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": nil})
	}
}

func (h *StubHandler) stubAdminMediaUpdate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": nil})
	}
}

func (h *StubHandler) stubAdminMediaDelete() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok"})
	}
}

func (h *StubHandler) stubAdminMediaStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"code":    0,
			"message": "ok",
			"data": map[string]any{
				"view_count":    0,
				"like_count":    0,
				"comment_count": 0,
				"share_count":   0,
			},
		})
	}
}

func (h *StubHandler) stubAdminMediaVariants() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubAdminMediaState() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok"})
	}
}

func (h *StubHandler) stubAdminMediaTasks() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubAdminMediaTaskRetry() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "retry initiated"})
	}
}

// ==================== Review Stubs ====================

func (h *StubHandler) stubReviewPending() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubReviewHistory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubReviewBatch() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "batch review completed"})
	}
}

func (h *StubHandler) stubReviewMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "review submitted"})
	}
}

func (h *StubHandler) stubReviewLogs() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

// ==================== Nav Items Stubs ====================

func (h *StubHandler) stubNavItemList() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubNavItemCreate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubNavItemUpdate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubNavItemDelete() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "deleted"})
	}
}

func (h *StubHandler) stubNavItemReorder() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "reordered"})
	}
}

// ==================== Banner Stubs ====================

func (h *StubHandler) stubBannerList() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubBannerCreate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubBannerUpdate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubBannerToggle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "toggled"})
	}
}

// ==================== Metadata Stubs ====================

func (h *StubHandler) stubMediaMetadata() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"duration":    0,
			"resolution":  "",
			"codec":       "",
			"bitrate":     0,
			"frame_rate":  0,
			"audio_codec": "",
		}})
	}
}

func (h *StubHandler) stubMediaMetadataMining() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "mining started"})
	}
}

func (h *StubHandler) stubMediaMetadataStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"status":   "pending",
			"progress": 0,
		}})
	}
}

func (h *StubHandler) stubMediaMetadataKeyFrames() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubMediaMetadataAudioWaveform() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubMediaMetadataTextContent() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": ""})
	}
}

func (h *StubHandler) stubMediaMetadataSceneChanges() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

// ==================== Sprite Stubs ====================

func (h *StubHandler) stubSpriteVTT() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return ctx.String(http.StatusOK, "WEBVTT\n\n")
	}
}

func (h *StubHandler) stubSpriteJPG() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		ctx.Response().WriteHeader(http.StatusNotFound)
		return nil
	}
}

// ==================== Subtitle Stubs ====================

func (h *StubHandler) stubSubtitleList() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubSubtitleCreate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubSubtitleDelete() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "deleted"})
	}
}

func (h *StubHandler) stubSubtitleLanguages() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

// ==================== Media Download/Stream/Thumbnail ====================

// stubMediaDownload serves the original media file for download.
func (h *StubHandler) stubMediaDownload() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media token")
		}

		m, err := h.mediaUC.GetByShortToken(ctx, token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		// Prefer HLS file for streaming, fallback to original URL
		filePath := m.Url
		if filePath == "" {
			return server.FailCtx(ctx, server.ErrNotFound, "media file not available")
		}

		// Resolve to local storage path if applicable
		fullPath := h.paths.FullPath(filePath)
		filename := m.Title
		if m.Extension != "" {
			filename += "." + m.Extension
		}

		ctx.Response().Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		return ctx.File(fullPath)
	}
}

// stubMediaStream serves media for streaming (supports range requests via ctx.File).
func (h *StubHandler) stubMediaStream() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media token")
		}

		m, err := h.mediaUC.GetByShortToken(ctx, token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		// Prefer HLS file for streaming, fallback to original URL
		filePath := m.HlsFile
		if filePath == "" {
			filePath = m.Url
		}
		if filePath == "" {
			return server.FailCtx(ctx, server.ErrNotFound, "media stream not available")
		}

		fullPath := h.paths.FullPath(filePath)
		return ctx.File(fullPath)
	}
}

// stubMediaThumbnail serves the media thumbnail image.
func (h *StubHandler) stubMediaThumbnail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media token")
		}

		m, err := h.mediaUC.GetByShortToken(ctx, token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		if m.Thumbnail == "" {
			return server.FailCtx(ctx, server.ErrNotFound, "thumbnail not available")
		}

		fullPath := h.paths.FullPath(m.Thumbnail)
		return ctx.File(fullPath)
	}
}

// ==================== Regeneration Stubs ====================

func (h *StubHandler) stubRegenerateSprite() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "sprite regeneration started"})
	}
}

func (h *StubHandler) stubRegenerateThumbnail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "thumbnail regeneration started"})
	}
}

// ==================== Admin Settings Category/Key Stubs ====================

func (h *StubHandler) stubAdminSettingsCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubAdminSettingsUpdateKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "updated"})
	}
}

func (h *StubHandler) stubAdminSettingsDeleteKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "deleted"})
	}
}

// ==================== Notification Delete Stub ====================

// ==================== User Subscription Stubs ====================

func (h *StubHandler) stubUserSubscription() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"is_subscribed": false,
		}})
	}
}

func (h *StubHandler) stubUserSubscribe() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "subscribed"})
	}
}

func (h *StubHandler) stubUserUnsubscribe() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "unsubscribed"})
	}
}

// ==================== User Update/Status Stubs ====================

func (h *StubHandler) stubUserUpdate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubUserStatusUpdate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "status updated"})
	}
}

// ==================== Media Update/Delete Stubs ====================

func (h *StubHandler) stubMediaUpdate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubMediaDelete() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "deleted"})
	}
}

// ==================== Media Likes/Favorites/Shares (plural) Stubs ====================

func (h *StubHandler) stubMediaLikes() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"is_liked":      false,
			"is_disliked":   false,
			"like_count":    0,
			"dislike_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaLikeToggle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"is_liked":   true,
			"like_count": 1,
		})
	}
}

func (h *StubHandler) stubMediaLikeRemove() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"is_liked":   false,
			"like_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaFavorites() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"is_favorited":   false,
			"favorite_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaFavoriteToggle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"is_favorited":   true,
			"favorite_count": 1,
		})
	}
}

func (h *StubHandler) stubMediaShares() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"share_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaShareCreate() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "shared"})
	}
}

// ==================== Media Upload Stub ====================

func (h *StubHandler) stubMediaUpload() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "use /uploads/multipart for upload"})
	}
}

// ==================== Encoding Status/Events Stubs ====================

func (h *StubHandler) stubEncodingStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"total_tasks":  0,
			"pending":      0,
			"processing":   0,
			"completed":    0,
			"failed":       0,
			"success_rate": "0%",
		})
	}
}

func (h *StubHandler) stubEncodingEvents() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"events": []interface{}{}})
	}
}

// ==================== Media Tasks (deprecated) Stubs ====================

func (h *StubHandler) stubMediaTasks() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": map[string]any{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubMediaTaskRetry() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "retry initiated"})
	}
}

func (h *StubHandler) stubPlaylistMediaReorder() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "reordered"})
	}
}

func (h *StubHandler) stubAdminSettingsDelete() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		key := ctx.Var("key")
		return server.OKCtx(ctx, map[string]any{"code": 0, "message": "deleted", "key": key})
	}
}

func (h *StubHandler) stubPortalHome() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"sections": []interface{}{}, "total": 0})
	}
}

func (h *StubHandler) stubPortalTrending() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"items": []interface{}{}, "total": 0})
	}
}

func (h *StubHandler) stubPortalSubscriptions() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{"items": []interface{}{}, "total": 0})
	}
}
