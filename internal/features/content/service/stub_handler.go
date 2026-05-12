/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Stub handler for routes that frontend calls but backend has not implemented yet.
 * Returns empty/mock data to prevent 404 errors.
 */

package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
)

// StubHandler registers all missing routes as stubs returning empty/mock data.
type StubHandler struct {
	jwt *auth.Manager
}

// NewStubHandler creates a new StubHandler.
func NewStubHandler(jwt *auth.Manager) *StubHandler {
	return &StubHandler{jwt: jwt}
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
		adminReview.GET("/pending", server.HTTPToHandlerFunc(h.stubReviewPending()))
		adminReview.GET("/history", server.HTTPToHandlerFunc(h.stubReviewHistory()))
		adminReview.POST("/batch", server.HTTPToHandlerFunc(h.stubReviewBatch()))
	}
	adminMediaReview := r.Group("/admin/medias/:id")
	adminMediaReview.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminMediaReview.PUT("/review", server.HTTPToHandlerFunc(h.stubReviewMedia()))
		adminMediaReview.GET("/review-logs", server.HTTPToHandlerFunc(h.stubReviewLogs()))
	}

	// ================================
	// 3. Portal / Config — MOVED to SystemHandler
	// ================================

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
		// Metadata
		medias.GET("/:id/metadata", server.HTTPToHandlerFunc(h.stubMediaMetadata()))
		medias.POST("/:id/metadata/mining", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubMediaMetadataMining())))
		medias.GET("/:id/metadata/status", server.HTTPToHandlerFunc(h.stubMediaMetadataStatus()))
		medias.GET("/:id/metadata/key-frames", server.HTTPToHandlerFunc(h.stubMediaMetadataKeyFrames()))
		medias.GET("/:id/metadata/audio-waveform", server.HTTPToHandlerFunc(h.stubMediaMetadataAudioWaveform()))
		medias.GET("/:id/metadata/text-content", server.HTTPToHandlerFunc(h.stubMediaMetadataTextContent()))
		medias.GET("/:id/metadata/scene-changes", server.HTTPToHandlerFunc(h.stubMediaMetadataSceneChanges()))

		// Subtitle
		medias.GET("/:id/subtitles", server.HTTPToHandlerFunc(h.stubSubtitleList()))
		medias.POST("/:id/subtitles", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubSubtitleCreate())))

		// Download/Stream/Thumbnail
		medias.GET("/:id/download", server.HTTPToHandlerFunc(h.stubMediaDownload()))
		medias.GET("/:id/stream", server.HTTPToHandlerFunc(h.stubMediaStream()))
		medias.GET("/:id/thumbnail", server.HTTPToHandlerFunc(h.stubMediaThumbnail()))

		// Shares
		medias.GET("/:id/shares", server.HTTPToHandlerFunc(h.stubMediaShares()))
		medias.POST("/:id/shares", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubMediaShareCreate())))

		// Update/Delete
		medias.PUT("/:id", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubMediaUpdate())))
		medias.DELETE("/:id", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubMediaDelete())))

		// Upload alias
		medias.POST("/upload", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubMediaUpload())))

		// Tasks (deprecated)
		medias.GET("/:id/tasks", server.HTTPToHandlerFunc(h.stubMediaTasks()))
		medias.POST("/:id/tasks/:taskId/retry", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubMediaTaskRetry())))
	}

	// ================================
	// 7. Subtitle (root level)
	// ================================
	subtitles := r.Group("/subtitles")
	{
		subtitles.DELETE("/:id", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubSubtitleDelete())))
		subtitles.GET("/languages", server.HTTPToHandlerFunc(h.stubSubtitleLanguages()))
	}

	// ================================
	// 8. Admin Sprite/Thumbnail regeneration (replaced by SpriteHandler)
	// ================================

	// ================================
	// 9. Admin Stats Revenue
	// ================================
	adminStatsRevenue := r.Group("/admin/stats")
	adminStatsRevenue.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminStatsRevenue.GET("/revenue", server.HTTPToHandlerFunc(h.stubAdminStatsRevenue()))
	}

	// ================================
	// 10. Admin Settings — MOVED to AdminHandler
	// ================================

	// ================================
	// 11. Admin Channels POST
	// ================================
	adminChannels := r.Group("/admin/channels")
	adminChannels.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminChannels.POST("", server.HTTPToHandlerFunc(h.stubAdminChannelCreate()))
	}

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
		users.GET("/:id/subscription", server.HTTPToHandlerFunc(h.stubUserSubscription()))
		users.POST("/:id/subscribe", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubUserSubscribe())))
		users.DELETE("/:id/subscribe", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubUserUnsubscribe())))
		users.PUT("/:id", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubUserUpdate())))
		users.PATCH("/:id/status", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.stubUserStatusUpdate())))
	}

	// ================================
	// 15. Playlist — MOVED to PlaylistHandler + MeHandler + AdminHandler
	// ================================

	// ================================
	// 16. Encoding status/events (public aliases)
	// ================================
	encoding := r.Group("/encoding")
	{
		encoding.GET("/status", server.HTTPToHandlerFunc(h.stubEncodingStatus()))
		encoding.GET("/events", server.HTTPToHandlerFunc(h.stubEncodingEvents()))
	}
}

// ==================== Admin Media Stubs ====================

func (h *StubHandler) stubAdminMediaList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		server.OK(gc, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     []interface{}{},
				"total":     0,
				"page":      page,
				"page_size": pageSize,
			},
		})
	}
}

func (h *StubHandler) stubAdminMediaGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": nil})
	}
}

func (h *StubHandler) stubAdminMediaUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": nil})
	}
}

func (h *StubHandler) stubAdminMediaDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok"})
	}
}

func (h *StubHandler) stubAdminMediaStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"view_count":    0,
				"like_count":    0,
				"comment_count": 0,
				"share_count":   0,
			},
		})
	}
}

func (h *StubHandler) stubAdminMediaVariants() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubAdminMediaState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok"})
	}
}

func (h *StubHandler) stubAdminMediaTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubAdminMediaTaskRetry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "retry initiated"})
	}
}

// ==================== Review Stubs ====================

func (h *StubHandler) stubReviewPending() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubReviewHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubReviewBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "batch review completed"})
	}
}

func (h *StubHandler) stubReviewMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "review submitted"})
	}
}

func (h *StubHandler) stubReviewLogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

// ==================== Nav Items Stubs ====================

func (h *StubHandler) stubNavItemList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubNavItemCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubNavItemUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubNavItemDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "deleted"})
	}
}

func (h *StubHandler) stubNavItemReorder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "reordered"})
	}
}

// ==================== Banner Stubs ====================

func (h *StubHandler) stubBannerList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubBannerCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubBannerUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubBannerToggle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "toggled"})
	}
}

// ==================== Metadata Stubs ====================

func (h *StubHandler) stubMediaMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"duration":    0,
			"resolution":  "",
			"codec":       "",
			"bitrate":     0,
			"frame_rate":  0,
			"audio_codec": "",
		}})
	}
}

func (h *StubHandler) stubMediaMetadataMining() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "mining started"})
	}
}

func (h *StubHandler) stubMediaMetadataStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"status":   "pending",
			"progress": 0,
		}})
	}
}

func (h *StubHandler) stubMediaMetadataKeyFrames() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubMediaMetadataAudioWaveform() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubMediaMetadataTextContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": ""})
	}
}

func (h *StubHandler) stubMediaMetadataSceneChanges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

// ==================== Sprite Stubs ====================

func (h *StubHandler) stubSpriteVTT() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		gc.String(http.StatusOK, "WEBVTT\n\n")
	}
}

func (h *StubHandler) stubSpriteJPG() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		gc.Status(http.StatusNotFound)
	}
}

// ==================== Subtitle Stubs ====================

func (h *StubHandler) stubSubtitleList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubSubtitleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubSubtitleDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "deleted"})
	}
}

func (h *StubHandler) stubSubtitleLanguages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

// ==================== Media Download/Stream/Thumbnail Stubs ====================

func (h *StubHandler) stubMediaDownload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.Fail(gc, server.ErrNotFound, "download not available")
	}
}

func (h *StubHandler) stubMediaStream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.Fail(gc, server.ErrNotFound, "stream not available")
	}
}

func (h *StubHandler) stubMediaThumbnail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		gc.Status(http.StatusNotFound)
	}
}

// ==================== Regeneration Stubs ====================

func (h *StubHandler) stubRegenerateSprite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "sprite regeneration started"})
	}
}

func (h *StubHandler) stubRegenerateThumbnail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "thumbnail regeneration started"})
	}
}

// ==================== Admin Stats Revenue Stub ====================

func (h *StubHandler) stubAdminStatsRevenue() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"total_revenue":   0,
				"today_revenue":   0,
				"monthly_revenue": 0,
				"revenue_by_type": []interface{}{},
			},
		})
	}
}

// ==================== Admin Settings Category/Key Stubs ====================

func (h *StubHandler) stubAdminSettingsCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubAdminSettingsUpdateKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "updated"})
	}
}

func (h *StubHandler) stubAdminSettingsDeleteKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "deleted"})
	}
}

// ==================== Admin Channel Create Stub ====================

func (h *StubHandler) stubAdminChannelCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

// ==================== Notification Delete Stub ====================

// ==================== User Subscription Stubs ====================

func (h *StubHandler) stubUserSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"is_subscribed": false,
		}})
	}
}

func (h *StubHandler) stubUserSubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "subscribed"})
	}
}

func (h *StubHandler) stubUserUnsubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "unsubscribed"})
	}
}

// ==================== User Update/Status Stubs ====================

func (h *StubHandler) stubUserUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubUserStatusUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "status updated"})
	}
}

// ==================== Media Update/Delete Stubs ====================

func (h *StubHandler) stubMediaUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubMediaDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "deleted"})
	}
}

// ==================== Media Likes/Favorites/Shares (plural) Stubs ====================

func (h *StubHandler) stubMediaLikes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"is_liked":      false,
			"is_disliked":   false,
			"like_count":    0,
			"dislike_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaLikeToggle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"is_liked":   true,
			"like_count": 1,
		})
	}
}

func (h *StubHandler) stubMediaLikeRemove() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"is_liked":   false,
			"like_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaFavorites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"is_favorited":   false,
			"favorite_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaFavoriteToggle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"is_favorited":   true,
			"favorite_count": 1,
		})
	}
}

func (h *StubHandler) stubMediaShares() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"share_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaShareCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "shared"})
	}
}

// ==================== Media Upload Stub ====================

func (h *StubHandler) stubMediaUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "use /uploads/multipart for upload"})
	}
}

// ==================== Encoding Status/Events Stubs ====================

func (h *StubHandler) stubEncodingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"total_tasks":  0,
			"pending":      0,
			"processing":   0,
			"completed":    0,
			"failed":       0,
			"success_rate": "0%",
		})
	}
}

func (h *StubHandler) stubEncodingEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"events": []interface{}{}})
	}
}

// ==================== Media Tasks (deprecated) Stubs ====================

func (h *StubHandler) stubMediaTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubMediaTaskRetry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "retry initiated"})
	}
}
