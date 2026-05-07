/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * SpriteHandler serves sprite sheet images and WebVTT files for video preview thumbnails.
 */

package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/features/media/biz"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
)

// SpriteHandler handles HTTP requests for sprite sheet and WebVTT files.
type SpriteHandler struct {
	mediaUC *biz.MediaUseCase
	baseDir string
	jwt     *auth.Manager
	logger  *log.Helper
}

// NewSpriteHandler creates a new SpriteHandler.
func NewSpriteHandler(mediaUC *biz.MediaUseCase, baseDir string, jwt *auth.Manager, logger log.Logger) *SpriteHandler {
	// Resolve baseDir to absolute path to avoid working directory dependency.
	// When the server is started from a different directory (e.g., framework root
	// instead of project root), relative paths like "./data/uploads" would
	// resolve incorrectly, causing file not found errors for sprite/VTT files.
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		absBaseDir = baseDir // fallback to original if resolution fails
	}
	return &SpriteHandler{
		mediaUC: mediaUC,
		baseDir: absBaseDir,
		jwt:     jwt,
		logger:  log.NewHelper(log.With(logger, "module", "service.sprite")),
	}
}

// GetSpriteVTT handles GET /medias/:id/sprite.vtt
// Returns the WebVTT file for sprite-based video preview thumbnails.
func (h *SpriteHandler) GetSpriteVTT(c *gin.Context) {
	shortToken := c.Param("id")
	if shortToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing media id"})
		return
	}

	info, err := h.mediaUC.GetSpriteInfoByShortToken(c.Request.Context(), shortToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	if info.Type != "video" {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite not available for non-video media"})
		return
	}

	if info.SpriteStatus != "success" {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite not available"})
		return
	}

	if info.VttPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite vtt path not set"})
		return
	}

	// Security: validate path to prevent directory traversal
	fullPath := filepath.Join(h.baseDir, info.VttPath)
	if err := validateSpritePath(h.baseDir, fullPath); err != nil {
		h.logger.Warnf("invalid vtt path for media %s: %v", shortToken, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		h.logger.Warnf("failed to read vtt file for media %s: %v", shortToken, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite vtt file not found"})
		return
	}

	c.Header("Content-Type", "text/vtt")
	c.Header("Cache-Control", "public, max-age=86400")
	c.String(http.StatusOK, string(data))
}

// GetSpriteImage handles GET /medias/:id/sprite.jpg
// Returns the JPEG sprite sheet image for video preview thumbnails.
func (h *SpriteHandler) GetSpriteImage(c *gin.Context) {
	shortToken := c.Param("id")
	if shortToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing media id"})
		return
	}

	info, err := h.mediaUC.GetSpriteInfoByShortToken(c.Request.Context(), shortToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	if info.Type != "video" {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite not available for non-video media"})
		return
	}

	if info.SpriteStatus != "success" {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite not available"})
		return
	}

	if info.SpritePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "sprite image path not set"})
		return
	}

	// Security: validate path to prevent directory traversal
	fullPath := filepath.Join(h.baseDir, info.SpritePath)
	if err := validateSpritePath(h.baseDir, fullPath); err != nil {
		h.logger.Warnf("invalid sprite path for media %s: %v", shortToken, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(fullPath)
}

// RegenerateSprite handles POST /admin/medias/:id/regenerate-sprite
// Triggers asynchronous sprite sheet and VTT regeneration for a video media.
func (h *SpriteHandler) RegenerateSprite(c *gin.Context) {
	mediaID := c.Param("id")
	if mediaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing media id"})
		return
	}

	info, err := h.mediaUC.GetSpriteInfoByID(c.Request.Context(), mediaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	if info.Type != "video" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot regenerate sprite for non-video media"})
		return
	}

	if info.SpriteStatus == "processing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sprite already processing"})
		return
	}

	// Trigger asynchronous regeneration.
	// IMPORTANT: Do NOT use c.Request.Context() here — it is cancelled as soon as
	// the HTTP response is sent, which would abort the ffmpeg subprocess via
	// exec.CommandContext. Use context.Background() so the regeneration runs to
	// completion independently of the request lifecycle.
	go func() {
		if err := h.mediaUC.RegenerateSprite(context.Background(), mediaID); err != nil {
			h.logger.Warnf("sprite regeneration failed for media %s: %v", mediaID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"media_id":      mediaID,
		"sprite_status": "pending",
		"message":       "sprite regeneration scheduled",
	})
}

// RegenerateThumbnail handles POST /admin/medias/:id/regenerate-thumbnail
// Triggers thumbnail regeneration for a video media at an optional timestamp.
func (h *SpriteHandler) RegenerateThumbnail(c *gin.Context) {
	mediaID := c.Param("id")
	if mediaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing media id"})
		return
	}

	var req struct {
		Timestamp *float64 `json:"timestamp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	timestamp := 0.0
	if req.Timestamp != nil {
		timestamp = *req.Timestamp
	}

	if err := h.mediaUC.RegenerateThumbnail(c.Request.Context(), mediaID, timestamp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("thumbnail regeneration failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media_id": mediaID,
		"message":  "thumbnail regenerated",
	})
}

// RegisterRoutes registers sprite-related routes on the given router group.
// This replaces the stub routes in StubHandler.
func (h *SpriteHandler) RegisterRoutes(r http2.Router) {
	// Public sprite routes (no auth required)
	medias := r.Group("/medias")
	{
		medias.GET("/:id/sprite.vtt", server.GinHandlerToHandlerFunc(h.GetSpriteVTT))
		medias.GET("/:id/sprite.jpg", server.GinHandlerToHandlerFunc(h.GetSpriteImage))
	}

	// Admin sprite/thumbnail regeneration routes (auth + admin required)
	adminMediaRegen := r.Group("/admin/medias/:id")
	if adapter, ok := adminMediaRegen.(*ginadapter.RouterAdapter); ok {
		adapter.UseGin(server.JWTMiddleware(h.jwt))
		adapter.UseGin(server.AdminMiddleware(h.jwt))
	}
	{
		adminMediaRegen.POST("/regenerate-sprite", server.GinHandlerToHandlerFunc(h.RegenerateSprite))
		adminMediaRegen.POST("/regenerate-thumbnail", server.GinHandlerToHandlerFunc(h.RegenerateThumbnail))
	}
}

// validateSpritePath ensures the resolved path is within the allowed base directory.
// This prevents directory traversal attacks.
func validateSpritePath(baseDir, targetPath string) error {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("failed to resolve base dir: %w", err)
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}
	// Check if the target path is within the base directory.
	if !strings.HasPrefix(absTarget, absBase+string(filepath.Separator)) && absTarget != absBase {
		return fmt.Errorf("path traversal detected: %s is outside %s", absTarget, absBase)
	}
	return nil
}
