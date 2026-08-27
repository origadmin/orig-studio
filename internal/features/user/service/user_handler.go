/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * User module - handles user CRUD and user-related public resources
 */

package service

import (
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/user"
	"origadmin/application/origstudio/api/gen/v1/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/user/biz"
	"origadmin/application/origstudio/internal/features/user/dto"
	"origadmin/application/origstudio/internal/server"
)

type UserHandler struct {
	uc             *biz.UserUseCase
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase
	playlistUC     *contentbiz.PlaylistChannelUseCase
	jwt            *auth.Manager
}

func NewUserHandler(
	uc *biz.UserUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistUC *contentbiz.PlaylistChannelUseCase,
	jwt *auth.Manager,
) *UserHandler {
	return &UserHandler{
		uc:             uc,
		likeFavoriteUC: likeFavoriteUC,
		playlistUC:     playlistUC,
		jwt:            jwt,
	}
}

func (h *UserHandler) RegisterRoutes(r http2.Router) {
	users := r.Group("/users")
	{
		users.GET("/by-username", h.getUserByUsername)
		users.GET("", h.listUsers)
		users.POST("", h.createUser)
		users.GET("/:slug/playlists", h.getUserPlaylists)
		users.GET("/:slug/favorites", server.WithJWTCtx(h.jwt, h.getUserFavorites))
		users.GET("/:slug/likes", server.WithJWTCtx(h.jwt, h.getUserLikes))
		users.GET("/:slug/subscriptions", server.WithJWTCtx(h.jwt, h.getUserSubscriptions))
		users.GET("/:slug/followers", h.getUserFollowers)
		users.GET("/:slug/stats", h.getUserStats)
		users.GET("/:slug/channels", h.getUserChannels)
		users.GET("/:slug", h.getUser)
		users.DELETE("/:slug", h.deleteUser)
	}
}

func (h *UserHandler) getUserByUsername(ctx http2.Context) error {
	username := ctx.QueryVar("username")
	if username == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "username is required")
	}
	u, err := h.uc.GetUserByUsername(ctx, username)
	if err != nil {
		return server.FailCtx(ctx, server.ErrNotFound, "User not found")
	}
	return server.OKCtx(ctx, &pb.GetUserResponse{User: u})
}

func (h *UserHandler) listUsers(ctx http2.Context) error {
	limit, _ := strconv.Atoi(ctx.QueryVar("limit"))
	pageSize, _ := strconv.Atoi(ctx.QueryVar("page_size"))
	if limit == 0 && pageSize == 0 {
		limit = 20
	}
	if limit == 0 {
		limit = pageSize
	}
	page, _ := strconv.Atoi(ctx.QueryVar("page"))
	if page == 0 {
		page = 1
	}
	page, limit = repotypes.NormalizeHTTPPagination(page, limit)

	entities, total, err := h.uc.ListUserEntities(ctx, &dto.UserQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(limit),
		},
	})
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	pbUsers := make([]*types.User, len(entities))
	for i, u := range entities {
		pbUsers[i] = &types.User{
			Id:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Avatar:      u.Logo,
			Role:        string(u.Role),
			Status:      userStatusToPB(string(u.Status)),
			DateJoined:  convertTimeToTimestamp(u.DateJoined),
		}
	}

	return server.OKCtx(ctx, &pb.ListUsersResponse{
		Items:     pbUsers,
		Total:     int32(total),
		Page:      int32(page),
		PageSize:  int32(limit),
	})
}

func (h *UserHandler) createUser(ctx http2.Context) error {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	hashedPassword, _ := h.uc.HashPassword(input.Password)
	u, err := h.uc.CreateUser(ctx, &types.User{
		Username: input.Username,
		Email:    input.Email,
		Nickname: input.Name,
	}, hashedPassword)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.CreatedCtx(ctx, &pb.CreateUserResponse{User: u})
}

func (h *UserHandler) getUserPlaylists(ctx http2.Context) error {
	slug := ctx.Var("slug")
	var userID string
	var isOwner bool

	if slug == "me" {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID = claims.GetUserID()
		isOwner = true
	} else {
		u, err := h.uc.GetUserBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "User not found")
		}
		userID = u.GetId()
		if claims, ok := server.GetClaimsCtx(ctx); ok && claims.GetUserID() == userID {
			isOwner = true
		}
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

	var (
		list  []*contentbiz.Playlist
		total int
		err   error
	)
	if isOwner {
		list, total, err = h.playlistUC.ListUserPlaylists(ctx, userID, page, pageSize)
	} else {
		list, total, err = h.playlistUC.ListUserPublicPlaylists(ctx, userID, page, pageSize)
	}
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UserHandler) getUserFavorites(ctx http2.Context) error {
	slug := ctx.Var("slug")
	var userID string
	var isOwner bool

	if slug == "me" {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID = claims.GetUserID()
		isOwner = true
	} else {
		u, err := h.uc.GetUserBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "User not found")
		}
		userID = u.GetId()
		if claims, ok := server.GetClaimsCtx(ctx); ok && claims.GetUserID() == userID {
			isOwner = true
		}
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

	if !isOwner {
		return server.OKCtx(ctx, map[string]any{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}

	favorites, total, err := h.likeFavoriteUC.ListUserFavoritesPaginated(ctx, userID, page, pageSize)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     favorites,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UserHandler) getUserLikes(ctx http2.Context) error {
	slug := ctx.Var("slug")
	var userID string
	var isOwner bool

	if slug == "me" {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID = claims.GetUserID()
		isOwner = true
	} else {
		u, err := h.uc.GetUserBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "User not found")
		}
		userID = u.GetId()
		if claims, ok := server.GetClaimsCtx(ctx); ok && claims.GetUserID() == userID {
			isOwner = true
		}
	}

	if !isOwner {
		return server.OKCtx(ctx, map[string]any{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 0,
		})
	}

	likes, err := h.likeFavoriteUC.ListUserLikes(ctx, userID)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     likes,
		"total":     len(likes),
		"page":      1,
		"page_size": len(likes),
	})
}

func (h *UserHandler) getUserSubscriptions(ctx http2.Context) error {
	slug := ctx.Var("slug")
	var userID string
	var isOwner bool

	if slug == "me" {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID = claims.GetUserID()
		isOwner = true
	} else {
		u, err := h.uc.GetUserBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "User not found")
		}
		userID = u.GetId()
		if claims, ok := server.GetClaimsCtx(ctx); ok && claims.GetUserID() == userID {
			isOwner = true
		}
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

	if !isOwner {
		return server.OKCtx(ctx, map[string]any{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}

	list, total, err := h.uc.GetSubscriptions(
		ctx,
		userID,
		page,
		pageSize,
	)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UserHandler) getUserFollowers(ctx http2.Context) error {
	slug := ctx.Var("slug")
	var userID string

	if slug == "me" {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID = claims.GetUserID()
	} else {
		u, err := h.uc.GetUserBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "User not found")
		}
		userID = u.GetId()
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

	list, total, err := h.uc.GetSubscribers(
		ctx,
		userID,
		page,
		pageSize,
	)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	items := make([]interface{}, 0, len(list))
	for _, u := range list {
		var createdAt string
		if u.CreateTime != nil {
			createdAt = u.CreateTime.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, map[string]any{
			"id":            u.Id,
			"user_id":       u.Id,
			"username":      u.Username,
			"nickname":      u.Nickname,
			"avatar":        u.Avatar,
			"subscribed_at": createdAt,
		})
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UserHandler) getUserStats(ctx http2.Context) error {
	slug := ctx.Var("slug")
	u, err := h.uc.GetUserBySlug(ctx, slug)
	if err != nil {
		return server.FailCtx(ctx, server.ErrNotFound, "User not found")
	}
	return server.OKCtx(ctx, map[string]any{
		"user_id": u.GetId(),
		"stats":   map[string]any{},
	})
}

func (h *UserHandler) getUserChannels(ctx http2.Context) error {
	slug := ctx.Var("slug")
	var userID string
	var isOwner bool

	if slug == "me" {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID = claims.GetUserID()
		isOwner = true
	} else {
		u, err := h.uc.GetUserBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "User not found")
		}
		userID = u.GetId()
		if claims, ok := server.GetClaimsCtx(ctx); ok && claims.GetUserID() == userID {
			isOwner = true
		}
	}

	limit, _ := strconv.Atoi(ctx.QueryVarDefault("limit", "100"))
	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	page, limit = repotypes.NormalizeHTTPPagination(page, limit)

	channels, total, err := h.playlistUC.ListUserChannels(ctx, userID, page, limit)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	_ = isOwner
	return server.OKCtx(ctx, map[string]any{
		"items":     channels,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

func (h *UserHandler) getUser(ctx http2.Context) error {
	slug := ctx.Var("slug")
	u, err := h.uc.GetUserBySlug(ctx, slug)
	if err != nil {
		return server.FailCtx(ctx, server.ErrNotFound, "User not found")
	}
	sanitizePublicUser(u)
	return server.OKCtx(ctx, &pb.GetUserResponse{User: u})
}

func (h *UserHandler) deleteUser(ctx http2.Context) error {
	slug := ctx.Var("slug")
	u, err := h.uc.GetUserBySlug(ctx, slug)
	if err != nil {
		return server.FailCtx(ctx, server.ErrNotFound, "User not found")
	}
	err = h.uc.DeleteUser(ctx, u.GetId())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}
	return server.OKCtx(ctx, &pb.DeleteUserResponse{Empty: &emptypb.Empty{}})
}

func sanitizePublicUser(u *types.User) *types.User {
	u.Username = ""
	u.Email = ""
	u.Password = ""
	u.Phone = ""
	u.LastLoginIp = ""
	u.LoginIp = ""
	return u
}

func userStatusToPB(status string) types.UserStatus {
	switch status {
	case "ACTIVE":
		return types.UserStatus_USER_STATUS_ACTIVE
	case "INACTIVE":
		return types.UserStatus_USER_STATUS_INACTIVE
	case "PENDING":
		return types.UserStatus_USER_STATUS_PENDING
	case "BANNED", "SUSPENDED":
		return types.UserStatus_USER_STATUS_SUSPENDED
	default:
		return types.UserStatus_USER_STATUS_ACTIVE
	}
}

func convertTimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
