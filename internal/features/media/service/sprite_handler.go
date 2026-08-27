/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * SpriteHandler serves sprite sheet images and WebVTT files for video preview thumbnails.
 */

package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/features/media/biz"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

const spriteProcessingStuckTimeout = 5 * time.Minute

// SpriteHandler handles HTTP requests for sprite sheet and WebVTT files.
type SpriteHandler struct {
	mediaUC *biz.MediaUseCase
	paths   *conf.StoragePaths
	jwt     *auth.Manager
	logger  *log.Helper
}

// NewSpriteHandler creates a new SpriteHandler.
func NewSpriteHandler(mediaUC *biz.MediaUseCase, paths *conf.StoragePaths, jwt *auth.Manager, logger log.Logger) *SpriteHandler {
	return &SpriteHandler{
		mediaUC: mediaUC,
		paths:   paths,
		jwt:     jwt,
		logger:  log.NewHelper(log.With(logger, "module", "service.sprite")),
	}
}

// readSpriteAsset returns the bytes of a generated sprite asset, preferring the
// local disk copy and falling back to object storage (S3) when the file is not
// present on this node.
//
// Why the fallback is required: sprites are generated on whichever node ran the
// transcode and are always uploaded to storage (see
// media/biz.SpriteUseCase.GenerateSpriteAndVTT). A different node serving the
// read request has no local copy, so a local-only os.ReadFile answers 404 even
// though the asset exists. relPath is both the DB-stored relative path and the
// storage key ("sprites/{mediaID}/sprite.jpg|.vtt"), so it can be passed to
// Storage.Download directly - same contract the gateway StorageReadHandler uses.
func (h *SpriteHandler) readSpriteAsset(ctx context.Context, relPath, fullPath string) ([]byte, error) {
	if data, err := os.ReadFile(fullPath); err == nil {
		return data, nil
	}
	data, err := h.mediaUC.DownloadSprite(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("sprite asset %s missing locally and unreadable from storage: %w", relPath, err)
	}
	return data, nil
}

// resolveMediaToken returns the media short_token from the path parameter.
// It checks both :token and :id parameter names because RegisterSpriteRoutes
// registers both literal shapes (:token for owner links, :id for the gateway
// exact-pattern /api/v1/medias/{id}/sprite.{vtt,jpg} that mirrors proto).
func resolveMediaToken(ctx http2.Context) string {
	if t := ctx.Var("token"); t != "" {
		return t
	}
	if t := ctx.Var("id"); t != "" {
		return t
	}
	return ""
}

// GetSpriteVTT handles GET /medias/:token/sprite.vtt and GET /medias/:id/sprite.vtt
// Returns the WebVTT file for sprite-based video preview thumbnails.
func (h *SpriteHandler) GetSpriteVTT() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		shortToken := resolveMediaToken(ctx)
		if shortToken == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media id")
		}

		info, err := h.mediaUC.GetSpriteInfoByShortToken(ctx, shortToken)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		if info.Type != "video" {
			return server.FailCtx(ctx, server.ErrNotFound, "sprite not available for non-video media")
		}

		if info.SpriteStatus != "success" {
			return server.FailCtx(ctx, server.ErrNotFound, "sprite not available")
		}

		if info.VttPath == "" {
			return server.FailCtx(ctx, server.ErrNotFound, "sprite vtt path not set")
		}

		// Security: validate path to prevent directory traversal
		fullPath := h.paths.FullPath(info.VttPath)
		if err := validateSpritePath(h.paths.BasePath(), fullPath); err != nil {
			h.logger.Warnf("invalid vtt path for media %s: %v", shortToken, err)
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid path")
		}

		data, err := h.readSpriteAsset(ctx, info.VttPath, fullPath)
		if err != nil {
			h.logger.Warnf("failed to read vtt file for media %s: %v", shortToken, err)
			return server.FailCtx(ctx, server.ErrNotFound, "sprite vtt file not found")
		}

		ctx.Response().Header().Set("Content-Type", "text/vtt")
		ctx.Response().Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
		return ctx.String(http.StatusOK, string(data))
	}
}

// GetSpriteImage handles GET /medias/:token/sprite.jpg and GET /medias/:id/sprite.jpg
// Returns the JPEG sprite sheet image for video preview thumbnails.
func (h *SpriteHandler) GetSpriteImage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		shortToken := resolveMediaToken(ctx)
		if shortToken == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media id")
		}

		info, err := h.mediaUC.GetSpriteInfoByShortToken(ctx, shortToken)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		if info.Type != "video" {
			return server.FailCtx(ctx, server.ErrNotFound, "sprite not available for non-video media")
		}

		if info.SpriteStatus != "success" {
			return server.FailCtx(ctx, server.ErrNotFound, "sprite not available")
		}

		if info.SpritePath == "" {
			return server.FailCtx(ctx, server.ErrNotFound, "sprite image path not set")
		}

		// Security: validate path to prevent directory traversal
		fullPath := h.paths.FullPath(info.SpritePath)
		if err := validateSpritePath(h.paths.BasePath(), fullPath); err != nil {
			h.logger.Warnf("invalid sprite path for media %s: %v", shortToken, err)
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid path")
		}

		data, err := h.readSpriteAsset(ctx, info.SpritePath, fullPath)
		if err != nil {
			h.logger.Warnf("failed to read sprite image for media %s: %v", shortToken, err)
			return server.FailCtx(ctx, server.ErrNotFound, "sprite image file not found")
		}

		ctx.Response().Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
		return ctx.Blob(http.StatusOK, "image/jpeg", data)
	}
}

// RegenerateSprite handles POST /admin/medias/:id/regenerate-sprite
// Triggers asynchronous sprite sheet and VTT regeneration for a video media.
func (h *SpriteHandler) RegenerateSprite() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		mediaID := ctx.Var("id")
		if mediaID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media id")
		}

		info, err := h.mediaUC.GetSpriteInfoByID(ctx, mediaID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		if info.Type != "video" {
			return server.FailCtx(ctx, server.ErrBadRequest, "cannot regenerate sprite for non-video media")
		}

		if info.SpriteStatus == "processing" {
			if time.Since(info.UpdateTime) < spriteProcessingStuckTimeout {
				return server.FailCtx(ctx, server.ErrBadRequest, "sprite already processing")
			}
			h.logger.Warnf("sprite status stuck at processing for media %s (since %v), allowing retry",
				mediaID, info.UpdateTime)
		}

		// Trigger asynchronous regeneration.
		// IMPORTANT: Do NOT use ctx here — it is cancelled as soon as
		// the HTTP response is sent, which would abort the ffmpeg subprocess via
		// exec.CommandContext. Use context.Background() so the regeneration runs to
		// completion independently of the request lifecycle.
		go func() {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Errorf("sprite regeneration panicked for media %s: %v\n%s", mediaID, r, string(debug.Stack()))
				}
			}()
			bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
			defer cancel()
			if err := h.mediaUC.RegenerateSprite(bgCtx, mediaID); err != nil {
				h.logger.Warnf("sprite regeneration failed for media %s: %v", mediaID, err)
			}
		}()

		return server.OKCtx(ctx, map[string]any{
			"media_id":      mediaID,
			"sprite_status": "pending",
			"message":       "sprite regeneration scheduled",
		})
	}
}

// RegenerateThumbnail handles POST /admin/medias/:id/regenerate-thumbnail
// Triggers thumbnail regeneration for a video media at an optional timestamp.
func (h *SpriteHandler) RegenerateThumbnail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		mediaID := ctx.Var("id")
		if mediaID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media id")
		}

		var req struct {
			Timestamp *float64 `json:"timestamp"`
		}
		if err := ctx.BindJSON(&req); err != nil && err.Error() != "EOF" {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid request body")
		}

		timestamp := 0.0
		if req.Timestamp != nil {
			timestamp = *req.Timestamp
		}

		if err := h.mediaUC.RegenerateThumbnail(ctx, mediaID, timestamp); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, fmt.Sprintf("thumbnail regeneration failed: %v", err))
		}

		info, err := h.mediaUC.GetThumbnailInfoByID(ctx, mediaID)
		if err != nil {
			h.logger.Warnf("failed to get updated thumbnail info for media %s: %v", mediaID, err)
			return server.OKCtx(ctx, map[string]any{
				"media_id": mediaID,
				"message":  "thumbnail regenerated",
			})
		}

		return server.OKCtx(ctx, map[string]any{
			"media_id":       mediaID,
			"thumbnail":      info.Thumbnail,
			"thumbnail_time": info.ThumbnailTime,
			"message":        "thumbnail regenerated",
		})
	}
}

// OwnerRegenerateThumbnail handles POST /me/medias/:token/regenerate-thumbnail
// Allows the media owner to regenerate the thumbnail at a specific timestamp.
func (h *SpriteHandler) OwnerRegenerateThumbnail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		shortToken := ctx.Var("token")
		if shortToken == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media token")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		currentUserID := claims.GetUserID()

		info, err := h.mediaUC.GetSpriteInfoByShortToken(ctx, shortToken)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		if info.Type != "video" {
			return server.FailCtx(ctx, server.ErrBadRequest, "cannot regenerate thumbnail for non-video media")
		}

		if info.UserID != currentUserID {
			return server.FailCtx(ctx, server.ErrForbidden, "you do not own this media")
		}

		var req struct {
			Timestamp *float64 `json:"timestamp"`
		}
		if err := ctx.BindJSON(&req); err != nil && err.Error() != "EOF" {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid request body")
		}

		timestamp := 0.0
		if req.Timestamp != nil {
			timestamp = *req.Timestamp
		}

		if err := h.mediaUC.RegenerateThumbnail(ctx, info.ID, timestamp); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, fmt.Sprintf("thumbnail regeneration failed: %v", err))
		}

		thumbInfo, err := h.mediaUC.GetThumbnailInfoByID(ctx, info.ID)
		if err != nil {
			h.logger.Warnf("failed to get updated thumbnail info for media %s: %v", info.ID, err)
			return server.OKCtx(ctx, map[string]any{
				"media_id": info.ID,
				"message":  "thumbnail regenerated",
			})
		}

		return server.OKCtx(ctx, map[string]any{
			"media_id":       info.ID,
			"thumbnail":      thumbInfo.Thumbnail,
			"thumbnail_time": thumbInfo.ThumbnailTime,
			"message":        "thumbnail regenerated",
		})
	}
}

// OwnerUploadThumbnail handles POST /me/medias/:token/thumbnail
// Allows the media owner to upload a custom cover image.
func (h *SpriteHandler) OwnerUploadThumbnail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		shortToken := ctx.Var("token")
		if shortToken == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media token")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		currentUserID := claims.GetUserID()

		info, err := h.mediaUC.GetSpriteInfoByShortToken(ctx, shortToken)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		if info.UserID != currentUserID {
			return server.FailCtx(ctx, server.ErrForbidden, "you do not own this media")
		}

		file, header, err := ctx.FormFile("file")
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing file")
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return server.FailCtx(ctx, server.ErrBadRequest, "file must be an image")
		}

		tmpFile, err := os.CreateTemp("", "custom-thumb-*.jpg")
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, "failed to create temp file")
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := io.Copy(tmpFile, file); err != nil {
			tmpFile.Close()
			return server.FailCtx(ctx, server.ErrInternal, "failed to save uploaded file")
		}
		tmpFile.Close()

		if err := h.mediaUC.SetCustomThumbnail(ctx, info.ID, tmpPath); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, fmt.Sprintf("failed to set custom thumbnail: %v", err))
		}

		thumbInfo, err := h.mediaUC.GetThumbnailInfoByID(ctx, info.ID)
		if err != nil {
			h.logger.Warnf("failed to get updated thumbnail info for media %s: %v", info.ID, err)
			return server.OKCtx(ctx, map[string]any{
				"media_id": info.ID,
				"message":  "custom thumbnail uploaded",
			})
		}

		return server.OKCtx(ctx, map[string]any{
			"media_id":       info.ID,
			"thumbnail":      thumbInfo.Thumbnail,
			"thumbnail_time": thumbInfo.ThumbnailTime,
			"message":        "custom thumbnail uploaded",
		})
	}
}

// AdminUploadThumbnail handles POST /admin/medias/:id/thumbnail
// Allows an admin to upload a custom cover image for any media.
func (h *SpriteHandler) AdminUploadThumbnail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		mediaID := ctx.Var("id")
		if mediaID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing media id")
		}

		info, err := h.mediaUC.GetSpriteInfoByID(ctx, mediaID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "media not found")
		}

		file, header, err := ctx.FormFile("file")
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "missing file")
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return server.FailCtx(ctx, server.ErrBadRequest, "file must be an image")
		}

		tmpFile, err := os.CreateTemp("", "custom-thumb-*.jpg")
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, "failed to create temp file")
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := io.Copy(tmpFile, file); err != nil {
			tmpFile.Close()
			return server.FailCtx(ctx, server.ErrInternal, "failed to save uploaded file")
		}
		tmpFile.Close()

		if err := h.mediaUC.SetCustomThumbnail(ctx, info.ID, tmpPath); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, fmt.Sprintf("failed to set custom thumbnail: %v", err))
		}

		thumbInfo, err := h.mediaUC.GetThumbnailInfoByID(ctx, info.ID)
		if err != nil {
			h.logger.Warnf("failed to get updated thumbnail info for media %s: %v", info.ID, err)
			return server.OKCtx(ctx, map[string]any{
				"media_id": info.ID,
				"message":  "custom thumbnail uploaded",
			})
		}

		return server.OKCtx(ctx, map[string]any{
			"media_id":       info.ID,
			"thumbnail":      thumbInfo.Thumbnail,
			"thumbnail_time": thumbInfo.ThumbnailTime,
			"message":        "custom thumbnail uploaded",
		})
	}
}

// RegisterRoutes registers sprite-related routes on the given router group.
// This replaces the stub routes in StubHandler.
func (h *SpriteHandler) RegisterRoutes(r http2.Router) {
	// Public sprite routes (no auth required)
	medias := r.Group("/medias")
	{
		medias.GET("/:token/sprite.vtt", h.GetSpriteVTT())
		medias.GET("/:token/sprite.jpg", h.GetSpriteImage())
	}

	// Owner thumbnail routes (auth required, must own the media)
	ownerMedias := r.Group("/me/medias")
	ownerMedias.Use(server.JWTMiddlewareCtx(h.jwt))
	{
		ownerMedias.POST("/:token/regenerate-thumbnail", h.OwnerRegenerateThumbnail())
		ownerMedias.POST("/:token/thumbnail", h.OwnerUploadThumbnail())
	}

	// Admin sprite/thumbnail regeneration routes (auth + admin required)
	adminMediaRegen := r.Group("/admin/medias/:id")
	adminMediaRegen.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminMediaRegen.POST("/regenerate-sprite", h.RegenerateSprite())
		adminMediaRegen.POST("/regenerate-thumbnail", h.RegenerateThumbnail())
		adminMediaRegen.POST("/thumbnail", h.AdminUploadThumbnail())
	}
}

// RegisterSpriteRoutes registers ONLY the public sprite (VTT/JPEG) read routes.
// Used by the media microservice, which already owns the thumbnail and admin
// sprite/thumbnail regeneration routes (see thumbnail_handler.go and
// admin_media_service.go). The path matches the proto-defined
// GET /api/v1/medias/{id}/sprite.{vtt,jpg} so it can shadow the (unimplemented)
// proto handler registered by media.RegisterMediaServiceHTTPServer. It MUST be
// registered BEFORE the proto routes in the media HTTP server.
//
// We register BOTH :token and :id variants because:
//   - gateway service.go L239/L240 uses Handle("/api/v1/medias/{id}/sprite.vtt",
//     mediaProxy) so the mux path parameter is "id" (matches proto token too).
//   - frontends and the owner endpoints consistently use short_token style ("token"
//     path var) consumed by GetSpriteInfoByShortToken.
//   - gorilla/mux does not honour registration order for param routes with the
//     same literal shape, so whichever shape lands first can shadow the other.
//     By registering both literal shapes explicitly, the route is served by our
//     handler regardless of whether the incoming param was bound to "token" or "id".
func (h *SpriteHandler) RegisterSpriteRoutes(r http2.Router) {
	medias := r.Group("/medias")
	{
		// :token path (owner endpoints, legacy short_token links)
		medias.GET("/:token/sprite.vtt", h.GetSpriteVTT())
		medias.GET("/:token/sprite.jpg", h.GetSpriteImage())
		// :id path (gateway exact-pattern + media proto GetMediaSpriteVTT/JPG)
		medias.GET("/:id/sprite.vtt", h.GetSpriteVTT())
		medias.GET("/:id/sprite.jpg", h.GetSpriteImage())
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
