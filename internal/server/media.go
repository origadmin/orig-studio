/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Media module - handles media CRUD, upload, and interactions (likes/favorites/shares)
 *
 * API paths:
 * - /api/v1/medias              - media collection
 * - /api/v1/medias/upload       - file upload
 * - /api/v1/medias/:id          - single media
 * - /api/v1/medias/:id/likes     - like operations
 * - /api/v1/medias/:id/favorites - favorite operations
 * - /api/v1/medias/:id/shares    - share operations
 */

// Package server provides HTTP handlers for media CRUD + upload.
package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"origadmin/application/origcms/internal/auth"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// UploadDir is the directory where uploaded media files are stored.
const UploadDir = "data/uploads"

// Upload limits
const (
	MaxUploadSizeVideo = 5 << 30   // 5 GB
	MaxUploadSizeImage = 50 << 20  // 50 MB
	MaxUploadSizeAudio = 500 << 20 // 500 MB
	MaxUploadSizeOther = 100 << 20 // 100 MB
)

// AllowedMIMEByType maps media type prefix → allowed MIME types.
var AllowedMIMEByType = map[string][]string{
	"video": {
		"video/mp4", "video/webm", "video/ogg", "video/quicktime",
		"video/x-msvideo", "video/x-matroska", "video/x-flv",
	},
	"image": {
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"image/svg+xml", "image/bmp", "image/tiff",
	},
	"audio": {
		"audio/mpeg", "audio/ogg", "audio/wav", "audio/flac",
		"audio/aac", "audio/webm", "audio/x-m4a",
	},
}

// detectMediaType maps MIME type to media type string (video/image/audio).
func detectMediaType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	default:
		return "video" // default fallback
	}
}

// maxUploadSizeByType returns the max upload size for a given media type.
func maxUploadSizeByType(mediaType string) int64 {
	switch mediaType {
	case "video":
		return MaxUploadSizeVideo
	case "image":
		return MaxUploadSizeImage
	case "audio":
		return MaxUploadSizeAudio
	default:
		return MaxUploadSizeOther
	}
}

// isMIMEAllowed checks if the MIME type is in the allowed list for the given media type.
func isMIMEAllowed(mimeType, mediaType string) bool {
	allowed, ok := AllowedMIMEByType[mediaType]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == mimeType {
			return true
		}
	}
	return false
}

// computeFileMD5 reads the file and returns its MD5 hash.
func computeFileMD5(r io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MediaHandler handles media requests.
type MediaHandler struct {
	jwtMgr         *auth.Manager
	uc             *biz.MediaUseCase
	uploadUC       *biz.UploadUseCase
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase
}

func NewMediaHandler(
	jwtMgr *auth.Manager,
	uc *biz.MediaUseCase,
	uploadUC *biz.UploadUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
) *MediaHandler {
	return &MediaHandler{jwtMgr: jwtMgr, uc: uc, uploadUC: uploadUC, likeFavoriteUC: likeFavoriteUC}
}

func (h *MediaHandler) Register(group *gin.RouterGroup) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0o755); err != nil {
		slog.Warn("failed to create upload directory", "err", err)
	}

	media := group.Group("/medias")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		// Upload
		media.POST("/upload", JWTMiddleware(h.jwtMgr), h.uploadMedia())

		// Encoding Profiles (Subgroup)
		profiles := media.Group("/profiles")
		{
			profiles.GET("", h.listEncodeProfiles())
			profiles.GET("/:profile_id", h.getEncodeProfile())
			profiles.POST("", JWTMiddleware(h.jwtMgr), h.createEncodeProfile())
			profiles.PUT("/:profile_id", JWTMiddleware(h.jwtMgr), h.updateEncodeProfile())
			profiles.DELETE("/:profile_id", JWTMiddleware(h.jwtMgr), h.deleteEncodeProfile())
		}

		// Transcoding & Encoding Status
		encoding := media.Group("/encoding")
		{
			encoding.GET("/status", h.getTranscodingStatus())
			encoding.GET("/tasks", h.getEncodingTasksFlat())
			encoding.POST("/retry", JWTMiddleware(h.jwtMgr), h.retryTaskByID())
			encoding.POST("/retry-failed", JWTMiddleware(h.jwtMgr), h.retryAllFailed())
		}

		// Collection Route
		media.GET("", h.listMedia())

		// ================================
		// 2. NESTED RESOURCE ROUTES (WITH :id) - MUST BE BEFORE MAIN :id ROUTES
		// ================================
		// Media Variants
		media.GET("/:id/variants", h.getMediaVariants())

		// Media Tasks & Retry
		media.GET("/:id/tasks", h.listEncodingTasks())
		media.POST("/:id/tasks/:taskId/retry", JWTMiddleware(h.jwtMgr), h.retryTranscode())

		// Like & Dislike
		media.GET("/:id/likes", h.getLikeStatus())
		media.POST("/:id/likes", JWTMiddleware(h.jwtMgr), h.toggleLike())
		media.DELETE("/:id/likes", JWTMiddleware(h.jwtMgr), h.toggleDislike())

		// Favorite
		media.GET("/:id/favorites", h.getFavoriteStatus())
		media.POST("/:id/favorites", JWTMiddleware(h.jwtMgr), h.toggleFavorite())
		media.DELETE("/:id/favorites", JWTMiddleware(h.jwtMgr), h.toggleFavorite())

		// Share
		media.GET("/:id/shares", h.getShareUrl())
		media.POST("/:id/shares", JWTMiddleware(h.jwtMgr), h.recordShare())

		// ================================
		// 3. MAIN RESOURCE PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		// Media CRUD Operations
		media.GET("/:id", h.getMedia())
		media.PUT("/:id", JWTMiddleware(h.jwtMgr), h.updateMedia())
		media.DELETE("/:id", JWTMiddleware(h.jwtMgr), h.deleteMedia())
	}
}

// --- List Media ---

func (h *MediaHandler) listMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		opt := &dto.MediaQueryOption{
			State:      c.Query("state"),
			MediaType:  c.Query("type"),
			OrderBy:    c.DefaultQuery("order_by", "created_at"),
			Descending: c.DefaultQuery("descending", "true") == "true",
		}
		opt.Page = int32(page)
		opt.PageSize = int32(pageSize)
		opt.Keyword = c.Query("keyword")

		if userIDStr := c.Query("user_id"); userIDStr != "" {
			if userID, err := strconv.Atoi(userIDStr); err == nil {
				v := int64(userID)
				opt.UserID = &v
			}
		}

		if catIDStr := c.Query("category_id"); catIDStr != "" {
			if catID, err := strconv.Atoi(catIDStr); err == nil {
				v := int64(catID)
				opt.CategoryID = &v
			}
		}

		if c.Query("featured") == "true" {
			v := true
			opt.Featured = &v
		}

		items, total, err := h.uc.ListMedias(ctx, opt)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		// Convert timestamps to ISO 8601 format
		for _, item := range items {
			if item.CreateTime != nil {
				// Timestamps are already in protobuf format, which will be serialized correctly
			}
			if item.UpdateTime != nil {
				// Timestamps are already in protobuf format, which will be serialized correctly
			}
		}

		OK(c, gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *MediaHandler) getMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		m, err := h.uc.GetMedia(ctx, int64(id))
		if err != nil {
			Fail(c, ErrMediaNotFound, "Media not found")
			return
		}

		go func() {
			bgCtx := context.Background()
			h.uc.IncrementViewCount(bgCtx, int64(id))
		}()

		OK(c, m)
	}
}

func (h *MediaHandler) uploadMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSizeVideo)

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			Fail(c, ErrBadRequest, "Failed to get file from request")
			return
		}
		defer file.Close()

		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mimeType := http.DetectContentType(buf[:n])
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		clientMIME := header.Header.Get("Content-Type")
		if clientMIME != "" && clientMIME != "application/octet-stream" {
			if mimeType == "application/octet-stream" {
				mimeType = clientMIME
			}
		}

		mediaType := detectMediaType(mimeType)

		if !isMIMEAllowed(mimeType, mediaType) {
			Fail(c, ErrBadRequest, fmt.Sprintf("File type %s is not allowed for %s", mimeType, mediaType))
			return
		}

		maxSize := maxUploadSizeByType(mediaType)
		if header.Size > maxSize {
			Fail(c, ErrBadRequest, fmt.Sprintf("File too large for %s (max %d bytes)", mediaType, maxSize))
			return
		}

		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		fileMD5, err := computeFileMD5(file)
		if err != nil {
			slog.Warn("failed to compute MD5", "err", err)
			fileMD5 = ""
		}

		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = mimeToExt(mimeType)
		}
		newFilename := uuid.New().String() + ext
		// Store in 'uploads' sub-dir to match Register routes
		relativePath := "uploads/" + newFilename
		filePath := filepath.Join(UploadDir, "uploads", newFilename)
		_ = os.MkdirAll(filepath.Dir(filePath), 0o755)

		out, err := os.Create(filePath)
		if err != nil {
			Fail(c, ErrInternal, "Failed to save file: "+err.Error())
			return
		}
		defer out.Close()

		written, err := io.Copy(out, file)
		if err != nil {
			os.Remove(filePath)
			Fail(c, ErrInternal, "Failed to write file")
			return
		}

		fileURL := relativePath

		title := c.PostForm("title")
		description := c.PostForm("description")
		categoryIDStr := c.PostForm("category_id")
		tagsStr := c.PostForm("tags")
		privacyStr := c.DefaultPostForm("privacy", "1")

		if title == "" {
			title = strings.TrimSuffix(header.Filename, ext)
		}

		var categoryID int
		if categoryIDStr != "" {
			categoryID, _ = strconv.Atoi(categoryIDStr)
		}

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		privacy, _ := strconv.Atoi(privacyStr)

		m := &biz.Media{
			Title:          title,
			Description:    description,
			Type:           mediaType,
			Url:            fileURL,
			UserId:         int64(claims.UserID),
			State:          "active",
			EncodingStatus: "pending",
			MimeType:       mimeType,
			Md5Sum:         fileMD5,
			Size:           written,
			Extension:      strings.TrimPrefix(ext, "."),
			Privacy:        int32(privacy),
			Tags:           tags,
			CategoryId:     int64(categoryID),
		}

		created, err := h.uc.CreateMedia(ctx, m)
		if err != nil {
			os.Remove(filePath)
			Fail(c, ErrInternal, "Failed to save media record: "+err.Error())
			return
		}

		// Re-fetch to get user and category details
		created, _ = h.uc.GetMedia(ctx, created.Id)

		c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: created})
	}
}

// updateMediaRequest is the JSON body for PUT /media/:id
type updateMediaRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CategoryID  *int     `json:"category_id"`
	Tags        []string `json:"tags"`
	Privacy     *int     `json:"privacy"`
	State       *string  `json:"state"`
	Featured    *bool    `json:"featured"`
}

func (h *MediaHandler) updateMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		m, err := h.uc.GetMedia(ctx, int64(id))
		if err != nil {
			Fail(c, ErrMediaNotFound, "Media not found")
			return
		}

		if m.UserId != int64(claims.UserID) && !claims.IsStaff {
			Fail(c, ErrForbidden, "you can only edit your own media")
			return
		}

		var req updateMediaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if req.Title != "" {
			m.Title = req.Title
		}
		if req.Description != "" {
			m.Description = req.Description
		}
		if req.CategoryID != nil {
			m.CategoryId = int64(*req.CategoryID)
		}
		if req.Tags != nil {
			m.Tags = req.Tags
		}
		if req.Privacy != nil {
			m.Privacy = int32(*req.Privacy)
		}
		if req.State != nil {
			m.State = *req.State
		}
		if req.Featured != nil {
			m.Featured = *req.Featured
		}

		updated, err := h.uc.UpdateMedia(ctx, m)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		// Re-fetch to get full details
		updated, _ = h.uc.GetMedia(ctx, int64(id))

		OK(c, updated)
	}
}

func (h *MediaHandler) deleteMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		m, err := h.uc.GetMedia(ctx, int64(id))
		if err != nil {
			Fail(c, ErrMediaNotFound, "Media not found")
			return
		}

		if m.UserId != int64(claims.UserID) && !claims.IsStaff {
			Fail(c, ErrForbidden, "you can only delete your own media")
			return
		}

		if m.Url != "" {
			filename := filepath.Base(m.Url)
			_ = os.Remove(filepath.Join(UploadDir, "uploads", filename))
		}

		if err := h.uc.DeleteMedia(ctx, int64(id)); err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{"message": "deleted"})
	}
}

func (h *MediaHandler) listEncodingTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		tasks, err := h.uc.ListEncodingTasks(c.Request.Context(), id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, tasks)
	}
}

func (h *MediaHandler) retryTranscode() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			Fail(c, ErrBadRequest, "invalid media ID")
			return
		}

		if err := h.uploadUC.RetryTranscode(c.Request.Context(), id); err != nil {
			if strings.Contains(err.Error(), "cannot retry") ||
				strings.Contains(err.Error(), "not found") {
				Fail(c, ErrBadRequest, err.Error())
			} else {
				Fail(c, ErrInternal, err.Error())
			}
			return
		}

		OK(c, gin.H{"message": "transcode retry initiated", "media_id": id})
	}
}

func (h *MediaHandler) getTranscodingStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := &biz.TranscodingStatusFilter{
			Status:   c.DefaultQuery("status", "active"),
			Page:     1,
			PageSize: 20,
		}

		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p >= 1 {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && ps >= 1 &&
			ps <= 100 {
			filter.PageSize = ps
		}

		status, err := h.uc.GetTranscodingStatus(c.Request.Context(), filter)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		// Build clean response to avoid protobuf field leakage.
		items := make([]gin.H, 0, len(status.Items))
		for _, item := range status.Items {
			m := item.Media
			items = append(items, gin.H{
				"media": gin.H{
					"id":              m.Id,
					"title":           m.Title,
					"description":     m.Description,
					"type":            m.Type,
					"url":             m.Url,
					"thumbnail":       m.Thumbnail,
					"hls_file":        m.HlsFile,
					"encoding_status": m.EncodingStatus,
					"duration":        m.Duration,
					"size":            m.Size,
					"mime_type":       m.MimeType,
					"user_id":         m.UserId,
					"view_count":      m.ViewCount,
					"like_count":      m.LikeCount,
					"create_time":     m.CreateTime,
					"update_time":     m.UpdateTime,
				},
				"tasks": item.Tasks,
			})
		}

		OK(c, gin.H{
			"processing_count": status.ProcessingCount,
			"pending_count":    status.PendingCount,
			"failed_count":     status.FailedCount,
			"success_count":    status.SuccessCount,
			"total_filtered":   status.TotalFiltered,
			"page":             status.Page,
			"page_size":        status.PageSize,
			"items":            items,
		})
	}
}

// getEncodingTasksFlat returns a flat, paginated list of encoding tasks (one row per task).
// Query params: status, page, page_size, media_id
func (h *MediaHandler) getEncodingTasksFlat() gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := &biz.TranscodingStatusFilter{
			Status:   c.DefaultQuery("status", "all"),
			Page:     1,
			PageSize: 25,
		}

		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p >= 1 {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "25")); err == nil && ps >= 1 &&
			ps <= 100 {
			filter.PageSize = ps
		}

		var mediaID *int64
		if m := c.Query("media_id"); m != "" {
			var id int64
			if _, err := fmt.Sscanf(m, "%d", &id); err == nil && id > 0 {
				mediaID = &id
			}
		}

		result, err := h.uc.ListEncodingTasksFlat(c.Request.Context(), filter, mediaID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, result)
	}
}

// getMediaVariants returns aggregated transcoding status for a single media.
// Used by the media management page to show variant details.
func (h *MediaHandler) getMediaVariants() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			Fail(c, ErrBadRequest, "invalid media ID")
			return
		}

		summary, err := h.uc.GetMediaVariants(c.Request.Context(), id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				Fail(c, ErrMediaNotFound, "media not found")
				return
			}
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, summary)
	}
}

// retryTaskByID retries a specific failed encoding task by task_id query param.
func (h *MediaHandler) retryTaskByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskIDStr := c.Query("task_id")
		if taskIDStr == "" {
			Fail(c, ErrBadRequest, "task_id is required")
			return
		}
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil || taskID <= 0 {
			Fail(c, ErrBadRequest, "invalid task_id: must be a positive integer")
			return
		}

		task, err := h.uc.RetryTask(c.Request.Context(), taskID)
		if err != nil {
			if strings.Contains(err.Error(), "cannot retry") ||
				strings.Contains(err.Error(), "not found") {
				Fail(c, ErrBadRequest, err.Error())
			} else {
				Fail(c, ErrInternal, err.Error())
			}
			return
		}

		OK(c, gin.H{"message": "retry queued", "task": task})
	}
}

// retryAllFailed retries all failed encoding tasks.
func (h *MediaHandler) retryAllFailed() gin.HandlerFunc {
	return func(c *gin.Context) {
		mediaIdStr := c.Query("media_id")
		var mediaID int64
		if mediaIdStr != "" {
			if _, err := fmt.Sscanf(mediaIdStr, "%d", &mediaID); err != nil {
				Fail(c, ErrBadRequest, "invalid media_id")
				return
			}
		}

		count, err := h.uc.RetryAllFailedTasks(c.Request.Context(), mediaID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{"message": "retry initiated", "retried_count": count})
	}
}

func (h *MediaHandler) transcodingEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		mediaIdStr := c.Query("media_id")
		var mediaID int64
		if mediaIdStr != "" {
			fmt.Sscanf(mediaIdStr, "%d", &mediaID)
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		ctx := c.Request.Context()
		events, cleanup := h.uc.Subscribe(ctx, mediaID)
		defer cleanup()

		c.Stream(func(w io.Writer) bool {
			select {
			case <-ctx.Done():
				return false
			case ev, ok := <-events:
				if !ok {
					return false
				}
				c.SSEvent("transcoding_progress", gin.H{
					"media_id": ev.MediaId,
					"task_id":  ev.Task.Id,
					"status":   ev.Task.Status,
					"progress": ev.Task.Progress,
				})
				return true
			}
		})
	}
}

// --- Encode Profile CRUD ---

func (h *MediaHandler) listEncodeProfiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		profiles, err := h.uc.ListEncodeProfiles(c.Request.Context())
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}
		OK(c, gin.H{"profiles": profiles})
	}
}

func (h *MediaHandler) getEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("profile_id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid Profile ID")
			return
		}
		p, err := h.uc.GetEncodeProfile(c.Request.Context(), id)
		if err != nil {
			Fail(c, ErrNotFound, "Profile not found")
			return
		}
		OK(c, gin.H{"profile": p})
	}
}

func (h *MediaHandler) createEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var profile biz.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}
		p, err := h.uc.CreateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}
		c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"profile": p}})
	}
}

func (h *MediaHandler) updateEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("profile_id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid Profile ID")
			return
		}
		var profile biz.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}
		profile.Id = id
		p, err := h.uc.UpdateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}
		OK(c, gin.H{"profile": p})
	}
}

func (h *MediaHandler) deleteEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("profile_id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid Profile ID")
			return
		}
		if err := h.uc.DeleteEncodeProfile(c.Request.Context(), id); err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}
		OK(c, gin.H{"message": "deleted"})
	}
}

// --- Like and Favorite Functions ---

func (h *MediaHandler) toggleLike() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, int(claims.UserID), id, "like")
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"is_liked":      stats.UserLikeType == "like",
			"is_disliked":   stats.UserLikeType == "dislike",
			"like_count":    stats.LikeCount,
			"dislike_count": stats.DislikeCount,
		})
	}
}

func (h *MediaHandler) toggleDislike() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, int(claims.UserID), id, "dislike")
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"is_liked":      stats.UserLikeType == "like",
			"is_disliked":   stats.UserLikeType == "dislike",
			"like_count":    stats.LikeCount,
			"dislike_count": stats.DislikeCount,
		})
	}
}

func (h *MediaHandler) getLikeStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var userID int
		if claims, ok := c.Get("claims"); ok {
			userID = int(claims.(*auth.Claims).UserID)
		}

		stats, err := h.likeFavoriteUC.GetMediaStats(ctx, userID, id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"is_liked":      stats.UserLikeType == "like",
			"is_disliked":   stats.UserLikeType == "dislike",
			"like_count":    stats.LikeCount,
			"dislike_count": stats.DislikeCount,
		})
	}
}

func (h *MediaHandler) toggleFavorite() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleFavorite(ctx, int(claims.UserID), id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"success":      true,
			"is_favorited": stats.IsFavorited,
		})
	}
}

func (h *MediaHandler) getFavoriteStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var userID int
		if claims, ok := c.Get("claims"); ok {
			userID = int(claims.(*auth.Claims).UserID)
		}

		stats, err := h.likeFavoriteUC.GetMediaStats(ctx, userID, id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"success":      true,
			"is_favorited": stats.IsFavorited,
		})
	}
}

// --- Helpers ---

func mimeToExt(mimeType string) string {
	exts := map[string]string{
		"video/mp4":        ".mp4",
		"video/webm":       ".webm",
		"video/ogg":        ".ogv",
		"video/quicktime":  ".mov",
		"video/x-msvideo":  ".avi",
		"video/x-matroska": ".mkv",
		"image/jpeg":       ".jpg",
		"image/png":        ".png",
		"image/gif":        ".gif",
		"image/webp":       ".webp",
		"image/svg+xml":    ".svg",
		"image/bmp":        ".bmp",
		"audio/mpeg":       ".mp3",
		"audio/ogg":        ".ogg",
		"audio/wav":        ".wav",
		"audio/flac":       ".flac",
		"audio/aac":        ".aac",
	}
	if ext, ok := exts[mimeType]; ok {
		return ext
	}
	return ""
}

// --- Share Functions ---

func (h *MediaHandler) getShareUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// Build share URL - assuming the frontend is at /watch/:id
		shareUrl := c.Request.Host + "/watch?v=" + strconv.Itoa(id)
		// Add https:// if not present
		if len(shareUrl) > 0 && shareUrl[0] != 'h' {
			shareUrl = "https://" + shareUrl
		}

		OK(c, gin.H{
			"url": shareUrl,
		})
	}
}

func (h *MediaHandler) recordShare() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("claims")
		if !exists {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		_, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// TODO: Implement share count increment in the future
		// For now, just return success

		OK(c, gin.H{
			"success": true,
		})
	}
}
