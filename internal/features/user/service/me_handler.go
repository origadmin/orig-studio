/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/infra/auth"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	userbiz "origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/server"
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

// RegisterRoutes registers the handler's routes.
func (h *MeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	me := rg.Group("/me")
	me.Use(server.JWTMiddleware(h.jwt))
	{
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
		me.POST("/playlists", h.CreatePlaylist)
		me.PATCH("/playlists/:id", h.UpdatePlaylist)
		me.DELETE("/playlists/:id", h.DeletePlaylist)
		me.POST("/playlists/:id/media", h.AddMediaToPlaylist)
		me.DELETE("/playlists/:id/media/:mediaId", h.RemoveMediaFromPlaylist)
		me.GET("/favorites", h.GetFavorites)
		me.DELETE("/favorites/:id", h.RemoveFavorite)
		me.GET("/likes", h.GetLikes)
		me.GET("/subscriptions", h.GetSubscriptions)
		me.GET("/followers", h.GetFollowers)
		me.GET("/history", h.GetHistory)
		me.DELETE("/history", h.ClearHistory)
		me.DELETE("/history/:id", h.RemoveHistoryItem)
		me.GET("/stats", h.GetStats)
	}
}

// GetMe returns the current user's information.
func (h *MeHandler) GetMe(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	user, err := h.userUC.GetUser(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrUserNotFound, "User not found")
		return
	}

	server.OK(c, user)
}

// UpdateMe updates the current user's information.
func (h *MeHandler) UpdateMe(c *gin.Context) {
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

	user, err := h.userUC.GetUser(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrUserNotFound, "User not found")
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
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, updated)
}

// UpdatePassword updates the current user's password.
func (h *MeHandler) UpdatePassword(c *gin.Context) {
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
	if err := h.userUC.VerifyPassword(c.Request.Context(), claims.GetUserID(), input.OldPassword); err != nil {
		server.Fail(c, server.ErrPasswordWrong, "Invalid old password")
		return
	}

	// TODO: Implement UpdatePassword in UserUseCase
	server.OK(c, gin.H{"message": "Password updated"})
}

// GetPlaylists returns the current user's playlists.
func (h *MeHandler) GetPlaylists(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.playlistUC.ListUserPlaylists(c.Request.Context(), claims.GetUserID(), page, pageSize)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreatePlaylist creates a new playlist for the current user.
func (h *MeHandler) CreatePlaylist(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	if input.Title == "" {
		server.Fail(c, server.ErrBadRequest, "title is required")
		return
	}

	if input.Description == "" {
		input.Description = " "
	}

	p, err := h.playlistUC.CreatePlaylist(c.Request.Context(), &contentbiz.Playlist{
		Title:       input.Title,
		Description: input.Description,
		UserID:      claims.GetUserID(),
	})
	if err != nil {
		_ = c.Error(fmt.Errorf("CreatePlaylist failed: userID=%q title=%q err=%w", claims.GetUserID(), input.Title, err))
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{"playlist": p})
}

// AddMediaToPlaylist adds a media item to a playlist.
func (h *MeHandler) AddMediaToPlaylist(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	playlistID := c.Param("id")
	if playlistID == "" {
		server.Fail(c, server.ErrBadRequest, "playlist id is required")
		return
	}

	var input struct {
		MediaID string `json:"media_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	if input.MediaID == "" {
		server.Fail(c, server.ErrBadRequest, "media_id is required")
		return
	}

	if err := h.playlistUC.AddMediaToPlaylist(c.Request.Context(), playlistID, input.MediaID, claims.GetUserID(), false); err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{"message": "media added to playlist"})
}

// UpdatePlaylist updates a playlist's title (and optionally description/is_public).
func (h *MeHandler) UpdatePlaylist(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	playlistID := c.Param("id")
	if playlistID == "" {
		server.Fail(c, server.ErrBadRequest, "playlist id is required")
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	existing, err := h.playlistUC.GetPlaylist(c.Request.Context(), playlistID)
	if err != nil {
		server.Fail(c, server.ErrNotFound, "playlist not found")
		return
	}

	if input.Title != "" {
		existing.Title = input.Title
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.IsPublic != nil {
		existing.IsPublic = *input.IsPublic
	}

	updated, err := h.playlistUC.UpdatePlaylist(c.Request.Context(), existing, claims.GetUserID(), false)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, updated)
}

// DeletePlaylist deletes a playlist owned by the current user.
func (h *MeHandler) DeletePlaylist(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	playlistID := c.Param("id")
	if playlistID == "" {
		server.Fail(c, server.ErrBadRequest, "playlist id is required")
		return
	}

	if err := h.playlistUC.DeletePlaylist(c.Request.Context(), playlistID, claims.GetUserID(), false); err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{"message": "playlist deleted successfully"})
}

// RemoveMediaFromPlaylist removes a media item from a playlist.
func (h *MeHandler) RemoveMediaFromPlaylist(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	playlistID := c.Param("id")
	mediaID := c.Param("mediaId")
	if playlistID == "" || mediaID == "" {
		server.Fail(c, server.ErrBadRequest, "playlist id and media id are required")
		return
	}

	if err := h.playlistUC.RemoveMediaFromPlaylist(c.Request.Context(), playlistID, mediaID, claims.GetUserID(), false); err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{"message": "media removed from playlist"})
}

func (h *MeHandler) GetFavorites(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	favorites, err := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{
		"items":     favorites,
		"total":     len(favorites),
		"page":      1,
		"page_size": len(favorites),
	})
}

// GetLikes returns the current user's likes.
func (h *MeHandler) GetLikes(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	likes, err := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{
		"items":     likes,
		"total":     len(likes),
		"page":      1,
		"page_size": len(likes),
	})
}

// GetSubscriptions returns the current user's subscriptions.
func (h *MeHandler) GetSubscriptions(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.userUC.GetSubscriptions(
		c.Request.Context(),
		claims.GetUserID(),
		page,
		pageSize,
	)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.OK(c, gin.H{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetHistory returns the current user's watch history.
func (h *MeHandler) GetHistory(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	favorites, favErr := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.GetUserID())
	if favErr != nil {
		server.Fail(c, server.ErrInternal, favErr.Error())
		return
	}

	likes, likeErr := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.GetUserID())
	if likeErr != nil {
		server.Fail(c, server.ErrInternal, likeErr.Error())
		return
	}

	items := make([]interface{}, 0, len(favorites)+len(likes))
	for _, f := range favorites {
		items = append(items, gin.H{
			"id":         f.ID,
			"media_id":   f.MediaID,
			"user_id":    claims.GetUserID(),
			"progress":   0,
			"watched_at": f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	for _, l := range likes {
		items = append(items, gin.H{
			"id":         l.ID,
			"media_id":   l.MediaID,
			"user_id":    claims.GetUserID(),
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

	server.OK(c, gin.H{
		"items":     items[start:end],
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStats returns the current user's statistics.
func (h *MeHandler) GetStats(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	// TODO: Implement user stats
	server.OK(c, gin.H{
		"user_id": claims.GetUserID(),
		"stats":   gin.H{},
	})
}

// RemoveFavorite removes a favorite by its ID.
func (h *MeHandler) RemoveFavorite(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	favoriteID := c.Param("id")
	if favoriteID == "" {
		server.Fail(c, server.ErrBadRequest, "favorite id is required")
		return
	}

	favorites, err := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	for _, fav := range favorites {
		if fav.ID == favoriteID {
			_, toggleErr := h.likeFavoriteUC.ToggleFavorite(c.Request.Context(), claims.GetUserID(), fav.MediaID)
			if toggleErr != nil {
				server.Fail(c, server.ErrInternal, toggleErr.Error())
				return
			}
			server.OK(c, gin.H{"message": "Favorite removed"})
			return
		}
	}

	server.Fail(c, server.ErrNotFound, "Favorite not found")
}

// GetFollowers returns users who follow the current user.
func (h *MeHandler) GetFollowers(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.userUC.GetSubscribers(
		c.Request.Context(),
		claims.GetUserID(),
		page,
		pageSize,
	)
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
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

	server.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ClearHistory clears all watch history for the current user.
func (h *MeHandler) ClearHistory(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	favorites, favErr := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.GetUserID())
	if favErr == nil && len(favorites) > 0 {
		for _, fav := range favorites {
			h.likeFavoriteUC.ToggleFavorite(c.Request.Context(), claims.GetUserID(), fav.MediaID)
		}
	}

	likes, likeErr := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.GetUserID())
	if likeErr == nil && len(likes) > 0 {
		for _, l := range likes {
			h.likeFavoriteUC.ToggleLike(c.Request.Context(), claims.GetUserID(), l.MediaID, "like")
		}
	}

	server.OK(c, gin.H{"message": "History cleared"})
}

// RemoveHistoryItem removes a single history item by its ID.
func (h *MeHandler) RemoveHistoryItem(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		server.Fail(c, server.ErrBadRequest, "item id is required")
		return
	}

	favorites, favErr := h.likeFavoriteUC.ListUserFavorites(c.Request.Context(), claims.GetUserID())
	if favErr == nil {
		for _, fav := range favorites {
			if fav.ID == itemID {
				_, _ = h.likeFavoriteUC.ToggleFavorite(c.Request.Context(), claims.GetUserID(), fav.MediaID)
				server.OK(c, gin.H{"message": "History item removed"})
				return
			}
		}
	}

	likes, likeErr := h.likeFavoriteUC.ListUserLikes(c.Request.Context(), claims.GetUserID())
	if likeErr == nil {
		for _, l := range likes {
			if l.ID == itemID {
				_, _ = h.likeFavoriteUC.ToggleLike(c.Request.Context(), claims.GetUserID(), l.MediaID, "like")
				server.OK(c, gin.H{"message": "History item removed"})
				return
			}
		}
	}

	server.Fail(c, server.ErrNotFound, "History item not found")
}
