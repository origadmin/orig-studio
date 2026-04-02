// Package server provides HTTP handlers
package server

import (
	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/svc-media/biz"
)

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(router *gin.Engine, client *entity.Client, jwtMgr *auth.Manager, uploadUC *biz.UploadUseCase, mediaUC *biz.MediaUseCase) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	feedHandler := NewFeedHandler(client)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Feed route
		v1.GET("/feed", feedHandler.GetFeed)

		// User routes
		RegisterUserRoutes(v1, client)

		// Media routes (with JWT for upload/update/delete)
		RegisterMediaRoutes(v1, client, jwtMgr, mediaUC)

		// Category routes
		RegisterCategoryRoutes(v1, client)

		// Tag routes
		RegisterTagRoutes(v1, client)

		// Comment routes
		RegisterCommentRoutes(v1, client)

		// Playlist routes
		RegisterPlaylistRoutes(v1, client)

		// Upload routes (chunked multipart upload, requires JWT)
		RegisterUploadRoutes(v1, uploadUC, jwtMgr)
	}
}
