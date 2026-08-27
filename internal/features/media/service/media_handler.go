/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/server"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
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

// NewMediaHandlerForMicroservice creates a MediaHandler for the media microservice.
// likeFavoriteUC is injected to support like/favorite endpoints routed to this service.
func NewMediaHandlerForMicroservice(
	jwtMgr *auth.Manager,
	mediaUC *biz.MediaUseCase,
	uploadUC *biz.UploadUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	mediaService *MediaService,
	settingUC *systembiz.SettingUseCase,
) *MediaHandler {
	return &MediaHandler{
		jwtMgr:            jwtMgr,
		mediaUC:           mediaUC,
		uploadUC:          uploadUC,
		likeFavoriteUC:    likeFavoriteUC,
		playlistChannelUC: nil,
		userUC:            nil,
		permChecker:       nil,
		mediaService:      mediaService,
		settingUC:         settingUC,
	}
}

// HTTPHandler returns an http.Handler that serves all media routes.
// It creates a Gin engine, registers the routes under /api/v1, and returns
// the engine as a standard http.Handler. Used to mount media endpoints on
// the Kratos HTTP server, bypassing the gateway's proto codec to return the
// flat task list format (items + counts) that the frontend expects.
func (h *MediaHandler) HTTPHandler() http.Handler {
	ginEngine := gin.New()
	ginEngine.Use(gin.Logger(), gin.Recovery())
	apiV1 := ginEngine.Group("/api/v1")
	router := ginadapter.NewRouterAdapter(apiV1)
	h.RegisterRoutes(router)
	return ginEngine
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

		// Transcoding & encoding routes (public status + admin-only SSE)
		medias.GET("/transcoding/status", server.HTTPToHandlerFunc(h.transcodingStatus))
		medias.GET("/encoding/tasks", server.HTTPToHandlerFunc(h.encodingTasks))
		medias.POST("/encoding/retry", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.retryTask)))
		medias.POST("/encoding/retry-all-failed", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.retryAllFailed)))

		// NOTE: SSE endpoint is admin-only, registered at /api/v1/admin/medias/transcoding/events

		// Parameter routes (portal uses short_token for public access)
		medias.GET("/:token", server.HTTPToHandlerFunc(h.getMedia))
		medias.GET("/:token/variants", server.HTTPToHandlerFunc(h.mediaVariants))
		medias.POST("/:token/view", server.HTTPToHandlerFunc(h.incrementViewCount))

		// Like/favorite routes (singular - proto canonical)
		medias.POST("/:token/like", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.likeMedia)))
		medias.DELETE("/:token/like", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unlikeMedia)))
		medias.POST("/:token/favorite", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.favoriteMedia)))
		medias.DELETE("/:token/favorite", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unfavoriteMedia)))

		// Like/favorite routes (plural - frontend compatibility)
		medias.GET("/:token/likes", server.WithOptionalJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.getMediaLikes)))
		medias.POST("/:token/likes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.likeMedia)))
		medias.DELETE("/:token/likes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unlikeMedia)))
		medias.GET("/:token/favorites", server.WithOptionalJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.getMediaFavorites)))
		medias.POST("/:token/favorites", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.favoriteMedia)))
		medias.DELETE("/:token/favorites", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.unfavoriteMedia)))
		// Dislike (no proto RPC exists for /dislikes; the gateway proxies this
		// path to media, so it is served here as a gin route -> real toggle).
		medias.POST("/:token/dislikes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.dislikeMedia)))
	}

	// NOTE: the admin review console endpoints (pending/history) are NOT
	// registered on this gin engine — the gin handler is mounted on the media
	// HTTP server only under the /api/v1/medias/ prefix (see
	// internal/enterprise/media/server/servers.go). They are registered there
	// as plain net/http handlers (ReviewHistoryList) so they
	// take priority over the proto GetPendingReviews/GetReviewHistory stubs.
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
	if pageSize > 50 {
		pageSize = 50
	}

	// BUG-226: parse order_by + seed for deterministic random ("为您推荐" 换一批).
	// order_by=random requires a validated positive uint32 seed; invalid/missing
	// seed is rejected (400) and never reaches SQL unvalidated.
	orderBy := gc.Query("order_by")
	var randomSeed *uint32
	if orderBy == "random" {
		seedStr := gc.Query("seed")
		parsed, perr := strconv.ParseUint(seedStr, 10, 32)
		if perr != nil || parsed == 0 {
			server.Fail(gc, server.ErrBadRequest, "invalid or missing seed for random order")
			return
		}
		s := uint32(parsed)
		randomSeed = &s
	}

	opts := &dto.MediaQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  gc.Query("keyword"),
		},
		OrderBy:    orderBy,
		RandomSeed: randomSeed,
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
	// BUG-195: user_id was previously ignored here (My Videos filter no-op).
	if uid := gc.Query("user_id"); uid != "" {
		opts.UserID = &uid
	}
	// BUG-105: channel filter. "__unassigned__" filters channel_id IS NULL;
	// otherwise a channel id filters by exact channel. Previously the handler
	// ignored channel_id entirely, so the My Videos channel filter was a no-op.
	if ch := gc.Query("channel_id"); ch != "" {
		if ch == "__unassigned__" {
			opts.ChannelUnassigned = true
		} else {
			opts.ChannelID = &ch
		}
	}
	if tagsStr := gc.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		opts.Tags = tags
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
		QueryOption: repotypes.QueryOption{
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
		QueryOption: repotypes.QueryOption{
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

// getMedia handles GET /medias/:token
func (h *MediaHandler) getMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	item, err := h.mediaUC.GetByShortToken(r.Context(), token)
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

// incrementViewCount handles POST /medias/:token/view
func (h *MediaHandler) incrementViewCount(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	id, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
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

// mediaVariants handles GET /medias/:token/variants
func (h *MediaHandler) mediaVariants(w http.ResponseWriter, r *http.Request) {
	if h.mediaService != nil {
		h.mediaService.MediaVariantsHTTPHandler(w, r)
		return
	}
	gc := ginadapter.GetGinContext(r)
	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}
	server.OK(gc, &pb.GetMediaVariantsResponse{
		Variants: []*types.MediaVariant{},
	})
}

// likeMedia handles POST /medias/:token/like
func (h *MediaHandler) likeMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	mediaID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
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

// unlikeMedia handles DELETE /medias/:token/like
func (h *MediaHandler) unlikeMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	mediaID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
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

// dislikeMedia handles POST /medias/:token/dislikes (BUG-222: no proto RPC exists,
// so this gin route serves the frontend toggleDislike via the real usecase).
func (h *MediaHandler) dislikeMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	mediaID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	if h.likeFavoriteUC != nil {
		stats, err := h.likeFavoriteUC.ToggleLike(r.Context(), claims.GetUserID(), mediaID, "dislike")
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

// favoriteMedia handles POST /medias/:token/favorite
func (h *MediaHandler) favoriteMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	mediaID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
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

// unfavoriteMedia handles DELETE /medias/:token/favorite
func (h *MediaHandler) unfavoriteMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	mediaID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "media not found")
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

// getMediaLikes handles GET /medias/:token/likes (plural path for frontend compatibility)
func (h *MediaHandler) getMediaLikes(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	resolvedID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		resolvedID = token
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

// getMediaFavorites handles GET /medias/:token/favorites (plural path for frontend compatibility)
func (h *MediaHandler) getMediaFavorites(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "media token is required")
		return
	}

	// Resolve short_token to internal ID
	resolvedID, err := h.mediaUC.ResolveToID(r.Context(), token)
	if err != nil {
		resolvedID = token
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

// ReviewHistoryList handles GET /api/v1/admin/medias/review/history — reviewed
// (approved) and rejected media. The repo only filters by a single
// review_status, so we query both statuses and merge, then sort by update_time
// desc and paginate in memory. Plain net/http handler mounted on the media server.
func (h *MediaHandler) ReviewHistoryList(w http.ResponseWriter, r *http.Request) {
	page, pageSize := reviewPageParams(r)

	var statuses []string
	switch r.URL.Query().Get("status") {
	case "rejected":
		statuses = []string{"rejected"}
	case "approved":
		statuses = []string{"reviewed"}
	default:
		statuses = []string{"reviewed", "rejected"}
	}

	var merged []*types.Media
	for _, rs := range statuses {
		opts := &dto.MediaQueryOption{
			QueryOption:  repotypes.QueryOption{Page: 1, PageSize: 1000},
			ReviewStatus: &rs,
		}
		items, _, err := h.mediaUC.ListMedias(r.Context(), opts)
		if err != nil {
			writeReviewError(w, err.Error())
			return
		}
		merged = append(merged, items...)
	}

	// BUG-138: honor the `keyword` param so the ReviewFlow search box works on
	// the history tab too (pending already passes keyword to ListMedias).
	if kw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("keyword"))); kw != "" {
		filtered := merged[:0]
		for _, m := range merged {
			if strings.Contains(strings.ToLower(m.GetTitle()), kw) {
				filtered = append(filtered, m)
			}
		}
		merged = filtered
	}

	sort.SliceStable(merged, func(i, j int) bool {
		return mediaUpdateTime(merged[i]).After(mediaUpdateTime(merged[j]))
	})

	total := int32(len(merged))
	start := (page - 1) * pageSize
	if start > len(merged) {
		start = len(merged)
	}
	end := start + pageSize
	if end > len(merged) {
		end = len(merged)
	}
	out := make([]map[string]any, 0, end-start)
	for _, m := range merged[start:end] {
		status := "approved"
		if m.GetReviewStatus() == "rejected" {
			status = "rejected"
		}
		out = append(out, reviewItemFromMedia(m, status))
	}
	writeReviewJSON(w, map[string]any{
		"items":     out,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// reviewPageParams parses page/page_size query params (clamped to [1,50]).
func reviewPageParams(r *http.Request) (page, pageSize int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

// writeReviewJSON writes a JSON response with 200 status.
func writeReviewJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeReviewError writes an internal-error JSON response.
func writeReviewError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 500, "message": msg})
}

// mediaUpdateTime returns the media update time (zero when unset) for sorting.
func mediaUpdateTime(m *types.Media) time.Time {
	if m == nil || m.GetUpdateTime() == nil {
		return time.Time{}
	}
	return m.GetUpdateTime().AsTime()
}

// reviewItemFromMedia maps a Media to the ReviewItem-shaped JSON the admin
// ReviewFlow page consumes. id == media_id so single-review PUT works directly.
// username/reviewer_name need a user-service lookup (out of scope in the media
// microservice) and are left empty here.
func reviewItemFromMedia(m *types.Media, reviewStatus string) map[string]any {
	return map[string]any{
		"id":            m.GetId(),
		"media_id":      m.GetId(),
		"media_title":   m.GetTitle(),
		"media_type":    m.GetType(),
		"user_id":       m.GetUserId(),
		"username":      "",
		"review_status": reviewStatus,
		"reason":        "",
		"create_time":   timestampStr(m.GetCreateTime()),
		"update_time":   timestampStr(m.GetUpdateTime()),
		"reviewer_id":   "",
		"reviewer_name": "",
	}
}

// timestampStr formats a protobuf timestamp as RFC3339 ("" when nil).
func timestampStr(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}
