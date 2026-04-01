/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides HTTP handlers for chunked multipart upload.
// Supports: init session → upload parts → complete/abort, with resume capability.
package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	entityupload "origadmin/application/origcms/internal/data/entity/uploadsession"
)

const (
	// ChunkSize is the default chunk size for multipart uploads (2 MB).
	ChunkSize = 2 * 1024 * 1024
	// MaxFileSize is the maximum allowed file size (5 GB).
	MaxFileSize = 5 << 30
	// SessionTTL is how long upload sessions stay valid (24 hours).
	SessionTTL = 24 * time.Hour
)

// --- Request/Response types ---

type initiateUploadRequest struct {
	Filename    string   `json:"filename"`
	FileSize    int64    `json:"file_size"`
	ContentType string   `json:"content_type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CategoryID  *int64   `json:"category_id"`
	Tags        []string `json:"tags"`
}

type initiateUploadResponse struct {
	UploadID   string `json:"upload_id"`
	TotalParts int32  `json:"total_parts"`
	ChunkSize  int32  `json:"chunk_size"`
}

type partInfo struct {
	PartNumber int32  `json:"part_number"`
	Etag       string `json:"etag"`
	Size       int64  `json:"size"`
}

type uploadPartResponse struct {
	Etag string `json:"etag"`
	Size int64  `json:"size"`
}

type listPartsResponse struct {
	Parts        []partInfo `json:"parts"`
	TotalParts   int32      `json:"total_parts"`
	UploadedSize int64      `json:"uploaded_size"`
	TotalSize    int64      `json:"total_size"`
	Status       string     `json:"status"`
}

type completeUploadRequest struct {
	UploadID string     `json:"upload_id"`
	Parts    []partInfo `json:"parts"`
	Sha256   string     `json:"sha256"`
}

type getUploadSessionResponse struct {
	UploadID     string     `json:"upload_id"`
	Filename     string     `json:"filename"`
	FileSize     int64      `json:"file_size"`
	ContentType  string     `json:"content_type"`
	TotalParts   int32      `json:"total_parts"`
	ChunkSize    int32      `json:"chunk_size"`
	UploadedSize int64      `json:"uploaded_size"`
	Status       string     `json:"status"`
	Parts        []partInfo `json:"parts"`
	CreatedAt    string     `json:"created_at"`
	ExpiresAt    string     `json:"expires_at"`
}

type listUploadSessionsResponse struct {
	Sessions []getUploadSessionResponse `json:"sessions"`
	Total    int32                      `json:"total"`
}

// --- Route registration ---

// RegisterUploadRoutes registers chunked upload routes under /api/v1/uploads.
func RegisterUploadRoutes(group *gin.RouterGroup, client *entity.Client, jwtMgr *auth.Manager) {
	uploads := group.Group("/uploads")
	{
		// Initiate a new multipart upload session (requires JWT)
		uploads.POST("/multipart", JWTMiddleware(jwtMgr), initiateMultipartUpload(client))

		// Upload a single part (requires JWT)
		uploads.POST("/:upload_id/parts/:part_number", JWTMiddleware(jwtMgr), uploadPart(client))

		// List uploaded parts (requires JWT, for resume)
		uploads.GET("/:upload_id/parts", JWTMiddleware(jwtMgr), listParts(client))

		// Complete multipart upload (requires JWT)
		uploads.POST("/:upload_id/complete", JWTMiddleware(jwtMgr), completeMultipartUpload(client))

		// Abort multipart upload (requires JWT)
		uploads.DELETE("/:upload_id", JWTMiddleware(jwtMgr), abortMultipartUpload(client))

		// Get upload session info (requires JWT)
		uploads.GET("/:upload_id", JWTMiddleware(jwtMgr), getUploadSession(client))

		// List user's upload sessions (requires JWT)
		uploads.GET("", JWTMiddleware(jwtMgr), listUploadSessions(client))
	}
}

// --- Handlers ---

// initiateMultipartUpload starts a new multipart upload session.
func initiateMultipartUpload(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := c.MustGet("claims").(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req initiateUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		// Validate
		if req.Filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
			return
		}
		if req.FileSize <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_size must be positive"})
			return
		}
		if req.FileSize > MaxFileSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("file too large (max %d bytes)", MaxFileSize)})
			return
		}

		// Calculate total parts
		totalParts := int(math.Ceil(float64(req.FileSize) / float64(ChunkSize)))

		// Generate upload ID
		uploadID := uuid.New().String()

		// Create temp directory for parts
		tempDir := filepath.Join(UploadDir, ".chunks", uploadID)
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			slog.Error("failed to create temp dir", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload session"})
			return
		}

		// Create upload session in DB
		create := client.UploadSession.Create().
			SetUploadID(uploadID).
			SetFilename(req.Filename).
			SetFileSize(req.FileSize).
			SetContentType(req.ContentType).
			SetTotalParts(totalParts).
			SetChunkSize(ChunkSize).
			SetUploadedSize(0).
			SetUserID(claims.UserID).
			SetStatus("pending").
			SetParts(map[int]string{}).
			SetTempDir(tempDir).
			SetExpiresAt(time.Now().Add(SessionTTL))

		if req.Title != "" {
			create.SetTitle(req.Title)
		}
		if req.Description != "" {
			create.SetDescription(req.Description)
		}
		if req.CategoryID != nil && *req.CategoryID > 0 {
			create.SetCategoryID(*req.CategoryID)
		}
		if len(req.Tags) > 0 {
			create.SetTags(req.Tags)
		}

		if _, err := create.Save(ctx); err != nil {
			slog.Error("failed to create upload session", "err", err)
			_ = os.RemoveAll(tempDir)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload session"})
			return
		}

		c.JSON(http.StatusOK, initiateUploadResponse{
			UploadID:   uploadID,
			TotalParts: int32(totalParts),
			ChunkSize:  int32(ChunkSize),
		})
	}
}

// uploadPart uploads a single part of a multipart upload.
func uploadPart(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, _ := c.MustGet("claims").(*auth.Claims)

		uploadID := c.Param("upload_id")
		partNumberStr := c.Param("part_number")
		partNumber, err := strconv.Atoi(partNumberStr)
		if err != nil || partNumber < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid part_number"})
			return
		}

		// Get upload session
		session, err := getUploadSessionByUploadID(ctx, client, uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}

		// Verify ownership
		if session.UserID != nil && *session.UserID != claims.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your upload session"})
			return
		}

		// Check session status
		if session.Status == "completed" || session.Status == "aborted" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("upload session is %s", session.Status)})
			return
		}

		// Check expiration
		if time.Now().After(session.ExpiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "upload session expired"})
			return
		}

		// Check part number
		if partNumber > session.TotalParts {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("part_number %d exceeds total_parts %d", partNumber, session.TotalParts)})
			return
		}

		// Limit body size to chunk size + 1 KB overhead
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(ChunkSize)+1024)

		// Read part data
		partPath := filepath.Join(session.TempDir, fmt.Sprintf("part_%d", partNumber))
		out, err := os.Create(partPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write part"})
			return
		}

		// Compute ETag while writing
		hasher := sha256.New()
		written, err := io.Copy(io.MultiWriter(out, hasher), c.Request.Body)
		out.Close()
		if err != nil {
			_ = os.Remove(partPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write part"})
			return
		}

		etag := hex.EncodeToString(hasher.Sum(nil))

		// Update session: add part to parts map, update uploaded size
		parts := session.Parts
		if parts == nil {
			parts = make(map[int]string)
		}
		parts[partNumber] = etag

		// Update uploaded size
		uploadedSize := session.UploadedSize + written

		_, err = client.UploadSession.UpdateOneID(session.ID).
			SetParts(parts).
			SetUploadedSize(uploadedSize).
			SetStatus("uploading").
			Save(ctx)
		if err != nil {
			_ = os.Remove(partPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session"})
			return
		}

		c.JSON(http.StatusOK, uploadPartResponse{
			Etag: etag,
			Size: written,
		})
	}
}

// listParts returns the list of already-uploaded parts (for resume).
func listParts(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, _ := c.MustGet("claims").(*auth.Claims)
		uploadID := c.Param("upload_id")

		session, err := getUploadSessionByUploadID(ctx, client, uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}

		if session.UserID != nil && *session.UserID != claims.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your upload session"})
			return
		}

		// Convert parts map to sorted slice
		parts := make([]partInfo, 0, len(session.Parts))
		for pn, etag := range session.Parts {
			partPath := filepath.Join(session.TempDir, fmt.Sprintf("part_%d", pn))
			var size int64
			if info, err := os.Stat(partPath); err == nil {
				size = info.Size()
			}
			parts = append(parts, partInfo{
				PartNumber: int32(pn),
				Etag:       etag,
				Size:       size,
			})
		}
		sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })

		c.JSON(http.StatusOK, listPartsResponse{
			Parts:        parts,
			TotalParts:   int32(session.TotalParts),
			UploadedSize: session.UploadedSize,
			TotalSize:    session.FileSize,
			Status:       session.Status,
		})
	}
}

// completeMultipartUpload merges all parts and creates the final media record.
func completeMultipartUpload(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, _ := c.MustGet("claims").(*auth.Claims)
		uploadID := c.Param("upload_id")

		var req completeUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			req.UploadID = uploadID
		}

		session, err := getUploadSessionByUploadID(ctx, client, uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}

		if session.UserID != nil && *session.UserID != claims.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your upload session"})
			return
		}

		if session.Status == "completed" {
			c.JSON(http.StatusOK, gin.H{"message": "already completed"})
			return
		}

		// Verify all parts are present
		if len(session.Parts) != session.TotalParts {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("not all parts uploaded (%d/%d)", len(session.Parts), session.TotalParts),
			})
			return
		}

		// Merge parts into final file
		ext := filepath.Ext(session.Filename)
		finalFilename := uuid.New().String() + ext
		finalPath := filepath.Join(UploadDir, finalFilename)
		finalFile, err := os.Create(finalPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create final file"})
			return
		}

		// Merge all parts in order
		var totalWritten int64
		for pn := 1; pn <= session.TotalParts; pn++ {
			partPath := filepath.Join(session.TempDir, fmt.Sprintf("part_%d", pn))
			partFile, err := os.Open(partPath)
			if err != nil {
				finalFile.Close()
				_ = os.Remove(finalPath)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("part %d not found", pn)})
				return
			}

			written, err := io.Copy(finalFile, partFile)
			partFile.Close()
			if err != nil {
				finalFile.Close()
				_ = os.Remove(finalPath)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to merge part %d", pn)})
				return
			}
			totalWritten += written
		}
		finalFile.Close()

		// Build the public URL
		fileURL := "/uploads/" + finalFilename
		title := session.Title
		if title == "" {
			title = strings.TrimSuffix(session.Filename, ext)
		}

		// Find the user to satisfy the Required edge
		u, err := client.User.Get(ctx, int(claims.UserID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
			return
		}

		// Create media record - MUST use SetUser to satisfy the Required edge
		create := client.Media.Create().
			SetTitle(title).
			SetDescription(session.Description).
			SetType("video").
			SetURL(fileURL).
			SetUser(u). // Set the edge directly using the object
			SetState("active").
			SetEncodingStatus("pending").
			SetMimeType(session.ContentType).
			SetSize(strconv.FormatInt(session.FileSize, 10)).
			SetExtension(strings.TrimPrefix(ext, ".")).
			SetPrivacy(1).
			SetTags(session.Tags)

		if session.CategoryID != nil && *session.CategoryID > 0 {
			create.SetCategoryID(int(*session.CategoryID))
		}

		media, err := create.Save(ctx)
		if err != nil {
			_ = os.Remove(finalPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create media record: " + err.Error()})
			return
		}

		// Update session status
		_, _ = client.UploadSession.UpdateOneID(session.ID).
			SetStatus("completed").
			SetStoragePath(finalPath).
			SetSha256(req.Sha256).
			Save(ctx)

		// Clean up temp directory
		go func() {
			_ = os.RemoveAll(session.TempDir)
		}()

		c.JSON(http.StatusOK, gin.H{"media": media})
	}
}

// abortMultipartUpload cancels an in-progress upload and cleans up.
func abortMultipartUpload(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, _ := c.MustGet("claims").(*auth.Claims)
		uploadID := c.Param("upload_id")

		session, err := getUploadSessionByUploadID(ctx, client, uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}

		if session.UserID != nil && *session.UserID != claims.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your upload session"})
			return
		}

		// Update status
		_, _ = client.UploadSession.UpdateOneID(session.ID).
			SetStatus("aborted").
			Save(ctx)

		// Clean up temp directory
		go func() {
			if session.TempDir != "" {
				_ = os.RemoveAll(session.TempDir)
			}
		}()

		c.JSON(http.StatusOK, gin.H{"message": "upload aborted"})
	}
}

// getUploadSession returns info about a single upload session.
func getUploadSession(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, _ := c.MustGet("claims").(*auth.Claims)
		uploadID := c.Param("upload_id")

		session, err := getUploadSessionByUploadID(ctx, client, uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}

		if session.UserID != nil && *session.UserID != claims.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your upload session"})
			return
		}

		c.JSON(http.StatusOK, sessionToResponse(session))
	}
}

// listUploadSessions returns all upload sessions for the current user.
func listUploadSessions(client *entity.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, _ := c.MustGet("claims").(*auth.Claims)

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		status := c.Query("status")

		query := client.UploadSession.Query().
			Where(entityupload.UserIDEQ(claims.UserID))

		if status != "" {
			query = query.Where(entityupload.StatusEQ(status))
		}

		total, err := query.Count(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		offset := (page - 1) * pageSize
		sessions, err := query.
			Order(entity.Desc("created_at")).
			Limit(pageSize).
			Offset(offset).
			All(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := make([]getUploadSessionResponse, len(sessions))
		for i, s := range sessions {
			result[i] = sessionToResponse(s)
		}

		c.JSON(http.StatusOK, listUploadSessionsResponse{
			Sessions: result,
			Total:    int32(total),
		})
	}
}

// --- Helpers ---

// getUploadSessionByUploadID finds an upload session by its upload_id string.
func getUploadSessionByUploadID(ctx context.Context, client *entity.Client, uploadID string) (*entity.UploadSession, error) {
	return client.UploadSession.Query().
		Where(entityupload.UploadIDEQ(uploadID)).
		Only(ctx)
}

// sessionToResponse converts an entity.UploadSession to the API response.
func sessionToResponse(s *entity.UploadSession) getUploadSessionResponse {
	parts := make([]partInfo, 0, len(s.Parts))
	for pn, etag := range s.Parts {
		parts = append(parts, partInfo{
			PartNumber: int32(pn),
			Etag:       etag,
		})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })

	return getUploadSessionResponse{
		UploadID:     s.UploadID,
		Filename:     s.Filename,
		FileSize:     s.FileSize,
		ContentType:  s.ContentType,
		TotalParts:   int32(s.TotalParts),
		ChunkSize:    int32(s.ChunkSize),
		UploadedSize: s.UploadedSize,
		Status:       s.Status,
		Parts:        parts,
		CreatedAt:    s.CreatedAt.Format(time.RFC3339),
		ExpiresAt:    s.ExpiresAt.Format(time.RFC3339),
	}
}

// init is called before main to ensure the predicate package is imported correctly.
var _ = entityupload.StatusIn
