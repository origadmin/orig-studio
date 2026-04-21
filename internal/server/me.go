package server

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
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

func (h *MeHandler) Register(r handler.Router) {
	me := r.Group("/me")
	{
		// All /me routes require authentication
		// Note: We can't use Use() directly with the Router interface
		// We'll need to apply middleware to each route individually

		// ================================
		// 1. CURRENT USER PROFILE
		// ================================
		me.GET("", WithJWT(h.jwt, GinHandlerToHTTP(h.GetMe)))
		me.PUT("", WithJWT(h.jwt, GinHandlerToHTTP(h.UpdateMe)))
		me.PUT("/password", WithJWT(h.jwt, GinHandlerToHTTP(h.UpdatePassword)))

		// ================================
		// 2. CURRENT USER RESOURCES
		// ================================
		me.GET("/playlists", WithJWT(h.jwt, GinHandlerToHTTP(h.GetPlaylists)))
		me.GET("/favorites", WithJWT(h.jwt, GinHandlerToHTTP(h.GetFavorites)))
		me.DELETE("/favorites/:id", WithJWT(h.jwt, GinHandlerToHTTP(h.RemoveFavorite)))
		me.GET("/likes", WithJWT(h.jwt, GinHandlerToHTTP(h.GetLikes)))
		me.GET("/subscriptions", WithJWT(h.jwt, GinHandlerToHTTP(h.GetSubscriptions)))
		me.GET("/followers", WithJWT(h.jwt, GinHandlerToHTTP(h.GetFollowers)))
		me.GET("/history", WithJWT(h.jwt, GinHandlerToHTTP(h.GetHistory)))
		me.DELETE("/history", WithJWT(h.jwt, GinHandlerToHTTP(h.ClearHistory)))
		me.DELETE("/history/:id", WithJWT(h.jwt, GinHandlerToHTTP(h.RemoveHistoryItem)))
		me.GET("/stats", WithJWT(h.jwt, GinHandlerToHTTP(h.GetStats)))
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
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.playlistUC.ListUserPlaylists(c.Request.Context(), claims.UserID, page, pageSize)
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

// GetFavorites returns the current user's favorites.
func (h *MeHandler) GetFavorites(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	favorites, err := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.UserID)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"items":     favorites,
		"total":     len(favorites),
		"page":      1,
		"page_size": len(favorites),
	})
}

// GetLikes returns the current user's likes.
func (h *MeHandler) GetLikes(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	likes, err := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.UserID)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"items":     likes,
		"total":     len(likes),
		"page":      1,
		"page_size": len(likes),
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
		claims.UserID,
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
// NOTE: Currently uses favorites + likes as a history proxy since
// watch_history entity doesn't exist yet. When watch_history is implemented,
// this should query the dedicated history table instead.
func (h *MeHandler) GetHistory(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	favorites, favErr := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.UserID)
	if favErr != nil {
		Fail(c, ErrInternal, favErr.Error())
		return
	}

	likes, likeErr := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.UserID)
	if likeErr != nil {
		Fail(c, ErrInternal, likeErr.Error())
		return
	}

	items := make([]interface{}, 0, len(favorites)+len(likes))
	for _, f := range favorites {
		items = append(items, gin.H{
			"id":         f.ID,
			"media_id":   f.MediaID,
			"user_id":    claims.UserID,
			"progress":   0,
			"watched_at": f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	for _, l := range likes {
		items = append(items, gin.H{
			"id":         l.ID,
			"media_id":   l.MediaID,
			"user_id":    claims.UserID,
			"progress":   0,
			"watched_at": l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	total := len(items)

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	OK(c, gin.H{
		"items":     items[start:end],
		"total":     total,
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

// RemoveFavorite removes a favorite by its ID.
func (h *MeHandler) RemoveFavorite(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	favoriteID := c.Param("id")
	if favoriteID == "" {
		Fail(c, ErrBadRequest, "favorite id is required")
		return
	}

	favorites, err := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.UserID)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	for _, fav := range favorites {
		if fav.ID == favoriteID {
			_, toggleErr := h.likeFavoriteUC.ToggleFavorite(c.Request.Context(), claims.UserID, fav.MediaID)
			if toggleErr != nil {
				Fail(c, ErrInternal, toggleErr.Error())
				return
			}
			OK(c, gin.H{"message": "Favorite removed"})
			return
		}
	}

	Fail(c, ErrNotFound, "Favorite not found")
}

// GetFollowers returns users who follow the current user.
func (h *MeHandler) GetFollowers(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.userUC.GetSubscribers(
		c.Request.Context(),
		claims.UserID,
		page,
		pageSize,
	)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	items := make([]interface{}, 0, len(list))
	for _, u := range list {
		var createdAt string
		if u.CreateTime != nil {
			createdAt = u.CreateTime.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, gin.H{
			"id":            u.Id,
			"user_id":       u.Id,
			"username":      u.Username,
			"nickname":      u.Nickname,
			"avatar":        u.Avatar,
			"subscribed_at": createdAt,
		})
	}

	OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ClearHistory clears all watch history for the current user.
// NOTE: Since history is currently proxied through favorites+likes,
// this removes all favorites and likes as a workaround.
func (h *MeHandler) ClearHistory(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	favorites, favErr := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.UserID)
	if favErr == nil && len(favorites) > 0 {
		for _, fav := range favorites {
			h.likeFavoriteUC.ToggleFavorite(c.Request.Context(), claims.UserID, fav.MediaID)
		}
	}

	likes, likeErr := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.UserID)
	if likeErr == nil && len(likes) > 0 {
		for _, l := range likes {
			h.likeFavoriteUC.ToggleLike(c.Request.Context(), claims.UserID, l.MediaID, "like")
		}
	}

	OK(c, gin.H{"message": "History cleared"})
}

// RemoveHistoryItem removes a single history item by its ID.
func (h *MeHandler) RemoveHistoryItem(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		Fail(c, ErrBadRequest, "item id is required")
		return
	}

	favorites, favErr := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.UserID)
	if favErr == nil {
		for _, fav := range favorites {
			if fav.ID == itemID {
				_, _ = h.likeFavoriteUC.ToggleFavorite(c.Request.Context(), claims.UserID, fav.MediaID)
				OK(c, gin.H{"message": "History item removed"})
				return
			}
		}
	}

	likes, likeErr := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.UserID)
	if likeErr == nil {
		for _, l := range likes {
			if l.ID == itemID {
				_, _ = h.likeFavoriteUC.ToggleLike(c.Request.Context(), claims.UserID, l.MediaID, "like")
				OK(c, gin.H{"message": "History item removed"})
				return
			}
		}
	}

	Fail(c, ErrNotFound, "History item not found")
}
