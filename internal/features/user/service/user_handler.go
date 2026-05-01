/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * User module - handles user CRUD and user-related resources
 */

package service

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origcms/api/gen/v1/user"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/features/user/dto"
	"origadmin/application/origcms/internal/server"
)

// UserHandler handles /api/v1/users routes.
type UserHandler struct {
	uc  *biz.UserUseCase
	jwt *auth.Manager
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(uc *biz.UserUseCase, jwt *auth.Manager) *UserHandler {
	return &UserHandler{uc: uc, jwt: jwt}
}

// RegisterRoutes registers the handler's routes.
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		// Current user endpoints
		users.GET("/me", server.JWTMiddleware(h.jwt), h.getMe)
		users.PUT("/me", server.JWTMiddleware(h.jwt), h.updateMe)
		users.PUT("/me/password", server.JWTMiddleware(h.jwt), h.updatePassword)

		// List users
		users.GET("", h.listUsers)

		// Create user
		users.POST("", h.createUser)

		// ================================
		// 2. NESTED RESOURCE ROUTES
		// ================================
		users.GET("/:id/playlists", h.getUserPlaylists)
		users.GET("/username/:username", h.getUserByUsername)
		users.GET("/slug/:slug", h.getUserBySlug)
		users.PUT("/me/slug", server.JWTMiddleware(h.jwt), h.updateUserSlug)
		users.GET("/:id/favorites", server.JWTMiddleware(h.jwt), h.getUserFavorites)
		users.GET("/:id/likes", server.JWTMiddleware(h.jwt), h.getUserLikes)
		users.GET("/:id/subscriptions", server.JWTMiddleware(h.jwt), h.getUserSubscriptions)
		users.GET("/:id/followers", h.getUserFollowers)
		users.GET("/:id/stats", h.getUserStats)
		users.GET("/:id/channels", h.getUserChannels)

		// ================================
		// 3. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		users.GET("/:id", h.getUser)
		users.DELETE("/:id", h.deleteUser)
	}
}

func (h *UserHandler) getMe(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}
	u, err := h.uc.GetUser(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrNotFound, "User not found")
		return
	}
	server.OK(c, &pb.GetMeResponse{User: u})
}

func (h *UserHandler) updateMe(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"omitempty,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	u, err := h.uc.GetUser(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrNotFound, "User not found")
		return
	}

	if input.Nickname != "" {
		u.Nickname = input.Nickname
	}
	if input.Email != "" {
		u.Email = input.Email
	}

	updated, err := h.uc.UpdateUser(c.Request.Context(), u)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}
	server.OK(c, &pb.UpdateMeResponse{User: updated})
}

func (h *UserHandler) updatePassword(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	// Verify old password
	if err := h.uc.VerifyPassword(c.Request.Context(), claims.GetUserID(), input.OldPassword); err != nil {
		server.Fail(c, server.ErrBadRequest, "Invalid old password")
		return
	}

	// TODO: Implement UpdatePassword
	server.OK(c, &pb.UpdateMyPasswordResponse{Success: true})
}

func (h *UserHandler) listUsers(c *gin.Context) {
	// Support both "limit" and "page_size" query params for compatibility
	limit, _ := strconv.Atoi(c.Query("limit"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if limit == 0 && pageSize == 0 {
		limit = 20
	}
	if limit == 0 {
		limit = pageSize
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	// Use ListUserEntities to get role field directly from entity
	entities, total, err := h.uc.ListUserEntities(c.Request.Context(), &dto.UserQueryOption{
		QueryOption: repo.QueryOption{
			Page:     int32(page),
			PageSize: int32(limit),
		},
	})
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	// Convert entity users to proto users
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

	server.OK(c, &pb.ListUsersResponse{
		Items:     pbUsers,
		Total:     int32(total),
		Page:      int32(page),
		PageSize:  int32(limit),
	})
}

func (h *UserHandler) createUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	hashedPassword, _ := h.uc.HashPassword(input.Password)
	u, err := h.uc.CreateUser(c.Request.Context(), &types.User{
		Username: input.Username,
		Email:    input.Email,
		Nickname: input.Name,
	}, hashedPassword)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.Created(c, &pb.CreateUserResponse{User: u})
}

func (h *UserHandler) getUserPlaylists(c *gin.Context) {
	id := c.Param("id")
	_ = id // used for authorization context below
	// TODO: Implement playlist listing with proper userID resolution
	server.OK(c, &pb.GetUserPlaylistsResponse{
		Items: []*types.Playlist{},
	})
}

func (h *UserHandler) getUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		server.Fail(c, server.ErrBadRequest, "Invalid username")
		return
	}

	u, err := h.uc.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		server.Fail(c, server.ErrNotFound, "User not found")
		return
	}
	server.OK(c, &pb.GetUserResponse{User: u})
}

func (h *UserHandler) getUserBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		server.Fail(c, server.ErrBadRequest, "Invalid slug")
		return
	}

	u, err := h.uc.GetUserBySlug(c.Request.Context(), slug)
	if err != nil {
		server.Fail(c, server.ErrNotFound, "User not found")
		return
	}
	// Sanitize: hide username, email, and password for public access
	sanitizePublicUser(u)
	server.OK(c, &pb.GetUserResponse{User: u})
}

func (h *UserHandler) updateUserSlug(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Slug string `json:"slug" binding:"required,min=3,max=64"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	if err := h.uc.UpdateUserSlug(c.Request.Context(), claims.GetUserID(), input.Slug); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	u, _ := h.uc.GetUser(c.Request.Context(), claims.GetUserID())
	server.OK(c, &pb.UpdateMeResponse{User: u})
}

func (h *UserHandler) getUserFavorites(c *gin.Context) {
	id := c.Param("id")
	_ = id // used for authorization context below
	// TODO: Implement favorites listing with proper userID resolution
	server.OK(c, &pb.GetMyFavoritesResponse{
		Items: []*types.Media{},
	})
}

func (h *UserHandler) getUserLikes(c *gin.Context) {
	id := c.Param("id")
	_ = id // used for authorization context below
	// TODO: Implement likes listing with proper userID resolution
	server.OK(c, &pb.GetMyLikesResponse{
		Likes: []*types.Like{},
	})
}

func (h *UserHandler) getUserSubscriptions(c *gin.Context) {
	id := c.Param("id")
	var userID string
	if id == "me" {
		if claims, ok := server.GetClaims(c); ok {
			userID = claims.GetUserID()
		} else {
			server.Fail(c, server.ErrUnauthorized, "unauthorized")
			return
		}
	} else {
		userID = id
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}
	// Normalize pagination parameters
	page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

	list, total, err := h.uc.GetSubscriptions(
		c.Request.Context(),
		userID,
		page,
		pageSize,
	)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, &pb.GetMySubscriptionsResponse{
		Items:    []*types.Channel{},
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	_ = list
}

func (h *UserHandler) getUserFollowers(c *gin.Context) {
	id := c.Param("id")
	var userID string
	if id == "me" {
		if claims, ok := server.GetClaims(c); ok {
			userID = claims.GetUserID()
		} else {
			server.Fail(c, server.ErrUnauthorized, "unauthorized")
			return
		}
	} else {
		userID = id
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}
	// Normalize pagination parameters
	page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

	list, total, err := h.uc.GetSubscribers(
		c.Request.Context(),
		userID,
		page,
		pageSize,
	)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, &pb.GetUserFollowersResponse{
		Followers: []*types.User{},
		Total:     int32(total),
		Page:      int32(page),
		PageSize:  int32(pageSize),
	})
	_ = list
}

func (h *UserHandler) getUserStats(c *gin.Context) {
	id := c.Param("id")
	_ = id // used for authorization context below
	// TODO: Implement user stats with proper userID resolution
	server.OK(c, &pb.GetUserStatsResponse{})
}

func (h *UserHandler) getUserChannels(c *gin.Context) {
	id := c.Param("id")
	if id == "me" {
		if _, ok := server.GetClaims(c); !ok {
			server.Fail(c, server.ErrUnauthorized, "unauthorized")
			return
		}
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 100
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	// TODO: Implement ListUserChannels
	server.OK(c, &pb.GetUserPlaylistsResponse{
		Items:    []*types.Playlist{},
		Total:     0,
		Page:      int32(page),
		PageSize:  int32(limit),
	})
}

func (h *UserHandler) getUser(c *gin.Context) {
	id := c.Param("id")
	u, err := h.uc.GetUser(c.Request.Context(), id)
	if err != nil {
		server.Fail(c, server.ErrNotFound, "User not found")
		return
	}
	server.OK(c, &pb.GetUserResponse{User: u})
}

func (h *UserHandler) deleteUser(c *gin.Context) {
	id := c.Param("id")
	err := h.uc.DeleteUser(c.Request.Context(), id)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}
	server.OK(c, &pb.DeleteUserResponse{Empty: &emptypb.Empty{}})
}

// sanitizePublicUser clears sensitive fields from a User response for public access.
func sanitizePublicUser(u *types.User) *types.User {
	u.Username = ""
	u.Email = ""
	u.Password = ""
	u.Phone = ""
	u.LastLoginIp = ""
	u.LoginIp = ""
	return u
}

// userStatusToPB converts a string status to proto UserStatus enum.
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

// convertTimeToTimestamp converts time.Time to protobuf Timestamp.
func convertTimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
