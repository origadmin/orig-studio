/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	pb "origadmin/application/origcms/api/gen/v1/media"
	types "origadmin/application/origcms/api/gen/v1/types"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dto"
	userbiz "origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemservice "origadmin/application/origcms/internal/features/system/service"
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
	settingUC         *systembiz.SettingUseCase
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
	mediaService *MediaService,
	settingUC *systembiz.SettingUseCase,
) *MediaHandler {
	return &MediaHandler{
		jwtMgr:            jwtMgr,
		mediaUC:           mediaUC,
		uploadUC:          uploadUC,
		likeFavoriteUC:    likeFavoriteUC,
		playlistChannelUC: playlistChannelUC,
		userUC:            userUC,
		permChecker:       permChecker,
		mediaService:      mediaService,
		settingUC:         settingUC,
	}
}

// RegisterRoutes registers the handler's routes.
func (h *MediaHandler) RegisterRoutes(r http2.Router) {
	mediasGroup := r.Group("/medias")
	mediasGroup.Use(systemservice.ModuleGuardCtx(h.settingUC, "module_videos"))

	medias := mediasGroup.Group("")
	{
		// Public routes
		medias.GET("", server.HTTPToHandlerFunc(h.listMedias))
		medias.GET("/featured", server.HTTPToHandlerFunc(h.listFeaturedMedias))
		medias.GET("/latest", server.HTTPToHandlerFunc(h.listLatestMedias))

		// Transcoding & encoding routes
		medias.GET("/transcoding/status", server.HTTPToHandlerFunc(h.transcodingStatus))
		medias.GET("/encoding/tasks", server.HTTPToHandlerFunc(h.encodingTasks))
		medias.POST("/encoding/retry", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.retryTask)))
		medias.POST("/encoding/retry-all-failed", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.retryAllFailed)))

		// NOTE: SSE endpoint moved to admin route group (/admin/medias/transcoding/events)
		// See admin_handler.go encoding group

		// Parameter routes
		medias.GET("/:id", server.HTTPToHandlerFunc(h.getMedia))
		medias.GET("/:id/variants", server.HTTPToHandlerFunc(h.mediaVariants))
		medias.POST("/:id/view", server.HTTPToHandlerFunc(h.incrementViewCount))

		// Like/favorite routes (singular - proto canonical)
		medias.POST("/:id/like", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.likeMedia)))
		medias.DELETE("/:id/like", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unlikeMedia)))
		medias.POST("/:id/favorite", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.favoriteMedia)))
		medias.DELETE("/:id/favorite", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unfavoriteMedia)))

		// Like/favorite routes (plural - frontend compatibility)
		medias.GET("/:id/likes", server.WithOptionalJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.getMediaLikes)))
		medias.POST("/:id/likes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.likeMedia)))
		medias.DELETE("/:id/likes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unlikeMedia)))
		medias.GET("/:id/favorites", server.WithOptionalJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.getMediaFavorites)))
		medias.POST("/:id/favorites", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.favoriteMedia)))
		medias.DELETE("/:id/favorites", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unfavoriteMedia)))
	}
}

// listMedias handles GET /medias
func (h *MediaHandler) listMedias(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	page, _ := strconv.Atoi(gc.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(gc.Query("page_size"))
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
			Keyword:  gc.Query("keyword"),
		},
	}

	if categoryIDStr := gc.Query("category_id"); categoryIDStr != "" {
		if cid, err := strconv.ParseInt(categoryIDStr, 10, 64); err == nil && cid > 0 {
			opts.CategoryID = &cid
		}
	}
	if categoryIDsStr := gc.Query("category_ids"); categoryIDsStr != "" {
		for _, idStr := range strings.Split(categoryIDsStr, ",") {
			if cid, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && cid > 0 {
				opts.CategoryIDs = append(opts.CategoryIDs, cid)
			}
		}
	}
	if state := gc.Query("state"); state != "" {
		opts.State = state
	}

	items, total, err := h.mediaUC.ListMedias(r.Context(), opts)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = (total + int32(pageSize) - 1) / int32(pageSize)
	}
	server.OK(gc, &pb.ListMediasResponse{
		Total:      total,
		Items:      items,
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: totalPages,
	})
}

// listFeaturedMedias handles GET /medias/featured
func (h *MediaHandler) listFeaturedMedias(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	limit, _ := strconv.Atoi(gc.Query("limit"))
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

	server.OK(gc, &pb.ListMediasResponse{
		Items:    items,
		Page:     1,
		PageSize: int32(limit),
	})
}

// listLatestMedias handles GET /medias/latest
func (h *MediaHandler) listLatestMedias(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	limit, _ := strconv.Atoi(gc.Query("limit"))
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

	server.OK(gc, &pb.ListMediasResponse{
		Items:    items,
		Page:     1,
		PageSize: int32(limit),
	})
}

// getMedia handles GET /medias/:id
func (h *MediaHandler) getMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	id := gc.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	item, err := h.mediaUC.GetMedia(r.Context(), id)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
		return
	}

	// Public API: hide private media (return 404 as if it doesn't exist)
	if item.Privacy == types.Privacy_PRIVACY_PRIVATE {
		server.Fail(gc, server.ErrNotFound, "media not found")
		return
	}

	server.OK(gc, &pb.GetMediaResponse{
		Media: item,
	})
}

// incrementViewCount handles POST /medias/:id/view
func (h *MediaHandler) incrementViewCount(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	id := gc.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	count, err := h.mediaUC.IncrementViewCount(r.Context(), id)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.IncrementViewCountResponse{
		ViewCount: count,
	})
}

// transcodingStatus handles GET /medias/transcoding/status
func (h *MediaHandler) transcodingStatus(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.TranscodingStatusHTTPHandler(w, r)
		return
	}
	gc := ginadapter.GetGinContext(r)
	status, err := h.mediaUC.GetTranscodingStatus(r.Context(), nil)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}
	server.OK(gc, &pb.GetEncodingStatusResponse{
		ProcessingCount: int32(status.ProcessingCount),
		PendingCount:    int32(status.PendingCount),
		FailedCount:     int32(status.FailedCount),
		SuccessCount:    int32(status.SuccessCount),
	})
}

// encodingTasks handles GET /medias/encoding/tasks
func (h *MediaHandler) encodingTasks(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.EncodingTasksHTTPHandler(w, r)
		return
	}
	gc := ginadapter.GetGinContext(r)
	server.OK(gc, &pb.ListEncodingTasksResponse{
		Tasks: []*types.EncodingTask{},
	})
}

// retryTask handles POST /medias/encoding/retry
func (h *MediaHandler) retryTask(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.RetryTaskHTTPHandler(w, r)
		return
	}
	gc := ginadapter.GetGinContext(r)
	server.OK(gc, &pb.RetryEncodingTaskResponse{})
}

// retryAllFailed handles POST /medias/encoding/retry-all-failed
func (h *MediaHandler) retryAllFailed(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.RetryAllFailedHTTPHandler(w, r)
		return
	}
	gc := ginadapter.GetGinContext(r)
	server.OK(gc, &pb.RetryAllFailedTasksResponse{})
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
	gc := ginadapter.GetGinContext(r)
	id := gc.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}
	server.OK(gc, &pb.GetMediaVariantsResponse{
		Variants: []*types.MediaVariant{},
	})
}

// likeMedia handles POST /medias/:id/like
func (h *MediaHandler) likeMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	mediaID := gc.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
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
		server.OK(gc, &pb.ToggleMediaLikeResponse{
			IsLiked:      stats.UserLikeType == "like",
			IsDisliked:   stats.UserLikeType == "dislike",
			LikeCount:    stats.LikeCount,
			DislikeCount: stats.DislikeCount,
		})
		return
	}

	server.OK(gc, &pb.ToggleMediaLikeResponse{})
}

// unlikeMedia handles DELETE /medias/:id/like
func (h *MediaHandler) unlikeMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	mediaID := gc.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
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
		server.OK(gc, &pb.ToggleMediaLikeResponse{
			IsLiked:      stats.UserLikeType == "like",
			IsDisliked:   stats.UserLikeType == "dislike",
			LikeCount:    stats.LikeCount,
			DislikeCount: stats.DislikeCount,
		})
		return
	}

	server.OK(gc, &pb.DeleteMediaLikeResponse{})
}

// favoriteMedia handles POST /medias/:id/favorite
func (h *MediaHandler) favoriteMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	mediaID := gc.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
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
		server.OK(gc, &pb.ToggleMediaFavoriteResponse{
			IsFavorited:   stats.IsFavorited,
			FavoriteCount: stats.FavoriteCount,
		})
		return
	}

	server.OK(gc, &pb.ToggleMediaFavoriteResponse{})
}

// unfavoriteMedia handles DELETE /medias/:id/favorite
func (h *MediaHandler) unfavoriteMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	mediaID := gc.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
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
		server.OK(gc, &pb.ToggleMediaFavoriteResponse{
			IsFavorited:   stats.IsFavorited,
			FavoriteCount: stats.FavoriteCount,
		})
		return
	}

	server.OK(gc, &pb.DeleteMediaFavoriteResponse{})
}

// getMediaLikes handles GET /medias/:id/likes (plural path for frontend compatibility)
func (h *MediaHandler) getMediaLikes(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	mediaID := gc.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	// Resolve short_token to internal ID if needed
	resolvedID, err := h.mediaUC.ResolveToID(r.Context(), mediaID)
	if err != nil {
		resolvedID = mediaID
	}

	val, exists := gc.Get("claims")
	userID := ""
	if exists && val != nil {
		claims := val.(*auth.Claims)
		userID = claims.GetUserID()
	}

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.GetMediaStats(r.Context(), userID, resolvedID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, gin.H{
			"is_liked":      stats.UserLikeType == "like",
			"is_disliked":   stats.UserLikeType == "dislike",
			"like_count":    stats.LikeCount,
			"dislike_count": stats.DislikeCount,
		})
		return
	}

	server.OK(gc, gin.H{
		"is_liked":      false,
		"is_disliked":   false,
		"like_count":    0,
		"dislike_count": 0,
	})
}

// getMediaFavorites handles GET /medias/:id/favorites (plural path for frontend compatibility)
func (h *MediaHandler) getMediaFavorites(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	mediaID := gc.Param("id")
	if mediaID == "" {
		server.Fail(gc, server.ErrBadRequest, "media id is required")
		return
	}

	// Resolve short_token to internal ID if needed
	resolvedID, err := h.mediaUC.ResolveToID(r.Context(), mediaID)
	if err != nil {
		resolvedID = mediaID
	}

	val, exists := gc.Get("claims")
	userID := ""
	if exists && val != nil {
		claims := val.(*auth.Claims)
		userID = claims.GetUserID()
	}

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.GetMediaStats(r.Context(), userID, resolvedID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, gin.H{
			"is_favorited":   stats.IsFavorited,
			"favorite_count": stats.FavoriteCount,
		})
		return
	}

	server.OK(gc, gin.H{
		"is_favorited":   false,
		"favorite_count": 0,
	})
}
