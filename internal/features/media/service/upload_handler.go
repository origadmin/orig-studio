/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/dal/enums"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/features/media/biz"
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

// RegisterRoutes registers the handler's routes with gin.RouterGroup
func (h *UploadHandler) RegisterRoutes(r http2.Router) {
	uploads := r.Group("/uploads")
	{
		// Simple upload (single file, multipart/form-data)
		uploads.POST("/simple", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.simpleUpload())))

		// Multipart upload (chunked, for large files)
		uploads.POST("/multipart", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.initiateMultipartUpload())))
		uploads.POST("/:uploadId/parts/:partNumber", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.uploadPart())))
		uploads.POST("/:uploadId/complete", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.completeMultipartUpload())))
		uploads.PATCH("/:uploadId/metadata", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.updateMetadata())))
		uploads.POST("/:uploadId/abort", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.abortMultipartUpload())))
		uploads.GET("/:uploadId/parts", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.listParts())))

		uploads.GET("/sessions", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.listUploadSessions())))
		uploads.GET("/sessions/:uploadId", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.getUploadSession())))
	}
}

// HTTPHandler returns an http.Handler that serves all upload routes.
// It creates a Gin engine, registers the routes under /api/v1, and returns
// the engine as a standard http.Handler. This is used to mount raw-binary
// upload endpoints on the Kratos HTTP server, bypassing the gateway's
// JSON-only codec.
func (h *UploadHandler) HTTPHandler() http.Handler {
	ginEngine := gin.New()
	ginEngine.Use(gin.Logger(), gin.Recovery())
	apiV1 := ginEngine.Group("/api/v1")
	router := ginadapter.NewRouterAdapter(apiV1)
	h.RegisterRoutes(router)
	return ginEngine
}

// --- Handlers (Refactored to use biz.UploadUseCase) ---

// simpleUpload handles a single-file upload via multipart/form-data.
// It initiates a session, uploads the file as one part, and completes
// the upload in a single request. Suitable for small files (<100MB).
func (h *UploadHandler) simpleUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		h.log.Infof("simpleUpload called")

		const simpleUploadMaxSize = 100 << 20
		r.Body = http.MaxBytesReader(w, r.Body, simpleUploadMaxSize)

		// Parse multipart form
		file, header, err := gc.Request.FormFile("file")
		if err != nil {
			if isRequestBodyTooLarge(err) {
				server.Fail(gc, server.ErrMediaTooLarge, "request body exceeds 100MB simple upload limit, please use multipart upload")
				return
			}
			h.log.Errorf("failed to read file from form: %v", err)
			server.Fail(gc, server.ErrBadRequest, "failed to read file: "+err.Error())
			return
		}
		defer file.Close()

		claims, ok := server.GetClaims(gc)
		if !ok || claims == nil {
			server.Fail(gc, server.ErrUnauthorized, "missing or expired token")
			return
		}
		userID := claims.GetUserID()

		title := gc.PostForm("title")
		if title == "" {
			title = header.Filename
		}
		description := gc.PostForm("description")
		thumbnail := gc.PostForm("thumbnail")

		var categoryID *int64
		if cidStr := gc.PostForm("category_id"); cidStr != "" {
			if cid, err := strconv.ParseInt(cidStr, 10, 64); err == nil {
				categoryID = &cid
			}
		}

		var channelID *string
		if chIDStr := gc.PostForm("channel_id"); chIDStr != "" {
			channelID = &chIDStr
		}

		var tags []string
		if tagsStr := gc.PostForm("tags"); tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
		}

		// Parse #hashtags from title and description, merge with explicit tags
		parsedHashtags := hashtag.ParseHashtags(title + " " + description)
		if len(parsedHashtags) > 0 {
			tags = mergeUploadTags(tags, parsedHashtags)
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		if header.Size > simpleUploadMaxSize {
			server.Fail(gc, server.ErrMediaTooLarge, fmt.Sprintf("file size %d exceeds simple upload limit %d, please use multipart upload", header.Size, simpleUploadMaxSize))
			return
		}

		// Read entire file into memory for simple upload (size limited)
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "failed to read file: "+err.Error())
			return
		}

		// Calculate SHA256 for integrity check
		hash := sha256.Sum256(fileBytes)
		fileSha256 := hex.EncodeToString(hash[:])

		// Initiate multipart upload session
		session, err := h.uc.InitiateMultipartUpload(
			r.Context(),
			header.Filename,
			header.Size,
			contentType,
			title,
			description,
			categoryID,
			channelID,
			tags,
			thumbnail,
			&userID,
		)
		if err != nil {
			h.log.Errorf("InitiateMultipartUpload failed: %v", err)
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		// Upload as single part
		etag, err := h.uc.UploadPart(r.Context(), session.UploadID, 1, bytes.NewReader(fileBytes), header.Size)
		if err != nil {
			h.log.Errorf("UploadPart failed: %v", err)
			_ = h.uc.AbortMultipartUpload(r.Context(), session.UploadID)
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		_ = etag

		// Complete the upload
		media, err := h.uc.CompleteMultipartUpload(
			r.Context(),
			session.UploadID,
			fileSha256,
			title,
			description,
			categoryID,
			channelID,
			tags,
			thumbnail,
		)
		if err != nil {
			h.log.Errorf("CompleteMultipartUpload failed: %v", err)
			server.Fail(gc, server.ErrInternal, "failed to complete upload: "+err.Error())
			return
		}

		h.log.Infof("simpleUpload completed: filename=%s, size=%d", header.Filename, header.Size)
		server.OK(gc, gin.H{
			"media_id":        media.Id,
			"state":           media.State,
			"encoding_status": media.EncodingStatus,
			"url":             media.Url,
			"thumbnail":       media.Thumbnail,
			"title":           media.Title,
			"type":            media.Type,
			"size":            media.Size,
			"mime_type":       media.MimeType,
		})
	}
}

// initiateMultipartUpload starts a new multipart upload session.
func (h *UploadHandler) initiateMultipartUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		h.log.Infof("initiateMultipartUpload called")

		claims, ok := server.GetClaims(gc)
		if !ok || claims == nil {
			server.Fail(gc, server.ErrUnauthorized, "missing or expired token")
			return
		}
		h.log.Infof("user_id: %s", claims.GetUserID())

		var req struct {
			Filename    string   `json:"filename"`
			FileSize    int64    `json:"file_size"`
			ContentType string   `json:"content_type"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			CategoryID  *int64   `json:"category_id"`
			ChannelID   *string  `json:"channel_id"`
			Tags        []string `json:"tags"`
			Thumbnail   string   `json:"thumbnail"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			h.log.Errorf("invalid request: %v", err)
			server.Fail(gc, server.ErrBadRequest, "invalid request: " + err.Error())
			return
		}
		
		h.log.Infof("request: filename=%s, file_size=%d, content_type=%s", req.Filename, req.FileSize, req.ContentType)

		userID := claims.GetUserID()
		session, err := h.uc.InitiateMultipartUpload(
			r.Context(),
			req.Filename,
			req.FileSize,
			req.ContentType,
			req.Title,
			req.Description,
			req.CategoryID,
			req.ChannelID,
			req.Tags,
			req.Thumbnail,
			&userID,
		)
		if err != nil {
			h.log.Errorf("InitiateMultipartUpload failed: %v", err)
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		
		h.log.Infof("session created: upload_id=%s, total_parts=%d", session.UploadID, session.TotalParts)

		server.OK(gc, gin.H{
			"upload_id":   session.UploadID,
			"total_parts": session.TotalParts,
			"chunk_size":  session.ChunkSize,
		})
	}
}

// uploadPart uploads a single part of a multipart upload.
func (h *UploadHandler) uploadPart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		uploadID := gc.Param("uploadId")
		partNumberStr := gc.Param("partNumber")
		h.log.Infof("uploadPart called: upload_id=%s, part_number=%s", uploadID, partNumberStr)

		partNumber, err := strconv.Atoi(partNumberStr)
		if err != nil {
			h.log.Errorf("invalid part number: %s, error: %v", partNumberStr, err)
			server.Fail(gc, server.ErrBadRequest, "invalid part number")
			return
		}

		// Read body into memory to make it seekable for S3 storage.
		// S3 PutObject requires a seekable stream (io.ReadSeeker) for content-length
		// computation and retry handling. HTTP request body (r.Body) is a
		// non-seekable io.ReadCloser, so we buffer it into a bytes.Reader.
		// Default chunk size is 5MB, which is acceptable to hold in memory.
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Errorf("failed to read part body: %v", err)
			server.Fail(gc, server.ErrBadRequest, "failed to read request body: "+err.Error())
			return
		}
		_ = r.Body.Close()

		size := int64(len(bodyBytes))
		h.log.Infof("uploading part: size=%d bytes", size)

		reader := bytes.NewReader(bodyBytes)
		etag, err := h.uc.UploadPart(r.Context(), uploadID, partNumber, reader, size)
		if err != nil {
			h.log.Errorf("UploadPart failed: %v", err)
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		h.log.Infof("part uploaded: upload_id=%s, part_number=%d, etag=%s", uploadID, partNumber, etag)
		server.OK(gc, gin.H{
			"etag": etag,
			"size": size,
		})
	}
}

// listParts lists all uploaded parts for a specific session.
func (h *UploadHandler) listParts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		uploadID := gc.Param("uploadId")
		session, err := h.uc.GetSession(r.Context(), uploadID)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "upload session not found")
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

		server.OK(gc, gin.H{
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
func (h *UploadHandler) completeMultipartUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		uploadID := gc.Param("uploadId")
		var req struct {
			Sha256      string   `json:"sha256"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			CategoryID  *int64   `json:"category_id"`
			ChannelID   *string  `json:"channel_id"`
			Tags        []string `json:"tags"`
			Thumbnail   string   `json:"thumbnail"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			// Even if JSON binding fails or is empty, we try to complete with defaults
		}

		// Parse #hashtags from title and description, merge with explicit tags
		if req.Title != "" || req.Description != "" {
			parsedHashtags := hashtag.ParseHashtags(req.Title + " " + req.Description)
			if len(parsedHashtags) > 0 {
				req.Tags = mergeUploadTags(req.Tags, parsedHashtags)
			}
		}

		media, err := h.uc.CompleteMultipartUpload(
			r.Context(),
			uploadID,
			req.Sha256,
			req.Title,
			req.Description,
			req.CategoryID,
			req.ChannelID,
			req.Tags,
			req.Thumbnail,
		)
		if err != nil {
			server.Fail(gc, server.ErrInternal, "failed to complete upload: " + err.Error())
			return
		}

		server.OK(gc, gin.H{
			"media_id":        media.Id,
			"state":           media.State,
			"encoding_status": media.EncodingStatus,
			"url":             media.Url,
			"thumbnail":       media.Thumbnail,
			"title":           media.Title,
			"type":            media.Type,
			"size":            media.Size,
			"mime_type":       media.MimeType,
		})
	}
}

// abortMultipartUpload aborts a multipart upload session.
func (h *UploadHandler) abortMultipartUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		uploadID := gc.Param("uploadId")
		if err := h.uc.AbortMultipartUpload(r.Context(), uploadID); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, gin.H{"message": "upload aborted"})
	}
}

// updateMetadata updates the metadata of an ongoing upload session.
func (h *UploadHandler) updateMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		uploadID := gc.Param("uploadId")
		var req struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			CategoryID  *int64   `json:"category_id"`
			ChannelID   *string  `json:"channel_id"`
			Tags        []string `json:"tags"`
			Thumbnail   string   `json:"thumbnail"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid request: "+err.Error())
			return
		}

		if err := h.uc.UpdateUploadMetadata(
			r.Context(),
			uploadID,
			req.Title,
			req.Description,
			req.CategoryID,
			req.ChannelID,
			req.Tags,
			req.Thumbnail,
		); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"message": "metadata updated"})
	}
}

// getUploadSession returns details for a specific upload session.
func (h *UploadHandler) getUploadSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		uploadID := gc.Param("uploadId")
		session, err := h.uc.GetSession(r.Context(), uploadID)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "upload session not found")
			return
		}
		server.OK(gc, session)
	}
}

// listUploadSessions lists all active upload sessions.
func (h *UploadHandler) listUploadSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		claims, ok := server.GetClaims(gc)
		if !ok || claims == nil {
			server.Fail(gc, server.ErrUnauthorized, "missing or expired token")
			return
		}
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		status := enums.ParseUploadStatus(gc.Query("status"))

		sessions, total, err := h.uc.ListSessions(
			r.Context(),
			claims.GetUserID(),
			status,
			page,
			pageSize,
		)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"sessions": sessions,
			"total":    total,
		})
	}
}

// isRequestBodyTooLarge checks if the error is from http.MaxBytesReader
// indicating the request body exceeds the configured size limit.
func isRequestBodyTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
