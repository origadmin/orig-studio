/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
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
	"origadmin/application/origcms/internal/data/entity"
	entitycategory "origadmin/application/origcms/internal/data/entity/category"
	entitymedia "origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/svc-media/biz"
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
	client   *entity.Client
	jwtMgr   *auth.Manager
	uc       *biz.MediaUseCase
	uploadUC *biz.UploadUseCase
}

func NewMediaHandler(client *entity.Client, jwtMgr *auth.Manager, uc *biz.MediaUseCase, uploadUC *biz.UploadUseCase) *MediaHandler {
	return &MediaHandler{client: client, jwtMgr: jwtMgr, uc: uc, uploadUC: uploadUC}
}

func (h *MediaHandler) Register(group *gin.RouterGroup) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0o755); err != nil {
		slog.Warn("failed to create upload directory", "err", err)
	}

	media := group.Group("/media")
	{
		// 1. Static/Fixed routes MUST come before parameter routes
		media.GET("/transcoding/status", h.getTranscodingStatus())
		media.GET("/transcoding/events", h.transcodingEvents())

		profiles := media.Group("/profiles")
		{
			profiles.GET("", h.listEncodeProfiles())
			profiles.GET("/:profile_id", h.getEncodeProfile())
			profiles.POST("", JWTMiddleware(h.jwtMgr), h.createEncodeProfile())
			profiles.PUT("/:profile_id", JWTMiddleware(h.jwtMgr), h.updateEncodeProfile())
			profiles.DELETE("/:profile_id", JWTMiddleware(h.jwtMgr), h.deleteEncodeProfile())
		}

		media.POST("/upload", JWTMiddleware(h.jwtMgr), h.uploadMedia())

		// 2. Collection routes
		media.GET("", h.listMedia())

		// 3. Parameter routes (/:id)
		media.GET("/:id", h.getMedia())
		media.PUT("/:id", JWTMiddleware(h.jwtMgr), h.updateMedia())
		media.DELETE("/:id", JWTMiddleware(h.jwtMgr), h.deleteMedia())
		media.GET("/:id/tasks", h.listEncodingTasks())
	}
}

// --- List Media ---

func (h *MediaHandler) listMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		query := h.client.Media.Query()

		if state := c.Query("state"); state != "" {
			query = query.Where(entitymedia.StateEQ(state))
		} else {
			query = query.Where(entitymedia.StateEQ("active"))
		}

		if mediaType := c.Query("type"); mediaType != "" {
			query = query.Where(entitymedia.TypeEQ(mediaType))
		}

		if userIDStr := c.Query("user_id"); userIDStr != "" {
			if userID, err := strconv.Atoi(userIDStr); err == nil {
				query = query.Where(entitymedia.UserIDEQ(userID))
			}
		}

		if catIDStr := c.Query("category_id"); catIDStr != "" {
			if catID, err := strconv.Atoi(catIDStr); err == nil {
				query = query.Where(entitymedia.HasCategoryWith(entitycategory.IDEQ(catID)))
			}
		}

		if keyword := c.Query("keyword"); keyword != "" {
			query = query.Where(entitymedia.TitleContains(keyword))
		}

		if c.Query("featured") == "true" {
			query = query.Where(entitymedia.FeaturedEQ(true))
		}

		total, err := query.Count(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		orderBy := c.DefaultQuery("order_by", "created_at")
		desc := c.DefaultQuery("descending", "true") == "true"

		switch orderBy {
		case "title":
			if desc {
				query = query.Order(entity.Desc("title"))
			} else {
				query = query.Order(entity.Asc("title"))
			}
		case "view_count":
			if desc {
				query = query.Order(entity.Desc("view_count"))
			} else {
				query = query.Order(entity.Asc("view_count"))
			}
		case "created_at":
			fallthrough
		default:
			if desc {
				query = query.Order(entity.Desc("created_at"))
			} else {
				query = query.Order(entity.Asc("created_at"))
			}
		}

		offset := (page - 1) * pageSize
		items, err := query.
			Limit(pageSize).
			Offset(offset).
			WithUser().
			WithCategory().
			All(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"list":      items,
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		m, err := h.client.Media.Query().
			Where(entitymedia.ID(id)).
			WithUser().
			WithCategory().
			Only(ctx)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}

		go func() {
			bgCtx := context.Background()
			h.client.Media.UpdateOneID(id).AddViewCount(1).Exec(bgCtx)
		}()

		c.JSON(http.StatusOK, m)
	}
}

func (h *MediaHandler) uploadMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		client := h.client

		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSizeVideo)

		file, header, err := c.Request.FormFile("file")
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
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": fmt.Sprintf("File type %s is not allowed for %s", mimeType, mediaType),
			})
			return
		}

		maxSize := maxUploadSizeByType(mediaType)
		if header.Size > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("File too large for %s (max %d bytes)", mediaType, maxSize),
			})
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
		_ = os.MkdirAll(filepath.Dir(filePath), 0755)

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

		create := client.Media.Create().
			SetTitle(title).
			SetDescription(description).
			SetType(mediaType).
			SetURL(fileURL).
			SetUserID(int(claims.UserID)).
			SetState("active").
			SetEncodingStatus("pending").
			SetMimeType(mimeType).
			SetMd5sum(fileMD5).
			SetSize(strconv.FormatInt(written, 10)).
			SetExtension(strings.TrimPrefix(ext, ".")).
			SetPrivacy(privacy).
			SetTags(tags)

		if categoryID > 0 {
			create.SetCategoryID(categoryID)
		}

		m, err := create.Save(ctx)
		if err != nil {
			os.Remove(filePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media record: " + err.Error()})
			return
		}

		// Background media processing (Transcoding)
		if mediaType == "video" && h.uploadUC != nil {
			go h.uploadUC.ProcessMedia(context.Background(), int64(m.ID), fileURL, mimeType)
		}

		m, _ = client.Media.Query().
			Where(entitymedia.ID(m.ID)).
			WithUser().
			WithCategory().
			Only(ctx)

		c.JSON(http.StatusCreated, m)
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
		client := h.client

		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		m, err := client.Media.Get(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}

		if m.UserID != int(claims.UserID) && !claims.IsStaff {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own media"})
			return
		}

		var req updateMediaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := client.Media.UpdateOneID(id)

		if req.Title != "" {
			update.SetTitle(req.Title)
		}
		if req.Description != "" {
			update.SetDescription(req.Description)
		}
		if req.CategoryID != nil {
			if *req.CategoryID > 0 {
				update.SetCategoryID(*req.CategoryID)
			} else {
				update.ClearCategory()
			}
		}
		if req.Tags != nil {
			update.SetTags(req.Tags)
		}
		if req.Privacy != nil {
			update.SetPrivacy(*req.Privacy)
		}
		if req.State != nil {
			update.SetState(*req.State)
		}
		if req.Featured != nil {
			update.SetFeatured(*req.Featured)
		}

		updated, err := update.Save(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updated, _ = client.Media.Query().
			Where(entitymedia.ID(id)).
			WithUser().
			WithCategory().
			Only(ctx)

		c.JSON(http.StatusOK, updated)
	}
}

func (h *MediaHandler) deleteMedia() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		client := h.client

		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		m, err := client.Media.Get(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}

		if m.UserID != int(claims.UserID) && !claims.IsStaff {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own media"})
			return
		}

		if m.URL != "" {
			filename := filepath.Base(m.URL)
			_ = os.Remove(filepath.Join(UploadDir, "uploads", filename))
		}

		if err := client.Media.DeleteOneID(id).Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func (h *MediaHandler) listEncodingTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		tasks, err := h.uc.ListEncodingTasks(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, tasks)
	}
}

func (h *MediaHandler) getTranscodingStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := h.uc.GetTranscodingStatus(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profiles": profiles})
	}
}

func (h *MediaHandler) getEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("profile_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
			return
		}
		p, err := h.uc.GetEncodeProfile(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": p})
	}
}

func (h *MediaHandler) createEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var profile biz.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p, err := h.uc.CreateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"profile": p})
	}
}

func (h *MediaHandler) updateEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("profile_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
			return
		}
		var profile biz.EncodeProfile
		if err := c.ShouldBindJSON(&profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		profile.Id = id
		p, err := h.uc.UpdateEncodeProfile(c.Request.Context(), &profile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": p})
	}
}

func (h *MediaHandler) deleteEncodeProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("profile_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Profile ID"})
			return
		}
		if err := h.uc.DeleteEncodeProfile(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
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
