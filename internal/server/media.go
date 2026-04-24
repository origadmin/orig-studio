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
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/helpers/ffmpeg"
	"origadmin/application/origcms/internal/helpers/repo"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/dto"
	userbiz "origadmin/application/origcms/internal/svc-user/biz"
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
	jwtMgr                *auth.Manager
	uc                    *biz.MediaUseCase
	uploadUC              *biz.UploadUseCase
	likeFavoriteUC        *contentbiz.LikeFavoriteUseCase
	playlistChannelUC     *contentbiz.PlaylistChannelUseCase
	userUC                *userbiz.UserUseCase
	entityClient          *entity.Client
}

func NewMediaHandler(
	jwtMgr *auth.Manager,
	uc *biz.MediaUseCase,
	uploadUC *biz.UploadUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistChannelUC *contentbiz.PlaylistChannelUseCase,
	userUC *userbiz.UserUseCase,
	entityClient *entity.Client,
) *MediaHandler {
	return &MediaHandler{jwtMgr: jwtMgr, uc: uc, uploadUC: uploadUC, likeFavoriteUC: likeFavoriteUC, playlistChannelUC: playlistChannelUC, userUC: userUC, entityClient: entityClient}
}

func (h *MediaHandler) Register(r handler.Router) {
	// Keep Register for backward compatibility
}

// RegisterGin directly registers routes with gin.RouterGroup
func (h *MediaHandler) RegisterGin(rg *gin.RouterGroup) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0o755); err != nil {
		slog.Warn("failed to create upload directory", "err", err)
	}

	// ================================
	// 1. Independent fixed paths - no conflict with variable paths
	// ================================

	// Media upload (requires JWT)
	rg.POST("/medias/upload", JWTMiddleware(h.jwtMgr), h.uploadMedia())

	// Media list (public)
	rg.GET("/medias", h.listMedia())

	// Encoding Profiles - independent path to avoid conflict with /medias/:id
	rg.GET("/encoding/profiles", h.listEncodeProfiles())
	rg.GET("/encoding/profiles/:profile_id", h.getEncodeProfile())
	rg.POST("/encoding/profiles", JWTMiddleware(h.jwtMgr), h.createEncodeProfile())
	rg.PUT("/encoding/profiles/:profile_id", JWTMiddleware(h.jwtMgr), h.updateEncodeProfile())
	rg.DELETE("/encoding/profiles/:profile_id", JWTMiddleware(h.jwtMgr), h.deleteEncodeProfile())

	// Transcoding & Encoding Status - independent path
	rg.GET("/encoding/tasks", h.getEncodingTasksFlat())
	rg.GET("/encoding/events", h.transcodingEvents())
	rg.POST("/encoding/retry", JWTMiddleware(h.jwtMgr), h.retryTaskByID())
	rg.POST("/encoding/retry-all-failed", JWTMiddleware(h.jwtMgr), h.retryAllFailed())

	// ================================
	// 2. Public Routes (short_token based) - MediaCMS style
	// 与 MediaCMS 的 /api/v1/media/{friendly_token} 一致
	// ================================
	publicMedias := rg.Group("/medias")
	{
		// 核心：使用 short_token，不接受 ID
		publicMedias.GET("/:short_token", h.getPublicDetail())

		// 交互操作（支持可选JWT）
		publicMedias.POST("/:short_token/likes", JWTMiddleware(h.jwtMgr), h.toggleLikeByShortToken())
		publicMedias.GET("/:short_token/likes", OptionalJWTMiddleware(h.jwtMgr), h.getLikeStatusByShortToken())
		publicMedias.DELETE("/:short_token/likes", JWTMiddleware(h.jwtMgr), h.toggleDislikeByShortToken())
		publicMedias.POST("/:short_token/favorites", JWTMiddleware(h.jwtMgr), h.toggleFavoriteByShortToken())
		publicMedias.GET("/:short_token/favorites", OptionalJWTMiddleware(h.jwtMgr), h.getFavoriteStatusByShortToken())
		publicMedias.GET("/:short_token/shares", h.getShareUrlByShortToken())
		publicMedias.POST("/:short_token/shares", JWTMiddleware(h.jwtMgr), h.recordShareByShortToken())

		publicMedias.GET("/:short_token/sprite.vtt", h.getSpriteVTT())
		publicMedias.GET("/:short_token/sprite.jpg", h.getSpriteImage())
	}

	// ================================
	// 3. Admin Routes (ID based, requires JWT + Admin role)
	// 与 MediaCMS 的 /api/v1/manage_* 一致
	// ================================
	adminMedias := rg.Group("/admin/medias").Use(JWTMiddleware(h.jwtMgr), AdminMiddleware(h.jwtMgr))
	{
		// 管理列表（支持更多过滤条件）
		adminMedias.GET("", h.adminListMedia())

		// 核心管理操作：使用 ID，返回完整信息
		adminMedias.GET("/:id", h.adminGetByID())
		adminMedias.PUT("/:id", h.adminUpdateMedia())
		adminMedias.DELETE("/:id", h.adminDeleteMedia())

		// 统计和变体信息
		adminMedias.GET("/:id/stats", h.adminGetStats())
		adminMedias.GET("/:id/variants", h.adminGetVariants())
		adminMedias.GET("/:id/tasks", h.listEncodingTasks())
		adminMedias.POST("/:id/tasks/:taskId/retry", h.retryTranscode())

		// 状态变更
		adminMedias.PUT("/:id/state", h.adminChangeState())

		// 审核操作
		adminMedias.PUT("/:id/review", h.reviewMedia())
		adminMedias.POST("/review/batch", h.batchReviewMedia())
		adminMedias.GET("/:id/review-logs", h.listReviewLogs())

		adminMedias.POST("/:id/regenerate-sprite", h.regenerateSprite())
		adminMedias.POST("/:id/regenerate-thumbnail", h.regenerateThumbnail())
	}

	// ================================
	// 4. Legacy Routes (backward compatibility)
	// 保持旧路由可用，使用 /media/:id (单数) 前缀避免冲突
	// TODO: Phase 6 删除这些路由
	// ================================
	legacyMedia := rg.Group("/media")
	{
		legacyMedia.GET("/:id", h.getMedia())
		legacyMedia.PUT("/:id", JWTMiddleware(h.jwtMgr), h.updateMedia())
		legacyMedia.DELETE("/:id", JWTMiddleware(h.jwtMgr), h.deleteMedia())
		legacyMedia.GET("/:id/variants", h.getMediaVariants())
		legacyMedia.GET("/:id/tasks", h.listEncodingTasks())
		legacyMedia.POST("/:id/tasks/:taskId/retry", JWTMiddleware(h.jwtMgr), h.retryTranscode())
		legacyMedia.GET("/:id/likes", h.getLikeStatus())
		legacyMedia.POST("/:id/likes", JWTMiddleware(h.jwtMgr), h.toggleLike())
		legacyMedia.DELETE("/:id/likes", JWTMiddleware(h.jwtMgr), h.toggleDislike())
		legacyMedia.GET("/:id/favorites", h.getFavoriteStatus())
		legacyMedia.POST("/:id/favorites", JWTMiddleware(h.jwtMgr), h.toggleFavorite())
		legacyMedia.DELETE("/:id/favorites", JWTMiddleware(h.jwtMgr), h.toggleFavorite())
		legacyMedia.GET("/:id/shares", h.getShareUrl())
		legacyMedia.POST("/:id/shares", JWTMiddleware(h.jwtMgr), h.recordShare())
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

	resultItems := make([]map[string]interface{}, len(items))
	for i, m := range items {
		item := filterPublicFields(m)
		edges := map[string]interface{}{}

		if m.UserId != "" && m.UserId != "0" && h.userUC != nil {
			if u, err := h.userUC.GetUser(ctx, m.UserId); err == nil {
				userMap := map[string]interface{}{
					"id":       u.Id,
					"username": u.Username,
					"nickname": u.Nickname,
					"avatar":   u.Avatar,
				}
				edges["user"] = []map[string]interface{}{userMap}
			}
		}

		if m.CategoryId != "" {
			catMap := map[string]interface{}{
				"id": m.CategoryId,
			}
			edges["category"] = []map[string]interface{}{catMap}
		}

		if len(edges) > 0 {
			item["edges"] = edges
		}

		resultItems[i] = item
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     resultItems,
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

	// For video files, just keep as "video" type
	// Short video detection (vertical aspect ratio) will be added later if needed

	m := &biz.Media{
		Title:          title,
		Description:    description,
		Type:           mediaType,
		Url:            fileURL,
		UserId:         claims.GetUserID(),
		State:          "active",
		EncodingStatus: "pending",
		MimeType:       mimeType,
		Md5Sum:         fileMD5,
		Size:           written,
		Extension:      strings.TrimPrefix(ext, "."),
		Privacy:        int32(privacy),
		Tags:           tags,
		CategoryId:     strconv.Itoa(categoryID),
		ReviewStatus:   "pending_review",
		Listable:       false,
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

	if m.UserId != claims.GetUserID() && !claims.IsStaff {
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

	if m.UserId != claims.GetUserID() && !claims.IsStaff {
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
				"progress": `+fmt.Sprintf("%d", ev.Progress)+`,
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

	var profile dto.EncodeProfile
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
	var profile dto.EncodeProfile
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

	stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), id, "like")
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

	stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), id, "dislike")
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
		userID = val.(*auth.Claims).GetUserID()
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

	stats, err := h.likeFavoriteUC.ToggleFavorite(ctx, claims.GetUserID(), id)
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
		userID = val.(*auth.Claims).GetUserID()
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

func (h *MediaHandler) reviewMediaHandler(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := c.Context()

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	if !claims.IsStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "only staff can review media"})
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Action     string `json:"action"`
		ReviewerID string `json:"reviewer_id"`
		Comment    string `json:"comment"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approve := req.Action == "approve"

	updated, err := h.uc.ReviewMedia(ctx, idStr, approve, req.Comment, claims.GetUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            updated.Id,
		"review_status": updated.ReviewStatus,
		"listable":      updated.Listable,
		"updated_at":    updated.UpdateTime,
	})
}

// --- List Media ---

func ptrBool(v bool) *bool       { return &v }
func ptrString(v string) *string { return &v }

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

		opt.Listable = ptrBool(true)

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

		resultItems := make([]map[string]interface{}, len(items))
		for i, m := range items {
			item := filterPublicFields(m)
			edges := map[string]interface{}{}

			if m.UserId != "" && m.UserId != "0" && h.userUC != nil {
				if u, err := h.userUC.GetUser(ctx, m.UserId); err == nil {
					userMap := map[string]interface{}{
						"id":       u.Id,
						"username": u.Username,
						"nickname": u.Nickname,
						"avatar":   u.Avatar,
					}
					edges["user"] = []map[string]interface{}{userMap}
				}
			}

			if m.CategoryId != "" {
				catMap := map[string]interface{}{
					"id": m.CategoryId,
				}
				edges["category"] = []map[string]interface{}{catMap}
			}

			if len(edges) > 0 {
				item["edges"] = edges
			}

			resultItems[i] = item
		}

		OK(c, gin.H{
			"items":     resultItems,
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
		claims, ok := GetClaims(c)
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
			UserId:         claims.GetUserID(),
			State:          "active",
			EncodingStatus: "pending",
			MimeType:       mimeType,
			Md5Sum:         fileMD5,
			Size:           written,
			Extension:      strings.TrimPrefix(ext, "."),
			Privacy:        int32(privacy),
			Tags:           tags,
			CategoryId:     strconv.Itoa(categoryID),
			ReviewStatus:   "pending_review",
			Listable:       false,
		}

		if m.UserId == "" || m.UserId == "0" {
			Fail(c, ErrInternal, fmt.Sprintf("invalid user_id: %q", m.UserId))
			return
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
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	CategoryID     *int     `json:"category_id"`
	Tags           []string `json:"tags"`
	Privacy        *int     `json:"privacy"`
	State          *string  `json:"state"`
	Featured       *bool    `json:"featured"`
	EnableComments *bool    `json:"enable_comments"`
	AllowDownload  *bool    `json:"allow_download"`
}

func (h *MediaHandler) updateMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := GetClaims(c)
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

		if m.UserId != claims.GetUserID() && !claims.IsStaff {
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
		if req.EnableComments != nil {
			m.EnableComments = *req.EnableComments
		}
		if req.AllowDownload != nil {
			m.AllowDownload = *req.AllowDownload
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

		claims, ok := GetClaims(c)
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

		if m.UserId != claims.GetUserID() && !claims.IsStaff {
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
		var profile dto.EncodeProfile
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
		var profile dto.EncodeProfile
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
		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), id, "like")
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
		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), id, "dislike")
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
			userID = claims.(*auth.Claims).GetUserID()
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
		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleFavorite(ctx, claims.GetUserID(), id)
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
			userID = claims.(*auth.Claims).GetUserID()
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

// ================================
// Public API Handlers (short_token based)
// MediaCMS style: /api/v1/medias/{short_token}
// ================================

// getPublicDetail 公开媒体详情 (MediaCMS: views.MediaDetail)
func (h *MediaHandler) getPublicDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		media, err := h.uc.GetByShortToken(ctx, shortToken)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		if media.State == "private" || media.Privacy == 2 {
			Fail(c, ErrForbidden, "private media")
			return
		}

		if !media.Listable || media.ReviewStatus != "reviewed" {
			Fail(c, ErrForbidden, "media not available")
			return
		}

		go func() {
			bgCtx := context.Background()
			h.uc.IncrementViewCount(bgCtx, media.Id)
		}()

		result := filterPublicFields(media)

		edges := map[string]interface{}{}

		if media.UserId != "" && media.UserId != "0" && h.userUC != nil {
			if u, err := h.userUC.GetUser(ctx, media.UserId); err == nil {
				userMap := map[string]interface{}{
					"id":       u.Id,
					"username": u.Username,
					"nickname": u.Nickname,
					"avatar":   u.Avatar,
				}
				if h.playlistChannelUC != nil {
					if subs, _, subErr := h.playlistChannelUC.ListUserChannels(ctx, u.Id, 1, 1); subErr == nil {
						userMap["subscriber_count"] = len(subs)
					}
				}
				edges["user"] = []map[string]interface{}{userMap}
			}
		}

		if media.ChannelId != "" && h.playlistChannelUC != nil {
			if ch, err := h.playlistChannelUC.GetChannel(ctx, media.ChannelId); err == nil {
				edges["channels"] = []map[string]interface{}{
					{
						"id":   ch.ID,
						"name": ch.Title,
					},
				}
			}
		}

		if len(edges) > 0 {
			result["edges"] = edges
		}

		OK(c, result)
	}
}

// toggleLikeByShortToken 通过 short_token 点赞
func (h *MediaHandler) toggleLikeByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		// 解析 short_token → ID
		mediaID, err := h.uc.ResolveToID(ctx, shortToken)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), mediaID, "like")
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

// getLikeStatusByShortToken 获取点赞状态（通过 short_token）
func (h *MediaHandler) getLikeStatusByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		// 解析 short_token → ID
		mediaID, err := h.uc.ResolveToID(ctx, shortToken)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		var userID string
		if claims, ok := GetClaims(c); ok {
			userID = claims.GetUserID()
		}

		stats, err := h.likeFavoriteUC.GetMediaStats(ctx, userID, mediaID)
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

// toggleDislikeByShortToken 通过 short_token 点踩
func (h *MediaHandler) toggleDislikeByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		// 解析 short_token → ID
		mediaID, err := h.uc.ResolveToID(ctx, shortToken)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), mediaID, "dislike")
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

// toggleFavoriteByShortToken 通过 short_token 收藏
func (h *MediaHandler) toggleFavoriteByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		// 解析 short_token → ID
		mediaID, err := h.uc.ResolveToID(ctx, shortToken)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		stats, err := h.likeFavoriteUC.ToggleFavorite(ctx, claims.GetUserID(), mediaID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"is_favorited":  stats.IsFavorited,
			"favorite_count": stats.FavoriteCount,
		})
	}
}

// getFavoriteStatusByShortToken 获取收藏状态（通过 short_token）
func (h *MediaHandler) getFavoriteStatusByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		// 解析 short_token → ID
		mediaID, err := h.uc.ResolveToID(ctx, shortToken)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		var userID string
		if claims, ok := c.Get("claims"); ok {
			userID = claims.(*auth.Claims).GetUserID()
		}

		stats, err := h.likeFavoriteUC.GetMediaStats(ctx, userID, mediaID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"is_favorited":  stats.IsFavorited,
			"favorite_count": stats.FavoriteCount,
		})
	}
}

// getShareUrlByShortToken 获取分享链接（通过 short_token）
func (h *MediaHandler) getShareUrlByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		shortToken := c.Param("short_token")

		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		// Build share URL - using short_token instead of id
		shareUrl := c.Request.Host + "/watch?v=" + shortToken
		// Add https:// if not present
		if len(shareUrl) > 0 && shareUrl[0] != 'h' {
			shareUrl = "https://" + shareUrl
		}

		OK(c, gin.H{
			"url": shareUrl,
		})
	}
}

// recordShareByShortToken 记录分享（通过 short_token）
func (h *MediaHandler) recordShareByShortToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("claims")
		if !exists {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		shortToken := c.Param("short_token")
		if shortToken == "" {
			Fail(c, ErrBadRequest, "Invalid short_token")
			return
		}

		// TODO: Implement share count increment in the future
		// For now, just return success

		OK(c, gin.H{
			"success": true,
		})
	}
}

// filterPublicFields 过滤公开字段，隐藏敏感信息
func filterPublicFields(m *biz.Media) map[string]interface{} {
	result := map[string]interface{}{
		"id":              m.Id,
		"short_token":     m.ShortToken,
		"title":           m.Title,
		"description":     m.Description,
		"url":             m.Url,
		"hls_file":        m.HlsFile,
		"thumbnail":       m.Thumbnail,
		"poster":          m.Poster,
		"preview_file":    m.PreviewFilePath,
		"duration":        m.Duration,
		"width":           m.Width,
		"height":          m.Height,
		"view_count":      m.ViewCount,
		"like_count":      m.LikeCount,
		"dislike_count":   m.DislikeCount,
		"comment_count":   m.CommentCount,
		"favorite_count":  m.FavoriteCount,
		"user_id":         m.UserId,
		"category_id":     m.CategoryId,
		"channel_id":      m.ChannelId,
		"tags":            m.Tags,
		"type":            m.Type,
		"encoding_status": m.EncodingStatus,
		"state":           m.State,
		"allow_download":  m.AllowDownload,
		"enable_comments": m.EnableComments,
		"featured":        m.Featured,
	}

	if m.CreateTime != nil {
		result["created_at"] = m.CreateTime.AsTime().Format(time.RFC3339)
	} else {
		result["created_at"] = time.Now().Format(time.RFC3339)
	}
	if m.UpdateTime != nil {
		result["updated_at"] = m.UpdateTime.AsTime().Format(time.RFC3339)
	}
	if m.PublishedAt != nil {
		result["published_at"] = m.PublishedAt.AsTime().Format(time.RFC3339)
	}

	return result
}

// ================================
// Admin API Handlers (ID based)
// MediaCMS style: /api/v1/admin/medias/:id
// ================================

// adminListMedia 管理列表（支持更多过滤条件）
func (h *MediaHandler) adminListMedia() gin.HandlerFunc {
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

		if reviewStatus := c.Query("review_status"); reviewStatus != "" {
			opt.ReviewStatus = ptrString(reviewStatus)
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

		resultItems := make([]map[string]interface{}, len(items))
		for i, m := range items {
			item := filterPublicFields(m)
			edges := map[string]interface{}{}

			if m.UserId != "" && m.UserId != "0" && h.userUC != nil {
				if u, err := h.userUC.GetUser(ctx, m.UserId); err == nil {
					userMap := map[string]interface{}{
						"id":       u.Id,
						"username": u.Username,
						"nickname": u.Nickname,
						"avatar":   u.Avatar,
					}
					edges["user"] = []map[string]interface{}{userMap}
				}
			}

			if m.CategoryId != "" {
				catMap := map[string]interface{}{
					"id": m.CategoryId,
				}
				edges["category"] = []map[string]interface{}{catMap}
			}

			if len(edges) > 0 {
				item["edges"] = edges
			}

			resultItems[i] = item
		}

		OK(c, gin.H{
			"items":     resultItems,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// adminGetByID Admin 获取媒体完整信息（使用 GetByID）
func (h *MediaHandler) adminGetByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// ✅ 只使用 GetByID，不接受 short_token
		m, err := h.uc.GetByID(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		// Admin 可以看到完整信息，包括私有视频
		OK(c, m)
	}
}

// adminUpdateMedia Admin 更新媒体
func (h *MediaHandler) adminUpdateMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// JWT + Admin 权限已由中间件验证
		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// 使用 GetByID 获取完整信息
		m, err := h.uc.GetByID(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
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
		if req.EnableComments != nil {
			m.EnableComments = *req.EnableComments
		}
		if req.AllowDownload != nil {
			m.AllowDownload = *req.AllowDownload
		}

		updated, err := h.uc.UpdateMedia(ctx, m)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		// Re-fetch to get full details
		updated, _ = h.uc.GetByID(ctx, idStr)

		OK(c, updated)
	}
}

// adminDeleteMedia Admin 删除媒体
func (h *MediaHandler) adminDeleteMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// JWT + Admin 权限已由中间件验证
		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// 使用 GetByID 验证存在性
		m, err := h.uc.GetByID(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		// Admin 可以删除任何媒体（不需要权限检查）

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

// adminGetStats 获取统计数据
func (h *MediaHandler) adminGetStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		// 使用 GetByID 获取媒体信息
		m, err := h.uc.GetByID(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		OK(c, gin.H{
			"id":             m.Id,
			"view_count":     m.ViewCount,
			"like_count":     m.LikeCount,
			"dislike_count":  m.DislikeCount,
			"comment_count":  m.CommentCount,
			"favorite_count": m.FavoriteCount,
			"encoding_status": m.EncodingStatus,
		})
	}
}

// adminGetVariants 获取变体信息（Admin 版本）
func (h *MediaHandler) adminGetVariants() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "invalid media ID")
			return
		}

		summary, err := h.uc.GetMediaVariantsByUUID(ctx, idStr)
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

// adminChangeState 变更媒体状态
func (h *MediaHandler) adminChangeState() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// JWT + Admin 权限已由中间件验证
		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var req struct {
			State   string `json:"state"`
			Comment string `json:"comment"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if req.State == "" {
			Fail(c, ErrBadRequest, "state is required")
			return
		}

		// 使用 GetByID 验证存在性
		_, err := h.uc.GetByID(ctx, idStr)
		if err != nil {
			Fail(c, ErrMediaNotFound, "media not found")
			return
		}

		err = h.uc.UpdateMediaState(ctx, idStr, req.State)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		// Re-fetch to get updated state
		updated, _ := h.uc.GetByID(ctx, idStr)

		// 获取操作者信息（用于审计日志）
		var changedBy string
		if cl, ok := c.Get("claims"); ok {
			changedBy = cl.(*auth.Claims).GetUserID()
		} else {
			changedBy = "system"
		}

		OK(c, gin.H{
			"id":         updated.Id,
			"state":      updated.State,
			"updated_at": updated.UpdateTime,
			"changed_by": changedBy,
		})
	}
}

// reviewMedia handles PUT /admin/medias/:id/review
func (h *MediaHandler) reviewMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		if !claims.IsStaff {
			Fail(c, ErrForbidden, "only staff can review media")
			return
		}

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var req struct {
			Action  string `json:"action"`
			Comment string `json:"comment"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if req.Action != "approve" && req.Action != "reject" {
			Fail(c, ErrBadRequest, "action must be 'approve' or 'reject'")
			return
		}

		approve := req.Action == "approve"

		updated, err := h.uc.ReviewMedia(ctx, idStr, approve, req.Comment, claims.GetUserID())
		if err != nil {
			if strings.Contains(err.Error(), "invalid state transition") {
				Fail(c, ErrBadRequest, err.Error())
			} else {
				Fail(c, ErrInternal, err.Error())
			}
			return
		}

		OK(c, gin.H{
			"id":            updated.Id,
			"review_status": updated.ReviewStatus,
			"listable":      updated.Listable,
			"updated_at":    updated.UpdateTime,
		})
	}
}

// batchReviewMediaRequest is the JSON body for POST /admin/medias/review/batch
type batchReviewMediaRequest struct {
	MediaIDs []string `json:"media_ids"`
	Action   string   `json:"action"`
	Comment  string   `json:"comment"`
}

// batchReviewMedia handles POST /admin/medias/review/batch
func (h *MediaHandler) batchReviewMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}

		if !claims.IsStaff {
			Fail(c, ErrForbidden, "only staff can review media")
			return
		}

		var req batchReviewMediaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if len(req.MediaIDs) == 0 {
			Fail(c, ErrBadRequest, "media_ids is required")
			return
		}

		if req.Action != "approve" && req.Action != "reject" {
			Fail(c, ErrBadRequest, "action must be 'approve' or 'reject'")
			return
		}

		approve := req.Action == "approve"

		var succeeded []string
		var failed []gin.H

		for _, mediaID := range req.MediaIDs {
			_, err := h.uc.ReviewMedia(ctx, mediaID, approve, req.Comment, claims.GetUserID())
			if err != nil {
				failed = append(failed, gin.H{
					"media_id": mediaID,
					"error":    err.Error(),
				})
			} else {
				succeeded = append(succeeded, mediaID)
			}
		}

		OK(c, gin.H{
			"succeeded": succeeded,
			"failed":    failed,
		})
	}
}

// listReviewLogs handles GET /admin/medias/:id/review-logs
func (h *MediaHandler) listReviewLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		idStr := c.Param("id")
		if idStr == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		logs, err := h.uc.ListReviewLogs(ctx, idStr)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"items": logs,
		})
	}
}

func (h *MediaHandler) regenerateSprite() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		entMedia, err := h.entityClient.Media.Get(ctx, id)
		if err != nil {
			Fail(c, ErrNotFound, "media not found")
			return
		}

		if entMedia.Type != "video" {
			Fail(c, ErrBadRequest, "cannot regenerate sprite for non-video media")
			return
		}

		if entMedia.SpriteStatus == "processing" {
			Fail(c, ErrBadRequest, "sprite generation already in progress")
			return
		}

		go func() {
			_ = h.uc.RegenerateSprite(context.Background(), id)
		}()

		OK(c, gin.H{
			"media_id":      id,
			"sprite_status": "pending",
			"message":       "sprite regeneration scheduled",
		})
	}
}

func (h *MediaHandler) regenerateThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "Invalid ID")
			return
		}

		var req struct {
			Timestamp float64 `json:"timestamp"`
		}
		_ = c.ShouldBindJSON(&req)

		if err := h.uc.RegenerateThumbnail(ctx, id, req.Timestamp); err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		entMedia, _ := h.entityClient.Media.Get(ctx, id)
		thumb := ""
		var thumbTime float64
		if entMedia != nil {
			thumb = entMedia.Thumbnail
			thumbTime = entMedia.ThumbnailTime
		}

		OK(c, gin.H{
			"media_id":       id,
			"thumbnail":      thumb,
			"thumbnail_time": thumbTime,
			"message":        "thumbnail regenerated",
		})
	}
}

func (h *MediaHandler) getSpriteVTT() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")
		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		entMedia, err := h.entityClient.Media.Query().Where(media.ShortToken(shortToken)).Only(ctx)
		if err != nil || entMedia.SpriteStatus != "success" {
			Fail(c, ErrNotFound, "sprite not available for this media")
			return
		}

		vttPath := filepath.Join(UploadDir, entMedia.VttPath)
		data, err := os.ReadFile(vttPath)
		if err != nil {
			Fail(c, ErrNotFound, "sprite not available for this media")
			return
		}

		c.Header("Content-Type", "text/vtt")
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(200, "text/vtt", data)
	}
}

func (h *MediaHandler) getSpriteImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		shortToken := c.Param("short_token")
		if shortToken == "" {
			Fail(c, ErrBadRequest, "short_token is required")
			return
		}

		entMedia, err := h.entityClient.Media.Query().Where(media.ShortToken(shortToken)).Only(ctx)
		if err != nil || entMedia.SpriteStatus != "success" {
			Fail(c, ErrNotFound, "sprite not available for this media")
			return
		}

		spritePath := filepath.Join(UploadDir, entMedia.SpritePath)
		data, err := os.ReadFile(spritePath)
		if err != nil {
			Fail(c, ErrNotFound, "sprite not available for this media")
			return
		}

		c.Header("Content-Type", "image/jpeg")
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(200, "image/jpeg", data)
	}
}
