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
		adminReview.GET("/pending", h.stubReviewPending())
		adminReview.GET("/history", h.stubReviewHistory())
		adminReview.POST("/batch", h.stubReviewBatch())
	}
	adminMediaReview := rg.Group("/admin/medias/:id")
	adminMediaReview.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		adminMediaReview.PUT("/review", h.stubReviewMedia())
		adminMediaReview.GET("/review-logs", h.stubReviewLogs())
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
		adminNavItems.GET("", h.stubNavItemList())
		adminNavItems.POST("", h.stubNavItemCreate())
		adminNavItems.PUT("/:id", h.stubNavItemUpdate())
		adminNavItems.DELETE("/:id", h.stubNavItemDelete())
		adminNavItems.PUT("/reorder", h.stubNavItemReorder())
	}

	// ================================
	// 5. Admin Banners
	// ================================
	adminBanners := rg.Group("/admin/banners")
	adminBanners.Use(server.JWTMiddleware(h.jwt), server.AdminMiddleware(h.jwt))
	{
		adminBanners.GET("", h.stubBannerList())
		adminBanners.POST("", h.stubBannerCreate())
		adminBanners.PUT("/:id", h.stubBannerUpdate())
		adminBanners.POST("/:id/toggle", h.stubBannerToggle())
	}

	// ================================
	// 6. Media Metadata / Sprite / Subtitle / Download/Stream/Thumbnail / Likes/Favorites/Shares / Update/Delete
	// ================================
	medias := rg.Group("/medias")
	{
		// Metadata
		medias.GET("/:id/metadata", h.stubMediaMetadata())
		medias.POST("/:id/metadata/mining", server.JWTMiddleware(h.jwt), h.stubMediaMetadataMining())
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
		medias.POST("/:id/subtitles", server.JWTMiddleware(h.jwt), h.stubSubtitleCreate())

		// Download/Stream/Thumbnail
		medias.GET("/:id/download", h.stubMediaDownload())
		medias.GET("/:id/stream", h.stubMediaStream())
		medias.GET("/:id/thumbnail", h.stubMediaThumbnail())

		// NOTE: Likes/Favorites plural paths are now handled by MediaHandler
		// (POST /:id/likes, GET /:id/likes, DELETE /:id/likes,
		//  POST /:id/favorites, GET /:id/favorites, DELETE /:id/favorites)

		// Shares
		medias.GET("/:id/shares", h.stubMediaShares())
		medias.POST("/:id/shares", server.JWTMiddleware(h.jwt), h.stubMediaShareCreate())

		// Update/Delete
		medias.PUT("/:id", server.JWTMiddleware(h.jwt), h.stubMediaUpdate())
		medias.DELETE("/:id", server.JWTMiddleware(h.jwt), h.stubMediaDelete())

		// Upload alias
		medias.POST("/upload", server.JWTMiddleware(h.jwt), h.stubMediaUpload())

		// Tasks (deprecated)
		medias.GET("/:id/tasks", h.stubMediaTasks())
		medias.POST("/:id/tasks/:taskId/retry", server.JWTMiddleware(h.jwt), h.stubMediaTaskRetry())
	}

	// ================================
	// 7. Subtitle (root level)
	// ================================
	subtitles := rg.Group("/subtitles")
	{
		subtitles.DELETE("/:id", server.JWTMiddleware(h.jwt), h.stubSubtitleDelete())
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
		adminStatsRevenue.GET("/revenue", h.stubAdminStatsRevenue())
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
		adminChannels.POST("", h.stubAdminChannelCreate())
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
		notifications.DELETE("/:id", h.stubNotificationDelete())
	}

	// ================================
	// 14. User subscription/update/status routes
	// ================================
	users := rg.Group("/users")
	{
		users.GET("/:id/subscription", h.stubUserSubscription())
		users.POST("/:id/subscribe", server.JWTMiddleware(h.jwt), h.stubUserSubscribe())
		users.DELETE("/:id/subscribe", server.JWTMiddleware(h.jwt), h.stubUserUnsubscribe())
		users.PUT("/:id", server.JWTMiddleware(h.jwt), h.stubUserUpdate())
		users.PATCH("/:id/status", server.JWTMiddleware(h.jwt), h.stubUserStatusUpdate())
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
		encoding.GET("/status", h.stubEncodingStatus())
		encoding.GET("/events", h.stubEncodingEvents())
	}
}

// ==================== Admin Media Stubs ====================

func (h *StubHandler) stubAdminMediaList() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		server.OK(c, gin.H{
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

func (h *StubHandler) stubAdminMediaGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": nil})
	}
}

func (h *StubHandler) stubAdminMediaUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": nil})
	}
}

func (h *StubHandler) stubAdminMediaDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok"})
	}
}

func (h *StubHandler) stubAdminMediaStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
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

func (h *StubHandler) stubAdminMediaVariants() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubAdminMediaState() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok"})
	}
}

func (h *StubHandler) stubAdminMediaTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubAdminMediaTaskRetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "retry initiated"})
	}
}

// ==================== Review Stubs ====================

func (h *StubHandler) stubReviewPending() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubReviewHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubReviewBatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "batch review completed"})
	}
}

func (h *StubHandler) stubReviewMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "review submitted"})
	}
}

func (h *StubHandler) stubReviewLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

// ==================== Nav Items Stubs ====================

func (h *StubHandler) stubNavItemList() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubNavItemCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubNavItemUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubNavItemDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "deleted"})
	}
}

func (h *StubHandler) stubNavItemReorder() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "reordered"})
	}
}

// ==================== Banner Stubs ====================

func (h *StubHandler) stubBannerList() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubBannerCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubBannerUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubBannerToggle() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "toggled"})
	}
}

// ==================== Metadata Stubs ====================

func (h *StubHandler) stubMediaMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"duration":    0,
			"resolution":  "",
			"codec":       "",
			"bitrate":     0,
			"frame_rate":  0,
			"audio_codec": "",
		}})
	}
}

func (h *StubHandler) stubMediaMetadataMining() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "mining started"})
	}
}

func (h *StubHandler) stubMediaMetadataStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"status":   "pending",
			"progress": 0,
		}})
	}
}

func (h *StubHandler) stubMediaMetadataKeyFrames() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubMediaMetadataAudioWaveform() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubMediaMetadataTextContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": ""})
	}
}

func (h *StubHandler) stubMediaMetadataSceneChanges() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

// ==================== Sprite Stubs ====================

func (h *StubHandler) stubSpriteVTT() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "WEBVTT\n\n")
	}
}

func (h *StubHandler) stubSpriteJPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	}
}

// ==================== Subtitle Stubs ====================

func (h *StubHandler) stubSubtitleList() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubSubtitleCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

func (h *StubHandler) stubSubtitleDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "deleted"})
	}
}

func (h *StubHandler) stubSubtitleLanguages() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

// ==================== Media Download/Stream/Thumbnail Stubs ====================

func (h *StubHandler) stubMediaDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.Fail(c, server.ErrNotFound, "download not available")
	}
}

func (h *StubHandler) stubMediaStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.Fail(c, server.ErrNotFound, "stream not available")
	}
}

func (h *StubHandler) stubMediaThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	}
}

// ==================== Regeneration Stubs ====================

func (h *StubHandler) stubRegenerateSprite() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "sprite regeneration started"})
	}
}

func (h *StubHandler) stubRegenerateThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "thumbnail regeneration started"})
	}
}

// ==================== Admin Stats Revenue Stub ====================

func (h *StubHandler) stubAdminStatsRevenue() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
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

func (h *StubHandler) stubAdminSettingsCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": []interface{}{}})
	}
}

func (h *StubHandler) stubAdminSettingsUpdateKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "updated"})
	}
}

func (h *StubHandler) stubAdminSettingsDeleteKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "deleted"})
	}
}

// ==================== Admin Channel Create Stub ====================

func (h *StubHandler) stubAdminChannelCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "created", "data": nil})
	}
}

// ==================== Notification Delete Stub ====================

func (h *StubHandler) stubNotificationDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "deleted"})
	}
}

// ==================== User Subscription Stubs ====================

func (h *StubHandler) stubUserSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"is_subscribed": false,
		}})
	}
}

func (h *StubHandler) stubUserSubscribe() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "subscribed"})
	}
}

func (h *StubHandler) stubUserUnsubscribe() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "unsubscribed"})
	}
}

// ==================== User Update/Status Stubs ====================

func (h *StubHandler) stubUserUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubUserStatusUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "status updated"})
	}
}

// ==================== Media Update/Delete Stubs ====================

func (h *StubHandler) stubMediaUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "updated", "data": nil})
	}
}

func (h *StubHandler) stubMediaDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "deleted"})
	}
}

// ==================== Media Likes/Favorites/Shares (plural) Stubs ====================

func (h *StubHandler) stubMediaLikes() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"is_liked":      false,
			"is_disliked":   false,
			"like_count":    0,
			"dislike_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaLikeToggle() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"is_liked":   true,
			"like_count": 1,
		})
	}
}

func (h *StubHandler) stubMediaLikeRemove() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"is_liked":   false,
			"like_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaFavorites() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"is_favorited":   false,
			"favorite_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaFavoriteToggle() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"is_favorited":   true,
			"favorite_count": 1,
		})
	}
}

func (h *StubHandler) stubMediaShares() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"share_count": 0,
		})
	}
}

func (h *StubHandler) stubMediaShareCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "shared"})
	}
}

// ==================== Media Upload Stub ====================

func (h *StubHandler) stubMediaUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "use /uploads/multipart for upload"})
	}
}

// ==================== Encoding Status/Events Stubs ====================

func (h *StubHandler) stubEncodingStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"total_tasks":  0,
			"pending":      0,
			"processing":   0,
			"completed":    0,
			"failed":       0,
			"success_rate": "0%",
		})
	}
}

func (h *StubHandler) stubEncodingEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"events": []interface{}{}})
	}
}

// ==================== Media Tasks (deprecated) Stubs ====================

func (h *StubHandler) stubMediaTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "ok", "data": gin.H{
			"items": []interface{}{},
			"total": 0,
		}})
	}
}

func (h *StubHandler) stubMediaTaskRetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{"code": 0, "message": "retry initiated"})
	}
}
