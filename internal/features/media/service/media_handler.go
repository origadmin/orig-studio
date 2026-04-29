/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/infra/auth"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dto"
	userbiz "origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
)

// MediaHandler handles media-related HTTP routes.
type MediaHandler struct {
	jwtMgr            *auth.Manager
	mediaUC           *biz.MediaUseCase
	uploadUC          *biz.UploadUseCase
	likeFavoriteUC    *contentbiz.LikeFavoriteUseCase
	playlistChannelUC *contentbiz.PlaylistChannelUseCase
	userUC            *userbiz.UserUseCase
	permChecker       authbiz.PermissionChecker
	mediaService      *MediaService
}

// NewMediaHandler creates a new MediaHandler.
func NewMediaHandler(
	jwtMgr *auth.Manager,
	mediaUC *biz.MediaUseCase,
	uploadUC *biz.UploadUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistChannelUC *contentbiz.PlaylistChannelUseCase,
	userUC *userbiz.UserUseCase,
	permChecker authbiz.PermissionChecker,
) *MediaHandler {
	return &MediaHandler{
		jwtMgr:            jwtMgr,
		mediaUC:           mediaUC,
		uploadUC:          uploadUC,
		likeFavoriteUC:    likeFavoriteUC,
		playlistChannelUC: playlistChannelUC,
		userUC:            userUC,
		permChecker:       permChecker,
	}
}

// RegisterRoutes registers the handler's routes.
func (h *MediaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	medias := r.Group("/medias")
	{
		// Public routes
		medias.GET("", h.listMedias)
		medias.GET("/featured", h.listFeaturedMedias)
		medias.GET("/latest", h.listLatestMedias)

		// Transcoding & encoding routes
		medias.GET("/transcoding/status", h.transcodingStatus)
		medias.GET("/encoding/tasks", h.encodingTasks)
		medias.POST("/encoding/retry", server.WithJWT(h.jwtMgr, h.retryTask))
		medias.POST("/encoding/retry-all-failed", server.WithJWT(h.jwtMgr, h.retryAllFailed))

		// SSE for transcoding progress
		medias.GET("/transcoding/events", h.sseHandler)

		// Parameter routes
		medias.GET("/:id", h.getMedia)
		medias.GET("/:id/variants", h.mediaVariants)
		medias.POST("/:id/increment-view", h.incrementViewCount)

		// Like/favorite routes
		medias.POST("/:id/like", server.WithJWT(h.jwtMgr, h.likeMedia))
		medias.DELETE("/:id/like", server.WithJWT(h.jwtMgr, h.unlikeMedia))
		medias.POST("/:id/favorite", server.WithJWT(h.jwtMgr, h.favoriteMedia))
		medias.DELETE("/:id/favorite", server.WithJWT(h.jwtMgr, h.unfavoriteMedia))
	}
}

// listMedias handles GET /medias
func (h *MediaHandler) listMedias(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  c.Query("keyword"),
		},
	}

	items, total, err := h.mediaUC.ListMedias(r.Context(), opts)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OKPage(gc, items, int64(total), page, pageSize)
}

// listFeaturedMedias handles GET /medias/featured
func (h *MediaHandler) listFeaturedMedias(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     1,
			PageSize: int32(limit),
		},
		Featured: boolPtr(true),
	}

	items, _, err := h.mediaUC.ListMedias(r.Context(), opts)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, items)
}

// listLatestMedias handles GET /medias/latest
func (h *MediaHandler) listLatestMedias(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     1,
			PageSize: int32(limit),
		},
	}

	items, _, err := h.mediaUC.ListMedias(r.Context(), opts)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, items)
}

// getMedia handles GET /medias/:id
func (h *MediaHandler) getMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	id := c.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	item, err := h.mediaUC.GetMedia(r.Context(), id)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
		return
	}

	server.OK(gc, item)
}

// incrementViewCount handles POST /medias/:id/increment-view
func (h *MediaHandler) incrementViewCount(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	id := c.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	count, err := h.mediaUC.IncrementViewCount(r.Context(), id)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"view_count": count})
}

// transcodingStatus handles GET /medias/transcoding/status
func (h *MediaHandler) transcodingStatus(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.TranscodingStatusHTTPHandler(w, r)
		return
	}
	status, err := h.mediaUC.GetTranscodingStatus(r.Context(), nil)
	if err != nil {
		server.Fail(handler.NewGinContextAdapterFromHTTP(w, r).GinContext(), server.ErrInternal, err.Error())
		return
	}
	server.OK(handler.NewGinContextAdapterFromHTTP(w, r).GinContext(), gin.H{
		"processing_count": status.ProcessingCount,
		"pending_count":    status.PendingCount,
		"failed_count":     status.FailedCount,
		"success_count":    status.SuccessCount,
	})
}

// encodingTasks handles GET /medias/encoding/tasks
func (h *MediaHandler) encodingTasks(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.EncodingTasksHTTPHandler(w, r)
		return
	}
	server.OK(handler.NewGinContextAdapterFromHTTP(w, r).GinContext(), gin.H{"tasks": []interface{}{}})
}

// retryTask handles POST /medias/encoding/retry
func (h *MediaHandler) retryTask(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.RetryTaskHTTPHandler(w, r)
		return
	}
	server.OK(handler.NewGinContextAdapterFromHTTP(w, r).GinContext(), gin.H{"success": true})
}

// retryAllFailed handles POST /medias/encoding/retry-all-failed
func (h *MediaHandler) retryAllFailed(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.RetryAllFailedHTTPHandler(w, r)
		return
	}
	server.OK(handler.NewGinContextAdapterFromHTTP(w, r).GinContext(), gin.H{"success": true})
}

// sseHandler handles GET /medias/transcoding/events
func (h *MediaHandler) sseHandler(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.SSEHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// mediaVariants handles GET /medias/:id/variants
func (h *MediaHandler) mediaVariants(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.MediaVariantsHTTPHandler(w, r)
		return
	}
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()
	id := c.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}
	server.OK(gc, gin.H{"variants": []interface{}{}})
}

// likeMedia handles POST /medias/:id/like
func (h *MediaHandler) likeMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	mediaID := c.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.ToggleLike(r.Context(), claims.GetUserID(), mediaID, "like")
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, stats)
		return
	}

	server.OK(gc, gin.H{"success": true})
}

// unlikeMedia handles DELETE /medias/:id/like
func (h *MediaHandler) unlikeMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	mediaID := c.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.ToggleLike(r.Context(), claims.GetUserID(), mediaID, "unlike")
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, stats)
		return
	}

	server.OK(gc, gin.H{"success": true})
}

// favoriteMedia handles POST /medias/:id/favorite
func (h *MediaHandler) favoriteMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	mediaID := c.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.ToggleFavorite(r.Context(), claims.GetUserID(), mediaID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, stats)
		return
	}

	server.OK(gc, gin.H{"success": true})
}

// unfavoriteMedia handles DELETE /medias/:id/favorite
func (h *MediaHandler) unfavoriteMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	mediaID := c.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.ToggleFavorite(r.Context(), claims.GetUserID(), mediaID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, stats)
		return
	}

	server.OK(gc, gin.H{"success": true})
}
