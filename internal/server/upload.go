/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-media/biz"
)

// --- Route registration ---

// UploadHandler handles file uploads.
type UploadHandler struct {
	uc     *biz.UploadUseCase
	jwtMgr *auth.Manager
	log    *log.Helper
}

func NewUploadHandler(uc *biz.UploadUseCase, jwtMgr *auth.Manager, logger log.Logger) *UploadHandler {
	return &UploadHandler{
		uc:     uc,
		jwtMgr: jwtMgr,
		log:    log.NewHelper(log.With(logger, "module", "server/upload")),
	}
}

func (h *UploadHandler) Register(r handler.Router) {
	// Create a direct gin registration since we need proper middleware chain
	// This is a workaround - the handler.Router interface doesn't support gin.HandlerFunc directly
}

// RegisterGin directly registers routes with gin.RouterGroup (to be called from routes.go)
func (h *UploadHandler) RegisterGin(rg *gin.RouterGroup) {
	uploads := rg.Group("/uploads")
	{
		uploads.POST("/multipart", JWTMiddleware(h.jwtMgr), h.initiateMultipartUpload())
		uploads.POST("/:uploadId/parts/:partNumber", JWTMiddleware(h.jwtMgr), h.uploadPart())
		uploads.POST("/:uploadId/complete", JWTMiddleware(h.jwtMgr), h.completeMultipartUpload())
		uploads.POST("/:uploadId/abort", JWTMiddleware(h.jwtMgr), h.abortMultipartUpload())
		uploads.GET("/:uploadId/parts", JWTMiddleware(h.jwtMgr), h.listParts())

		uploads.GET("/sessions", JWTMiddleware(h.jwtMgr), h.listUploadSessions())
		uploads.GET("/sessions/:uploadId", JWTMiddleware(h.jwtMgr), h.getUploadSession())
	}
}

// --- Handlers (Refactored to use biz.UploadUseCase) ---

// initiateMultipartUpload starts a new multipart upload session.
func (h *UploadHandler) initiateMultipartUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.log.Infof("initiateMultipartUpload called")
		
		claims, _ := c.MustGet("claims").(*auth.Claims)
		h.log.Infof("user_id: %s", claims.UserID)

		var req struct {
			Filename    string   `json:"filename"`
			FileSize    int64    `json:"file_size"`
			ContentType string   `json:"content_type"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			CategoryID  *string  `json:"category_id"`
			Tags        []string `json:"tags"`
			Thumbnail   string   `json:"thumbnail"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			h.log.Errorf("invalid request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}
		
		h.log.Infof("request: filename=%s, file_size=%d, content_type=%s", req.Filename, req.FileSize, req.ContentType)

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
			h.log.Errorf("InitiateMultipartUpload failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		h.log.Infof("session created: upload_id=%s, total_parts=%d", session.UploadID, session.TotalParts)

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
		partNumberStr := c.Param("partNumber")
		h.log.Infof("uploadPart called: upload_id=%s, part_number=%s", uploadID, partNumberStr)
		
		partNumber, err := strconv.Atoi(partNumberStr)
		if err != nil {
			h.log.Errorf("invalid part number: %s, error: %v", partNumberStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid part number"})
			return
		}

		data, err := c.GetRawData()
		if err != nil {
			h.log.Errorf("failed to read part data: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read part data"})
			return
		}
		h.log.Infof("read part data: size=%d bytes", len(data))

		etag, err := h.uc.UploadPart(c.Request.Context(), uploadID, partNumber, data)
		if err != nil {
			h.log.Errorf("UploadPart failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		h.log.Infof("part uploaded: upload_id=%s, part_number=%d, etag=%s", uploadID, partNumber, etag)
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
			CategoryID  *string  `json:"category_id"`
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
		status := enums.ParseUploadStatus(c.Query("status"))

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
