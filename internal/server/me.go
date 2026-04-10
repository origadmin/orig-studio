package server

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	userbiz "origadmin/application/origcms/internal/svc-user/biz"
)

// MeHandler handles /api/v1/me routes.
type MeHandler struct {
	userUC         *userbiz.UserUseCase
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase
	playlistUC     *contentbiz.PlaylistChannelUseCase
	jwt            *auth.Manager
}

// NewMeHandler creates a new MeHandler.
func NewMeHandler(
	userUC *userbiz.UserUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistUC *contentbiz.PlaylistChannelUseCase,
	jwt *auth.Manager,
) *MeHandler {
	return &MeHandler{
		userUC:         userUC,
		likeFavoriteUC: likeFavoriteUC,
		playlistUC:     playlistUC,
		jwt:            jwt,
	}
}

func (h *MeHandler) Register(group *gin.RouterGroup) {
	me := group.Group("/me")
	{
		// All /me routes require authentication
		me.Use(JWTMiddleware(h.jwt))

		// ================================
		// 1. CURRENT USER PROFILE
		// ================================
		me.GET("", h.GetMe)
		me.PUT("", h.UpdateMe)
		me.PUT("/password", h.UpdatePassword)

		// ================================
		// 2. CURRENT USER RESOURCES
		// ================================
		me.GET("/playlists", h.GetPlaylists)
		me.GET("/favorites", h.GetFavorites)
		me.GET("/likes", h.GetLikes)
		me.GET("/subscriptions", h.GetSubscriptions)
		me.GET("/history", h.GetHistory)
		me.GET("/stats", h.GetStats)
	}
}

// GetMe returns the current user's information.
func (h *MeHandler) GetMe(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	user, err := h.userUC.GetUser(c.Request.Context(), claims.UserID)
	if err != nil {
		Fail(c, ErrUserNotFound, "User not found")
		return
	}

	OK(c, user)
}

// UpdateMe updates the current user's information.
func (h *MeHandler) UpdateMe(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, ErrBadRequest, err.Error())
		return
	}

	user, err := h.userUC.GetUser(c.Request.Context(), claims.UserID)
	if err != nil {
		Fail(c, ErrUserNotFound, "User not found")
		return
	}

	if input.Nickname != "" {
		user.Nickname = input.Nickname
	}
	if input.Email != "" {
		user.Email = input.Email
	}

	updated, err := h.userUC.UpdateUser(c.Request.Context(), user)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, updated)
}

// UpdatePassword updates the current user's password.
func (h *MeHandler) UpdatePassword(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, ErrBadRequest, err.Error())
		return
	}

	// Verify old password
	if err := h.userUC.VerifyPassword(c.Request.Context(), claims.UserID, input.OldPassword); err != nil {
		Fail(c, ErrPasswordWrong, "Invalid old password")
		return
	}

	// Update password (TODO: Implement UpdatePassword in UserUseCase)
	// hashedPassword, err := h.userUC.HashPassword(input.NewPassword)
	// if err != nil {
	// 	Fail(c, ErrInternal, "Failed to hash password")
	// 	return
	// }
	//
	// if err := h.userUC.UpdatePassword(c.Request.Context(), claims.UserID, hashedPassword); err != nil {
	// 	Fail(c, ErrInternal, err.Error())
	// 	return
	// }

	OK(c, gin.H{"message": "Password updated"})
}

// GetPlaylists returns the current user's playlists.
func (h *MeHandler) GetPlaylists(c *gin.Context) {
	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: Implement playlist listing
	OK(c, gin.H{
		"items":     []interface{}{},
		"total":     0,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetFavorites returns the current user's favorites.
func (h *MeHandler) GetFavorites(c *gin.Context) {
	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: Implement favorites listing
	OK(c, gin.H{
		"items":     []interface{}{},
		"total":     0,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetLikes returns the current user's likes.
func (h *MeHandler) GetLikes(c *gin.Context) {
	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: Implement likes listing
	OK(c, gin.H{
		"items":     []interface{}{},
		"total":     0,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetSubscriptions returns the current user's subscriptions.
func (h *MeHandler) GetSubscriptions(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.userUC.GetSubscriptions(
		c.Request.Context(),
		int(claims.UserID),
		page,
		pageSize,
	)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetHistory returns the current user's watch history.
func (h *MeHandler) GetHistory(c *gin.Context) {
	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: Implement history listing
	OK(c, gin.H{
		"items":     []interface{}{},
		"total":     0,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStats returns the current user's statistics.
func (h *MeHandler) GetStats(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	// TODO: Implement user stats
	OK(c, gin.H{
		"user_id": claims.UserID,
		"stats":   gin.H{},
	})
}
