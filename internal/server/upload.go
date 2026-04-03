/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-media/biz"
)

// --- Route registration ---

// UploadHandler handles file uploads.
type UploadHandler struct {
	uc     *biz.UploadUseCase
	jwtMgr *auth.Manager
}

func NewUploadHandler(uc *biz.UploadUseCase, jwtMgr *auth.Manager) *UploadHandler {
	return &UploadHandler{uc: uc, jwtMgr: jwtMgr}
}

func (h *UploadHandler) Register(group *gin.RouterGroup) {
	uploads := group.Group("/uploads")
	uploads.Use(JWTMiddleware(h.jwtMgr))
	{
		// Multipart upload routes
		uploads.POST("/multipart", h.initiateMultipartUpload())
		uploads.POST("/:uploadId/parts/:partNumber", h.uploadPart())
		uploads.POST("/:uploadId/complete", h.completeMultipartUpload())
		uploads.POST("/:uploadId/abort", h.abortMultipartUpload())
		uploads.GET("/:uploadId/parts", h.listParts())

		// Session management
		uploads.GET("/sessions", h.listUploadSessions())
		uploads.GET("/sessions/:uploadId", h.getUploadSession())
	}
}

// --- Handlers (Refactored to use biz.UploadUseCase) ---

// initiateMultipartUpload starts a new multipart upload session.
func (h *UploadHandler) initiateMultipartUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.MustGet("claims").(*auth.Claims)

		var req struct {
			Filename    string   `json:"filename"`
			FileSize    int64    `json:"file_size"`
			ContentType string   `json:"content_type"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			CategoryID  *int64   `json:"category_id"`
			Tags        []string `json:"tags"`
			Thumbnail   string   `json:"thumbnail"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		session, err := h.uc.InitiateMultipartUpload(
			c.Request.Context(),
			req.Filename,
			req.FileSize,
			req.ContentType,
			req.Title,
			req.Description,
			req.CategoryID,
			req.Tags,
			req.Thumbnail,
			&claims.UserID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"upload_id":   session.UploadID,
			"total_parts": session.TotalParts,
			"chunk_size":  session.ChunkSize,
		})
	}
}

// uploadPart uploads a single part of a multipart upload.
func (h *UploadHandler) uploadPart() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("uploadId")
		partNumber, _ := strconv.Atoi(c.Param("partNumber"))

		data, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read part data"})
			return
		}

		etag, err := h.uc.UploadPart(c.Request.Context(), uploadID, partNumber, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"etag": etag,
			"size": len(data),
		})
	}
}

// listParts lists all uploaded parts for a specific session.
func (h *UploadHandler) listParts() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("uploadId")
		session, err := h.uc.GetSession(c.Request.Context(), uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}

		// Map biz parts to response format
		type partInfo struct {
			PartNumber int32  `json:"part_number"`
			Etag       string `json:"etag"`
		}
		parts := make([]partInfo, 0, len(session.Parts))
		for pn, etag := range session.Parts {
			parts = append(parts, partInfo{PartNumber: int32(pn), Etag: etag})
		}

		c.JSON(http.StatusOK, gin.H{
			"parts":         parts,
			"total_parts":   session.TotalParts,
			"chunk_size":    session.ChunkSize,
			"uploaded_size": session.UploadedSize,
			"total_size":    session.FileSize,
			"status":        session.Status,
		})
	}
}

// completeMultipartUpload completes a multipart upload session.
func (h *UploadHandler) completeMultipartUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("uploadId")
		var req struct {
			Sha256      string   `json:"sha256"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			CategoryID  *int64   `json:"category_id"`
			Tags        []string `json:"tags"`
			Thumbnail   string   `json:"thumbnail"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			// Even if JSON binding fails or is empty, we try to complete with defaults
		}

		media, err := h.uc.CompleteMultipartUpload(
			c.Request.Context(),
			uploadID,
			req.Sha256,
			req.Title,
			req.Description,
			req.CategoryID,
			req.Tags,
			req.Thumbnail,
		)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "failed to complete upload: " + err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{"media": media})
	}
}

// abortMultipartUpload aborts a multipart upload session.
func (h *UploadHandler) abortMultipartUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("uploadId")
		if err := h.uc.AbortMultipartUpload(c.Request.Context(), uploadID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "upload aborted"})
	}
}

// getUploadSession returns details for a specific upload session.
func (h *UploadHandler) getUploadSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("uploadId")
		session, err := h.uc.GetSession(c.Request.Context(), uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

// listUploadSessions lists all active upload sessions.
func (h *UploadHandler) listUploadSessions() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.MustGet("claims").(*auth.Claims)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		status := c.Query("status")

		sessions, total, err := h.uc.ListSessions(
			c.Request.Context(),
			claims.UserID,
			status,
			page,
			pageSize,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sessions": sessions,
			"total":    total,
		})
	}
}
