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

// RegisterUploadRoutes registers chunked upload routes using the UploadUseCase.
func RegisterUploadRoutes(group *gin.RouterGroup, uc *biz.UploadUseCase, jwtMgr *auth.Manager) {
	uploads := group.Group("/uploads")
	{
		// Require JWT for all upload operations
		uploads.Use(JWTMiddleware(jwtMgr))

		// Initiate a new multipart upload session
		uploads.POST("/multipart", initiateMultipartUpload(uc))

		// Upload a single part
		uploads.POST("/:upload_id/parts/:part_number", uploadPart(uc))

		// List uploaded parts (for resume)
		uploads.GET("/:upload_id/parts", listParts(uc))

		// Complete multipart upload
		uploads.POST("/:upload_id/complete", completeMultipartUpload(uc))

		// Abort multipart upload
		uploads.DELETE("/:upload_id", abortMultipartUpload(uc))

		// Get upload session info
		uploads.GET("/:upload_id", getUploadSession(uc))

		// List user's upload sessions
		uploads.GET("", listUploadSessions(uc))
	}
}

// --- Handlers (Refactored to use biz.UploadUseCase) ---

func initiateMultipartUpload(uc *biz.UploadUseCase) gin.HandlerFunc {
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
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		session, err := uc.InitiateMultipartUpload(
			c.Request.Context(),
			req.Filename,
			req.FileSize,
			req.ContentType,
			req.Title,
			req.Description,
			req.CategoryID,
			req.Tags,
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

func uploadPart(uc *biz.UploadUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("upload_id")
		partNumber, _ := strconv.Atoi(c.Param("part_number"))

		data, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read part data"})
			return
		}

		etag, err := uc.UploadPart(c.Request.Context(), uploadID, partNumber, data)
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

func listParts(uc *biz.UploadUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("upload_id")
		session, err := uc.GetSession(c.Request.Context(), uploadID)
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
			"uploaded_size": session.UploadedSize,
			"total_size":    session.FileSize,
			"status":        session.Status,
		})
	}
}

func completeMultipartUpload(uc *biz.UploadUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("upload_id")
		var req struct {
			Sha256 string `json:"sha256"`
		}
		_ = c.ShouldBindJSON(&req)

		media, err := uc.CompleteMultipartUpload(c.Request.Context(), uploadID, req.Sha256)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete upload: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"media": media})
	}
}

func abortMultipartUpload(uc *biz.UploadUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("upload_id")
		if err := uc.AbortMultipartUpload(c.Request.Context(), uploadID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "upload aborted"})
	}
}

func getUploadSession(uc *biz.UploadUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := c.Param("upload_id")
		session, err := uc.GetSession(c.Request.Context(), uploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

func listUploadSessions(uc *biz.UploadUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.MustGet("claims").(*auth.Claims)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		status := c.Query("status")

		sessions, total, err := uc.ListSessions(c.Request.Context(), claims.UserID, status, page, pageSize)
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
