/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel module - handles channel CRUD and subscription management
 *
 * API paths:
 * - /api/v1/channels              - channel collection
 * - /api/v1/channels/:id          - single channel
 * - /api/v1/channels/user/:userId - user's channels
 * - /api/v1/channels/:id/subscribers - channel subscribers
 * - /api/v1/channels/:id/subscription - subscription status/operations
 * - /api/v1/channels/:id/media    - media management
 */

package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
)

// ChannelHandler handles /api/v1/channels routes.
type ChannelHandler struct {
	uc  *biz.PlaylistChannelUseCase
	jwt *auth.Manager
}

// NewChannelHandler creates a new ChannelHandler.
func NewChannelHandler(uc *biz.PlaylistChannelUseCase, jwt *auth.Manager) *ChannelHandler {
	return &ChannelHandler{uc: uc, jwt: jwt}
}

func (h *ChannelHandler) Register(group *gin.RouterGroup) {
	channels := group.Group("/channels")
	{
		// Public read routes
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		channels.GET("", h.ListChannels)

		// ================================
		// 2. NESTED RESOURCE ROUTES
		// ================================
		// Channel subscribers and subscription
		channels.GET("/:id/subscribers", h.GetChannelSubscribers)
		channels.GET("/:id/subscription", JWTMiddleware(h.jwt), h.GetSubscriptionStatus)
		channels.POST("/:id/subscription", JWTMiddleware(h.jwt), h.SubscribeToChannel)
		channels.DELETE("/:id/subscription", JWTMiddleware(h.jwt), h.UnsubscribeFromChannel)

		// Protected write routes
		protected := channels.Group("")
		protected.Use(JWTMiddleware(h.jwt))
		{
			protected.POST("", h.CreateChannel)
			// Media management within channel
			protected.POST("/:id/medias", h.AddMedia)
			protected.DELETE("/:id/medias/:mediaId", h.RemoveMedia)
		}

		// ================================
		// 3. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		channels.GET("/:id", h.GetChannel)
		channels.PUT("/:id", JWTMiddleware(h.jwt), h.UpdateChannel)
		channels.DELETE("/:id", JWTMiddleware(h.jwt), h.DeleteChannel)
	}
}

// ListChannels returns all channels with optional pagination.
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListChannels(c.Request.Context(), page, limit)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

// GetChannel returns a single channel by ID with its media items.
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid ID")
		return
	}

	chItem, err := h.uc.GetChannel(c.Request.Context(), id)
	if err != nil {
		Fail(c, ErrNotFound, "channel not found")
		return
	}

	OK(c, chItem)
}

// GetUserChannels returns channels for a specific user.
func (h *ChannelHandler) GetUserChannels(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid user ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListUserChannels(c.Request.Context(), userId, page, limit)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

// CreateChannel creates a new channel for the authenticated user.
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Title         string `json:"title" binding:"required,max=90"`
		Description   string `json:"description"`
		BannerLogo    string `json:"banner_logo"`
		FriendlyToken string `json:"friendly_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, ErrBadRequest, err.Error())
		return
	}

	chItem := &biz.Channel{
		Title:         input.Title,
		Description:   input.Description,
		BannerLogo:    input.BannerLogo,
		FriendlyToken: input.FriendlyToken,
		UserID:        int(claims.UserID),
	}

	created, err := h.uc.CreateChannel(c.Request.Context(), chItem)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: created})
}

// UpdateChannel updates a channel. Only the owner can update.
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid ID")
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BannerLogo  string `json:"banner_logo"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, ErrBadRequest, err.Error())
		return
	}

	chItem := &biz.Channel{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		BannerLogo:  input.BannerLogo,
	}

	updated, err := h.uc.UpdateChannel(
		c.Request.Context(),
		chItem,
		int(claims.UserID),
		claims.IsStaff,
	)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, updated)
}

// DeleteChannel deletes a channel. Only the owner or admin can delete.
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid ID")
		return
	}

	err = h.uc.DeleteChannel(c.Request.Context(), id, int(claims.UserID), claims.IsStaff)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{"message": "deleted"})
}

// AddMedia adds a media item to a channel.
func (h *ChannelHandler) AddMedia(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid channel ID")
		return
	}

	var input struct {
		MediaID int `json:"media_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, ErrBadRequest, err.Error())
		return
	}

	err = h.uc.AddMediaToChannel(
		c.Request.Context(),
		id,
		input.MediaID,
		int(claims.UserID),
		claims.IsStaff,
	)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"message":    "media added to channel",
		"channel_id": id,
		"media_id":   input.MediaID,
	})
}

// RemoveMedia removes a media item from a channel (sets channel_id to null).
func (h *ChannelHandler) RemoveMedia(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid channel ID")
		return
	}
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid media ID")
		return
	}

	err = h.uc.RemoveMediaFromChannel(
		c.Request.Context(),
		id,
		mediaId,
		int(claims.UserID),
		claims.IsStaff,
	)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{
		"message":    "media removed from channel",
		"channel_id": id,
		"media_id":   mediaId,
	})
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (h *ChannelHandler) GetChannelSubscribers(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid channel ID")
		return
	}

	// Check if count parameter is present
	if c.Query("count") == "true" {
		// Return only count
		// TODO: Implement GetSubscriberCount
		OK(c, gin.H{"count": 0})
		return
	}

	// Return subscribers list
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: Implement GetSubscribers
	OK(c, gin.H{
		"items":     []interface{}{},
		"total":     0,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetSubscriptionStatus returns the subscription status for the current user.
func (h *ChannelHandler) GetSubscriptionStatus(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid channel ID")
		return
	}

	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	// TODO: Implement IsSubscribed
	OK(c, gin.H{"is_subscribed": false})
}

// SubscribeToChannel subscribes the current user to a channel.
func (h *ChannelHandler) SubscribeToChannel(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid channel ID")
		return
	}

	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	// TODO: Implement Subscribe
	OK(c, gin.H{"success": true, "message": "Subscribed to channel"})
}

// UnsubscribeFromChannel unsubscribes the current user from a channel.
func (h *ChannelHandler) UnsubscribeFromChannel(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid channel ID")
		return
	}

	_, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}

	// TODO: Implement Unsubscribe
	OK(c, gin.H{"success": true, "message": "Unsubscribed from channel"})
}
