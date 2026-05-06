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
func (h *StubHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// ================================
	// 1. Admin Media Management — MOVED to AdminHandler (B079)
	// CRUD routes (GET/PUT/DELETE /admin/medias/:id, GET /admin/medias, etc.)
	// are now handled by AdminHandler.adminListMedias/adminGetMedia/adminUpdateMedia/etc.
	// Only review stubs remain here.
	// ================================

	// ================================
	// 2. Review Module
	// ================================
	adminReview := rg.Group("/admin/medias/review")
	adminReview.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(adminReview)
		r.GET("/pending", h.stubReviewPending())
		r.GET("/history", h.stubReviewHistory())
		r.POST("/batch", h.stubReviewBatch())
	}
	adminMediaReview := rg.Group("/admin/medias/:id")
	adminMediaReview.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(adminMediaReview)
		r.PUT("/review", h.stubReviewMedia())
		r.GET("/review-logs", h.stubReviewLogs())
	}

	// ================================
	// 3. Portal / Config — MOVED to SystemHandler
	// SystemHandler.getPortalConfig() provides real implementation
	// reading from settings. Stub removed to avoid route conflict.
	// ================================

	// ================================
	// 4. Admin Nav Items
	// ================================
	adminNavItems := rg.Group("/admin/nav-items")
	adminNavItems.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(adminNavItems)
		r.GET("", h.stubNavItemList())
		r.POST("", h.stubNavItemCreate())
		r.PUT("/:id", h.stubNavItemUpdate())
		r.DELETE("/:id", h.stubNavItemDelete())
		r.PUT("/reorder", h.stubNavItemReorder())
	}

	// ================================
	// 5. Admin Banners
	// ================================
	adminBanners := rg.Group("/admin/banners")
	adminBanners.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(adminBanners)
		r.GET("", h.stubBannerList())
		r.POST("", h.stubBannerCreate())
		r.PUT("/:id", h.stubBannerUpdate())
		r.POST("/:id/toggle", h.stubBannerToggle())
	}

	// ================================
	// 6. Media Metadata / Sprite / Subtitle / Download/Stream/Thumbnail / Likes/Favorites/Shares / Update/Delete
	// ================================
	mediasR := ginadapter.NewStdRouterAdapter(rg)
	medias := mediasR.Group("/medias")
	{
		// Metadata
		medias.GET("/:id/metadata", h.stubMediaMetadata())
		medias.POST("/:id/metadata/mining", server.WithJWT(h.jwt, h.stubMediaMetadataMining()))
		medias.GET("/:id/metadata/status", h.stubMediaMetadataStatus())
		medias.GET("/:id/metadata/key-frames", h.stubMediaMetadataKeyFrames())
		medias.GET("/:id/metadata/audio-waveform", h.stubMediaMetadataAudioWaveform())
		medias.GET("/:id/metadata/text-content", h.stubMediaMetadataTextContent())
		medias.GET("/:id/metadata/scene-changes", h.stubMediaMetadataSceneChanges())

		// Sprite (replaced by SpriteHandler - routes registered separately)
		// medias.GET("/:id/sprite.vtt", h.stubSpriteVTT())
		// medias.GET("/:id/sprite.jpg", h.stubSpriteJPG())

		// Subtitle
		medias.GET("/:id/subtitles", h.stubSubtitleList())
		medias.POST("/:id/subtitles", server.WithJWT(h.jwt, h.stubSubtitleCreate()))

		// Download/Stream/Thumbnail
		medias.GET("/:id/download", h.stubMediaDownload())
		medias.GET("/:id/stream", h.stubMediaStream())
		medias.GET("/:id/thumbnail", h.stubMediaThumbnail())

		// NOTE: Likes/Favorites plural paths are now handled by MediaHandler
		// (POST /:id/likes, GET /:id/likes, DELETE /:id/likes,
		//  POST /:id/favorites, GET /:id/favorites, DELETE /:id/favorites)

		// Shares
		medias.GET("/:id/shares", h.stubMediaShares())
		medias.POST("/:id/shares", server.WithJWT(h.jwt, h.stubMediaShareCreate()))

		// Update/Delete
		medias.PUT("/:id", server.WithJWT(h.jwt, h.stubMediaUpdate()))
		medias.DELETE("/:id", server.WithJWT(h.jwt, h.stubMediaDelete()))

		// Upload alias
		medias.POST("/upload", server.WithJWT(h.jwt, h.stubMediaUpload()))

		// Tasks (deprecated)
		medias.GET("/:id/tasks", h.stubMediaTasks())
		medias.POST("/:id/tasks/:taskId/retry", server.WithJWT(h.jwt, h.stubMediaTaskRetry()))
	}

	// ================================
	// 7. Subtitle (root level)
	// ================================
	subtitlesR := ginadapter.NewStdRouterAdapter(rg)
	subtitles := subtitlesR.Group("/subtitles")
	{
		subtitles.DELETE("/:id", server.WithJWT(h.jwt, h.stubSubtitleDelete()))
		subtitles.GET("/languages", h.stubSubtitleLanguages())
	}

	// ================================
	// 8. Admin Sprite/Thumbnail regeneration (replaced by SpriteHandler)
	// ================================
	// adminMediaRegen := rg.Group("/admin/medias/:id")
	// adminMediaRegen.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	// {
	// 	adminMediaRegen.POST("/regenerate-sprite", h.stubRegenerateSprite())
	// 	adminMediaRegen.POST("/regenerate-thumbnail", h.stubRegenerateThumbnail())
	// }

	// ================================
	// 9. Admin Stats Revenue
	// ================================
	adminStatsRevenue := rg.Group("/admin/stats")
	adminStatsRevenue.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(adminStatsRevenue)
		r.GET("/revenue", h.stubAdminStatsRevenue())
	}

	// ================================
	// 10. Admin Settings — MOVED to AdminHandler
	// AdminHandler provides real GET/PUT /admin/settings implementation.
	// Per-key routes removed to avoid route conflict with AdminHandler.
	// ================================

	// ================================
	// 11. Admin Channels POST
	// ================================
	adminChannels := rg.Group("/admin/channels")
	adminChannels.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(adminChannels)
		r.POST("", h.stubAdminChannelCreate())
	}

	// ================================
	// 12. Admin Comments DELETE — MOVED to CommentModerationHandler (B087)
	// ================================

	// ================================
	// 13. Notification DELETE
	// ================================
	notifications := rg.Group("/notifications")
	notifications.Use(server.JWTMiddleware(h.jwt))
	{
		r := ginadapter.NewStdRouterAdapter(notifications)
		r.DELETE("/:id", h.stubNotificationDelete())
	}

	// ================================
	// 14. User subscription/update/status routes
	// ================================
	users := rg.Group("/users")
	{
		r := ginadapter.NewStdRouterAdapter(users)
		r.GET("/:id/subscription", h.stubUserSubscription())
		r.POST("/:id/subscribe", server.WithJWT(h.jwt, h.stubUserSubscribe()))
		r.DELETE("/:id/subscribe", server.WithJWT(h.jwt, h.stubUserUnsubscribe()))
		r.PUT("/:id", server.WithJWT(h.jwt, h.stubUserUpdate()))
		r.PATCH("/:id/status", server.WithJWT(h.jwt, h.stubUserStatusUpdate()))
	}

	// ================================
	// 15. Playlist — MOVED to PlaylistHandler + MeHandler + AdminHandler
	// PlaylistHandler: GET /playlists, GET /playlists/:token (portal, short_token)
	// MeHandler: GET/POST/PATCH/DELETE /me/playlists (user CRUD)
	// AdminHandler: GET/POST/PUT/DELETE /admin/playlists (admin CRUD)
	// ================================

	// ================================
	// 16. Encoding status/events (public aliases)
	// ================================
	encoding := rg.Group("/encoding")
	{
		r := ginadapter.NewStdRouterAdapter(encoding)
		r.GET("/status", h.stubEncodingStatus())
		r.GET("/events", h.stubEncodingEvents())
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

func (h *StubHandler) stubNotificationDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{"code": 0, "message": "deleted"})
	}
}

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
