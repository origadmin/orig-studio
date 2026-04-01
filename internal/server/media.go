/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides HTTP handlers for media CRUD + upload.
// T2.1: Enhanced upload with type detection, size limits, MIME/MD5,
// JWT protection, and full field population.
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

// RegisterMediaRoutes registers media CRUD routes.
func RegisterMediaRoutes(group *gin.RouterGroup, client *entity.Client, jwtMgr *auth.Manager) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		slog.Warn("failed to create upload directory", "err", err)
	}

	media := group.Group("/media")
	{
		// List media (public, with filters)
		media.GET("", listMedia(client))

		// Get media by ID (public, increments view count)
		media.GET("/:id", getMedia(client))

		// Upload media file (requires JWT)
		media.POST("/upload", JWTMiddleware(jwtMgr), uploadMedia(client))

		// Update media (requires JWT + owner check)
		media.PUT("/:id", JWTMiddleware(jwtMgr), updateMedia(client))

		// Delete media (requires JWT + owner/admin check)
		media.DELETE("/:id", JWTMiddleware(jwtMgr), deleteMedia(client))
	}
}

// --- List Media ---

// listMedia returns a paginated, filterable list of media.
func listMedia(client *entity.Client) gin.HandlerFunc {
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

		query := client.Media.Query()

		// Only show active state by default
		if state := c.Query("state"); state != "" {
			query = query.Where(entitymedia.StateEQ(state))
		} else {
			query = query.Where(entitymedia.StateEQ("active"))
		}

		// Filter by type
		if mediaType := c.Query("type"); mediaType != "" {
			query = query.Where(entitymedia.TypeEQ(mediaType))
		}

		// Filter by user_id
		if userIDStr := c.Query("user_id"); userIDStr != "" {
			if userID, err := strconv.Atoi(userIDStr); err == nil {
				query = query.Where(entitymedia.UserIDEQ(userID))
			}
		}

		// Filter by category_id (via edge)
		if catIDStr := c.Query("category_id"); catIDStr != "" {
			if catID, err := strconv.Atoi(catIDStr); err == nil {
				query = query.Where(entitymedia.HasCategoryWith(entitycategory.IDEQ(catID)))
			}
		}

		// Filter by keyword (title search)
		if keyword := c.Query("keyword"); keyword != "" {
			query = query.Where(entitymedia.TitleContains(keyword))
		}

		// Filter: featured
		if c.Query("featured") == "true" {
			query = query.Where(entitymedia.FeaturedEQ(true))
		}

		// Count total (before pagination)
		total, err := query.Count(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Ordering
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

		// Pagination
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

// --- Get Media ---

// getMedia returns a single media by ID and increments view count.
func getMedia(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		m, err := client.Media.Query().
			Where(entitymedia.ID(id)).
			WithUser().
			WithCategory().
			Only(ctx)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}

		// Increment view count (fire-and-forget)
		go func() {
			bgCtx := context.Background()
			client.Media.UpdateOneID(id).AddViewCount(1).Exec(bgCtx)
		}()

		c.JSON(http.StatusOK, m)
	}
}

// --- Upload Media ---

// uploadMedia handles multipart file upload with validation.
func uploadMedia(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get user ID from JWT claims
		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Limit request body size (5 GB max)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSizeVideo)

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
			return
		}
		defer file.Close()

		// Detect MIME type (use http.DetectContentType for the first 512 bytes)
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mimeType := http.DetectContentType(buf[:n])
		// Seek back to start
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		// Also consider the client-provided Content-Type as a fallback
		clientMIME := header.Header.Get("Content-Type")
		if clientMIME != "" && clientMIME != "application/octet-stream" {
			// Trust client MIME if it's more specific (e.g. "video/mp4" vs "application/octet-stream")
			if mimeType == "application/octet-stream" {
				mimeType = clientMIME
			}
		}

		mediaType := detectMediaType(mimeType)

		// Validate MIME type
		if !isMIMEAllowed(mimeType, mediaType) {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": fmt.Sprintf("File type %s is not allowed for %s", mimeType, mediaType),
			})
			return
		}

		// Validate file size
		maxSize := maxUploadSizeByType(mediaType)
		if header.Size > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("File too large for %s (max %d bytes)", mediaType, maxSize),
			})
			return
		}

		// Compute MD5
		// Reset reader for MD5 computation
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		fileMD5, err := computeFileMD5(file)
		if err != nil {
			slog.Warn("failed to compute MD5", "err", err)
			fileMD5 = ""
		}

		// Reset reader for saving
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		// Generate unique filename preserving extension
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = mimeToExt(mimeType)
		}
		newFilename := uuid.New().String() + ext
		filePath := filepath.Join(UploadDir, newFilename)

		// Save file to disk
		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer out.Close()

		written, err := io.Copy(out, file)
		if err != nil {
			os.Remove(filePath) // cleanup
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
			return
		}

		// Build the public URL
		fileURL := "/uploads/" + newFilename

		// Get form fields
		title := c.PostForm("title")
		description := c.PostForm("description")
		categoryIDStr := c.PostForm("category_id")
		tagsStr := c.PostForm("tags")
		privacyStr := c.DefaultPostForm("privacy", "1")

		if title == "" {
			title = strings.TrimSuffix(header.Filename, ext)
		}

		// Parse optional fields
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

		// Create database record
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
			os.Remove(filePath) // cleanup
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media record: " + err.Error()})
			return
		}

		// Reload with user edge
		m, _ = client.Media.Query().
			Where(entitymedia.ID(m.ID)).
			WithUser().
			WithCategory().
			Only(ctx)

		c.JSON(http.StatusCreated, m)
	}
}

// --- Update Media ---

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

// updateMedia updates an existing media record.
func updateMedia(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get user from JWT
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

		// Fetch existing media
		m, err := client.Media.Get(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}

		// Owner or admin check
		if m.UserID != int(claims.UserID) && !claims.IsStaff {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own media"})
			return
		}

		// Parse request body
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

		// Reload with edges
		updated, _ = client.Media.Query().
			Where(entitymedia.ID(id)).
			WithUser().
			WithCategory().
			Only(ctx)

		c.JSON(http.StatusOK, updated)
	}
}

// --- Delete Media ---

// deleteMedia deletes a media record and its file.
func deleteMedia(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get user from JWT
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

		// Fetch existing media
		m, err := client.Media.Get(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}

		// Owner or admin check
		if m.UserID != int(claims.UserID) && !claims.IsStaff {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own media"})
			return
		}

		// Delete file from disk
		if m.URL != "" {
			filename := filepath.Base(m.URL)
			_ = os.Remove(filepath.Join(UploadDir, filename))
		}

		// Delete from database
		if err := client.Media.DeleteOneID(id).Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// --- Helpers ---

// mimeToExt returns a file extension for a given MIME type.
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

// Helper to sort query order direction
