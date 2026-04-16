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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/helpers/ffmpeg"
	"origadmin/application/origcms/internal/helpers/repo"
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

func (h *MediaHandler) Register(r handler.Router) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0o755); err != nil {
		slog.Warn("failed to create upload directory", "err", err)
	}

	// ================================
	// 1. Independent fixed paths - no conflict with variable paths
	// ================================

	// Media upload
	r.POST("/medias/upload", WithJWT(h.jwtMgr, h.uploadMediaHandler))

	// Media list
	r.GET("/medias", h.listMediaHandler)

	// Encoding Profiles - independent path to avoid conflict with /medias/:id
	r.GET("/encoding/profiles", h.listEncodeProfilesHandler)
	r.GET("/encoding/profiles/:profile_id", h.getEncodeProfileHandler)
	r.POST("/encoding/profiles", WithJWT(h.jwtMgr, h.createEncodeProfileHandler))
	r.PUT("/encoding/profiles/:profile_id", WithJWT(h.jwtMgr, h.updateEncodeProfileHandler))
	r.DELETE("/encoding/profiles/:profile_id", WithJWT(h.jwtMgr, h.deleteEncodeProfileHandler))

	// Transcoding & Encoding Status - independent path
	r.GET("/encoding/tasks", h.getEncodingTasksFlatHandler)
	r.GET("/encoding/events", h.transcodingEventsHandler)
	r.POST("/encoding/retry", WithJWT(h.jwtMgr, h.retryTaskByIDHandler))
	r.POST("/encoding/retry-all-failed", WithJWT(h.jwtMgr, h.retryAllFailedHandler))

	// ================================
	// 2. Media resource paths - with variable parameters
	// ================================
	media := r.Group("/media")
	{
		// Media CRUD Operations
		media.GET("/:id", h.getMediaHandler)
		media.PUT("/:id", WithJWT(h.jwtMgr, h.updateMediaHandler))
		media.DELETE("/:id", WithJWT(h.jwtMgr, h.deleteMediaHandler))

		// Media Variants
		media.GET("/:id/variants", h.getMediaVariantsHandler)

		// Media Tasks & Retry
		media.GET("/:id/tasks", h.listEncodingTasksHandler)
		media.POST("/:id/tasks/:taskId/retry", WithJWT(h.jwtMgr, h.retryTranscodeHandler))

		// Like & Dislike
		media.GET("/:id/likes", h.getLikeStatusHandler)
		media.POST("/:id/likes", WithJWT(h.jwtMgr, h.toggleLikeHandler))
		media.DELETE("/:id/likes", WithJWT(h.jwtMgr, h.toggleDislikeHandler))

		// Favorite
		media.GET("/:id/favorites", h.getFavoriteStatusHandler)
		media.POST("/:id/favorites", WithJWT(h.jwtMgr, h.toggleFavoriteHandler))
		media.DELETE("/:id/favorites", WithJWT(h.jwtMgr, h.toggleFavoriteHandler))

		// Share
		media.GET("/:id/shares", h.getShareUrlHandler)
		media.POST("/:id/shares", WithJWT(h.jwtMgr, h.recordShareHandler))
	}

	// ================================
	// 3. Medias resource paths - compatible with frontend requests
	// ================================
	medias := r.Group("/medias")
	{
		// Media CRUD Operations
		medias.GET("/:id", h.getMediaHandler)
		medias.PUT("/:id", WithJWT(h.jwtMgr, h.updateMediaHandler))
		medias.DELETE("/:id", WithJWT(h.jwtMgr, h.deleteMediaHandler))

		// Media Variants
		medias.GET("/:id/variants", h.getMediaVariantsHandler)

		// Media Tasks & Retry
		medias.GET("/:id/tasks", h.listEncodingTasksHandler)
		medias.POST("/:id/tasks/:taskId/retry", WithJWT(h.jwtMgr, h.retryTranscodeHandler))

		// Like & Dislike
		medias.GET("/:id/likes", h.getLikeStatusHandler)
		medias.POST("/:id/likes", WithJWT(h.jwtMgr, h.toggleLikeHandler))
		medias.DELETE("/:id/likes", WithJWT(h.jwtMgr, h.toggleDislikeHandler))

		// Favorite
		medias.GET("/:id/favorites", h.getFavoriteStatusHandler)
		medias.POST("/:id/favorites", WithJWT(h.jwtMgr, h.toggleFavoriteHandler))
		medias.DELETE("/:id/favorites", WithJWT(h.jwtMgr, h.toggleFavoriteHandler))

		// Share
		medias.GET("/:id/shares", h.getShareUrlHandler)
		medias.POST("/:id/shares", WithJWT(h.jwtMgr, h.recordShareHandler))
	}
}

// --- HTTP Handlers ---

func (h *MediaHandler) listMediaHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	opt := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  c.Query("keyword"),
		},
		State:      c.Query("state"),
		MediaType:  c.Query("type"),
		OrderBy:    c.Query("order_by"),
		Descending: c.Query("descending") == "true",
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		opt.UserID = &userIDStr
	}

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		opt.CategoryID = &catIDStr
	}

	if c.Query("featured") == "true" {
		v := true
		opt.Featured = &v
	}

	// Handle tags filtering
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		opt.Tags = tags
	}

	items, total, err := h.uc.ListMedias(ctx, opt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *MediaHandler) getMediaHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	m, err := h.uc.GetMedia(ctx, idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	go func() {
		bgCtx := context.Background()
		h.uc.IncrementViewCount(bgCtx, idStr)
	}()

	c.JSON(http.StatusOK, m)
}

func (h *MediaHandler) uploadMediaHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSizeVideo)

	file, header, err := r.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
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
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("File type %s is not allowed for %s", mimeType, mediaType)},
		)
		return
	}

	maxSize := maxUploadSizeByType(mediaType)
	if header.Size > maxSize {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("File too large for %s (max %d bytes)", mediaType, maxSize)},
		)
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
	relativePath := "uploads/" + newFilename
	filePath := filepath.Join(UploadDir, "uploads", newFilename)
	_ = os.MkdirAll(filepath.Dir(filePath), 0o755)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}
	defer out.Close()

	written, err := io.Copy(out, file)
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		return
	}

	fileURL := relativePath

	title := r.FormValue("title")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	tagsStr := r.FormValue("tags")
	privacyStr := r.FormValue("privacy")
	if privacyStr == "" {
		privacyStr = "1"
	}

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

	// For video files, detect if it's a short video based on duration
	if strings.HasPrefix(mimeType, "video/") {
		// Get video duration
		var duration time.Duration
		if d, err := ffmpeg.GetVideoDuration(ctx, filePath); err == nil {
			duration = d
			// Check if it's a short video (duration < 60 seconds)
			if duration.Seconds() < 60 {
				mediaType = "short_video"
			}
		}
	}

	m := &biz.Media{
		Title:          title,
		Description:    description,
		Type:           mediaType,
		Url:            fileURL,
		UserId:         claims.UserID,
		State:          "active",
		EncodingStatus: "pending",
		MimeType:       mimeType,
		Md5Sum:         fileMD5,
		Size:           written,
		Extension:      strings.TrimPrefix(ext, "."),
		Privacy:        int32(privacy),
		Tags:           tags,
		CategoryId:     strconv.Itoa(categoryID),
	}

	created, err := h.uc.CreateMedia(ctx, m)
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media record: " + err.Error()})
		return
	}

	// Re-fetch to get user and category details
	created, _ = h.uc.GetMedia(ctx, created.Id)

	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": created})
}

func (h *MediaHandler) updateMediaHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	m, err := h.uc.GetMedia(ctx, idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	if m.UserId != claims.UserID && !claims.IsStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own media"})
		return
	}

	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		CategoryID  *int     `json:"category_id"`
		Tags        []string `json:"tags"`
		Privacy     *int     `json:"privacy"`
		State       *string  `json:"state"`
		Featured    *bool    `json:"featured"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		m.Title = req.Title
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.CategoryID != nil {
		m.CategoryId = strconv.Itoa(*req.CategoryID)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-fetch to get full details
	updated, _ = h.uc.GetMedia(ctx, idStr)

	c.JSON(http.StatusOK, updated)
}

func (h *MediaHandler) deleteMediaHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	m, err := h.uc.GetMedia(ctx, idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	if m.UserId != claims.UserID && !claims.IsStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own media"})
		return
	}

	if m.Url != "" {
		filename := filepath.Base(m.Url)
		_ = os.Remove(filepath.Join(UploadDir, "uploads", filename))
	}

	if err := h.uc.DeleteMedia(ctx, idStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *MediaHandler) listEncodingTasksHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	tasks, err := h.uc.ListEncodingTasks(c.Context(), idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *MediaHandler) retryTranscodeHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	if err := h.uploadUC.RetryTranscode(c.Context(), idStr); err != nil {
		if strings.Contains(err.Error(), "cannot retry") ||
			strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transcode retry initiated", "media_id": idStr})
}

func (h *MediaHandler) getEncodingTasksFlatHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	status := c.Query("status")

	// Determine filter type based on status
	var filterType biz.FilterType
	var specificStatus string

	switch status {
	case "":
		filterType = biz.FilterTypeAll
	case "active":
		filterType = biz.FilterTypeActive
	case "all":
		filterType = biz.FilterTypeAll
	default:
		// Validate specific status
		parsedStatus := enums.ParseEncodingTaskStatus(status)
		if parsedStatus == enums.EncodingTaskStatusUnknown {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter"})
			return
		}
		filterType = biz.FilterTypeSpecific
		specificStatus = status
	}

	filter := &biz.TranscodingStatusFilter{
		FilterType:    filterType,
		Status:        specificStatus,
		Page:          1,
		PageSize:      25,
		OnlyStats:     false,
		ProfileFilter: c.Query("profile"),
		ChunkFilter:   c.Query("chunk"),
		SearchQuery:   c.Query("search"),
	}

	if os := c.Query("only_stats"); os == "true" {
		filter.OnlyStats = true
	}

	if p, err := strconv.Atoi(c.Query("page")); err == nil && p >= 1 {
		filter.Page = p
	} else {
		filter.Page = 1
	}
	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps >= 1 &&
		ps <= 100 {
		filter.PageSize = ps
	} else {
		filter.PageSize = 25
	}

	var mediaIDStr *string
	if m := c.Query("media_id"); m != "" {
		mediaIDStr = &m
	}

	result, err := h.uc.ListEncodingTasksFlat(c.Context(), filter, mediaIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MediaHandler) getMediaVariantsHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	summary, err := h.uc.GetMediaVariantsByUUID(c.Context(), idStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *MediaHandler) retryTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	task, err := h.uc.RetryTask(c.Context(), taskIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "cannot retry") ||
			strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "retry queued", "task": task})
}

func (h *MediaHandler) retryAllFailedHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	mediaIdStr := c.Query("media_id")

	count, err := h.uc.RetryAllFailedTasks(c.Context(), mediaIdStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "retry initiated", "retried_count": count})
}

func (h *MediaHandler) transcodingEventsHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	mediaID := c.Query("media_id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()
	events, cleanup := h.uc.Subscribe(ctx, mediaID)
	defer cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: transcoding_progress\n")
			fmt.Fprintf(w, "data: %s\n\n", `{
				"media_id": "`+ev.MediaId+`",
				"task_id": "`+ev.Task.Id+`",
				"status": "`+string(ev.Task.Status)+`",
				"progress": `+fmt.Sprintf("%f", ev.Progress)+`,
				"speed": "`+ev.Speed+`",
				"fps": `+fmt.Sprintf("%f", ev.Fps)+`,
				"time": "`+fmt.Sprintf("%f", ev.Time)+`"
			}`)
			w.(http.Flusher).Flush()
		}
	}
}

func (h *MediaHandler) listEncodeProfilesHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	profiles, err := h.uc.ListEncodeProfiles(c.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

func (h *MediaHandler) getEncodeProfileHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	id, err := strconv.Atoi(c.Param("profile_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
		return
	}
	p, err := h.uc.GetEncodeProfile(c.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": p})
}

func (h *MediaHandler) createEncodeProfileHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	var profile biz.EncodeProfile
	if err := c.Bind(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.uc.CreateEncodeProfile(c.Context(), &profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(
		http.StatusCreated,
		gin.H{"code": 0, "message": "ok", "data": gin.H{"profile": p}},
	)
}

func (h *MediaHandler) updateEncodeProfileHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	id, err := strconv.Atoi(c.Param("profile_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
		return
	}
	var profile biz.EncodeProfile
	if err := c.Bind(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile.Id = id
	p, err := h.uc.UpdateEncodeProfile(c.Context(), &profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": p})
}

func (h *MediaHandler) deleteEncodeProfileHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	id, err := strconv.Atoi(c.Param("profile_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
		return
	}
	if err := h.uc.DeleteEncodeProfile(c.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *MediaHandler) toggleLikeHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.UserID, id, "like")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_liked":      stats.UserLikeType == "like",
		"is_disliked":   stats.UserLikeType == "dislike",
		"like_count":    stats.LikeCount,
		"dislike_count": stats.DislikeCount,
	})
}

func (h *MediaHandler) toggleDislikeHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.UserID, id, "dislike")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_liked":      stats.UserLikeType == "like",
		"is_disliked":   stats.UserLikeType == "dislike",
		"like_count":    stats.LikeCount,
		"dislike_count": stats.DislikeCount,
	})
}

func (h *MediaHandler) getLikeStatusHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var userID string
	if val := c.Get("claims"); val != nil {
		userID = val.(*auth.Claims).UserID
	}

	stats, err := h.likeFavoriteUC.GetMediaStats(ctx, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_liked":      stats.UserLikeType == "like",
		"is_disliked":   stats.UserLikeType == "dislike",
		"like_count":    stats.LikeCount,
		"dislike_count": stats.DislikeCount,
	})
}

func (h *MediaHandler) toggleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	stats, err := h.likeFavoriteUC.ToggleFavorite(ctx, claims.UserID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"is_favorited": stats.IsFavorited,
	})
}

func (h *MediaHandler) getFavoriteStatusHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var userID string
	if val := c.Get("claims"); val != nil {
		userID = val.(*auth.Claims).UserID
	}

	stats, err := h.likeFavoriteUC.GetMediaStats(ctx, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"is_favorited": stats.IsFavorited,
	})
}

func (h *MediaHandler) getShareUrlHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Build share URL - assuming the frontend is at /watch/:id
	shareUrl := r.Host + "/watch?v=" + id
	// Add https:// if not present
	if len(shareUrl) > 0 && shareUrl[0] != 'h' {
		shareUrl = "https://" + shareUrl
	}

	c.JSON(http.StatusOK, gin.H{
		"url": shareUrl,
	})
}

func (h *MediaHandler) recordShareHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// TODO: Implement share count increment in the future
	// For now, just return success

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
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
			opt.UserID = &userIDStr
		}

		if catIDStr := c.Query("category_id"); catIDStr != "" {
			opt.CategoryID = &catIDStr
		}

		if c.Query("featured") == "true" {
			v := true
			opt.Featured = &v
		}

		// Handle tags filtering
		if tagsStr := c.Query("tags"); tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			opt.Tags = tags
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

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		m, err := h.uc.GetMedia(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "Media not found")
			return
		}

		go func() {
			bgCtx := context.Background()
			h.uc.IncrementViewCount(bgCtx, idStr)
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
			Fail(
				c,
				ErrBadRequest,
				fmt.Sprintf("File type %s is not allowed for %s", mimeType, mediaType),
			)
			return
		}

		maxSize := maxUploadSizeByType(mediaType)
		if header.Size > maxSize {
			Fail(
				c,
				ErrBadRequest,
				fmt.Sprintf("File too large for %s (max %d bytes)", mediaType, maxSize),
			)
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

		// For video files, detect if it's a short video based on duration
		if strings.HasPrefix(mimeType, "video/") {
			// Get video duration
			var duration time.Duration
			if d, err := ffmpeg.GetVideoDuration(ctx, filePath); err == nil {
				duration = d
				// Check if it's a short video (duration < 60 seconds)
				if duration.Seconds() < 60 {
					mediaType = "short_video"
				}
			}
		}

		m := &biz.Media{
			Title:          title,
			Description:    description,
			Type:           mediaType,
			Url:            fileURL,
			UserId:         claims.UserID,
			State:          "active",
			EncodingStatus: "pending",
			MimeType:       mimeType,
			Md5Sum:         fileMD5,
			Size:           written,
			Extension:      strings.TrimPrefix(ext, "."),
			Privacy:        int32(privacy),
			Tags:           tags,
			CategoryId:     strconv.Itoa(categoryID),
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

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		m, err := h.uc.GetMedia(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "Media not found")
			return
		}

		if m.UserId != claims.UserID && !claims.IsStaff {
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
			m.CategoryId = strconv.Itoa(*req.CategoryID)
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
		updated, _ = h.uc.GetMedia(ctx, idStr)

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

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		m, err := h.uc.GetMedia(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "Media not found")
			return
		}

		if m.UserId != claims.UserID && !claims.IsStaff {
			Fail(c, ErrForbidden, "you can only delete your own media")
			return
		}

		if m.Url != "" {
			filename := filepath.Base(m.Url)
			_ = os.Remove(filepath.Join(UploadDir, "uploads", filename))
		}

		if err := h.uc.DeleteMedia(ctx, idStr); err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{"message": "deleted"})
	}
}

func (h *MediaHandler) listEncodingTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		tasks, err := h.uc.ListEncodingTasks(c.Request.Context(), idStr)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, tasks)
	}
}

func (h *MediaHandler) retryTranscode() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "invalid media ID")
			return
		}

		if err := h.uploadUC.RetryTranscode(c.Request.Context(), idStr); err != nil {
			if strings.Contains(err.Error(), "cannot retry") ||
				strings.Contains(err.Error(), "not found") {
				Fail(c, ErrBadRequest, err.Error())
			} else {
				Fail(c, ErrInternal, err.Error())
			}
			return
		}

		OK(c, gin.H{"message": "transcode retry initiated", "media_id": idStr})
	}
}

// getEncodingTasksFlat returns a flat, paginated list of encoding tasks (one row per task).
// Query params: status, page, page_size, media_id, only_stats, profile_filter, chunk_filter, search_query
// When only_stats=true, returns only statistics without items list (for cards)
func (h *MediaHandler) getEncodingTasksFlat() gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")

		// Determine filter type based on status
		var filterType biz.FilterType
		var specificStatus string

		switch status {
		case "":
			filterType = biz.FilterTypeAll
		case "active":
			filterType = biz.FilterTypeActive
		case "all":
			filterType = biz.FilterTypeAll
		default:
			// Validate specific status
			parsedStatus := enums.ParseEncodingTaskStatus(status)
			if parsedStatus == enums.EncodingTaskStatusUnknown {
				Fail(c, ErrBadRequest, "Invalid status parameter")
				return
			}
			filterType = biz.FilterTypeSpecific
			specificStatus = status
		}

		filter := &biz.TranscodingStatusFilter{
			FilterType:    filterType,
			Status:        specificStatus,
			Page:          1,
			PageSize:      25,
			OnlyStats:     false,
			ProfileFilter: c.Query("profile"),
			ChunkFilter:   c.Query("chunk"),
			SearchQuery:   c.Query("search"),
		}

		if os := c.Query("only_stats"); os == "true" {
			filter.OnlyStats = true
		}

		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p >= 1 {
			filter.Page = p
		}
		if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "25")); err == nil && ps >= 1 &&
			ps <= 100 {
			filter.PageSize = ps
		}

		var mediaIDStr *string
		if m := c.Query("media_id"); m != "" {
			mediaIDStr = &m
		}

		result, err := h.uc.ListEncodingTasksFlat(c.Request.Context(), filter, mediaIDStr)
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
		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "invalid media ID")
			return
		}

		summary, err := h.uc.GetMediaVariantsByUUID(c.Request.Context(), idStr)
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

		task, err := h.uc.RetryTask(c.Request.Context(), taskIDStr)
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

		count, err := h.uc.RetryAllFailedTasks(c.Request.Context(), mediaIdStr)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{"message": "retry initiated", "retried_count": count})
	}
}

func (h *MediaHandler) transcodingEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		mediaID := c.Query("media_id")

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
					"progress": ev.Progress,
					"speed":    ev.Speed,
					"fps":      ev.Fps,
					"time":     ev.Time,
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
		c.JSON(
			http.StatusCreated,
			Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"profile": p}},
		)
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

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.UserID, id, "like")
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

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.UserID, id, "dislike")
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

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var userID string
		if claims, ok := c.Get("claims"); ok {
			userID = claims.(*auth.Claims).UserID
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

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleFavorite(ctx, claims.UserID, id)
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

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var userID string
		if claims, ok := c.Get("claims"); ok {
			userID = claims.(*auth.Claims).UserID
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
		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// Build share URL - assuming the frontend is at /watch/:id
		shareUrl := c.Request.Host + "/watch?v=" + id
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

		id := c.Param("id")
		if id == "" {
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
