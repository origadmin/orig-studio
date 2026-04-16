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
	"origadmin/application/origcms/internal/handler"
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

func (h *ChannelHandler) Register(r handler.Router) {
	channels := r.Group("/channels")
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
		channels.GET("/:id/subscription", WithJWT(h.jwt, h.GetSubscriptionStatus))
		channels.POST("/:id/subscription", WithJWT(h.jwt, h.SubscribeToChannel))
		channels.DELETE("/:id/subscription", WithJWT(h.jwt, h.UnsubscribeFromChannel))

		// Protected write routes
		// Note: We can't use Use() directly with the Router interface
		// We'll need to apply middleware to each route individually
		{
			channels.POST("", WithJWT(h.jwt, h.CreateChannel))
			// Media management within channel
			channels.POST("/:id/medias", WithJWT(h.jwt, h.AddMedia))
			channels.DELETE("/:id/medias/:mediaId", WithJWT(h.jwt, h.RemoveMedia))
			// Invitation management
			channels.POST("/:id/invitations", WithJWT(h.jwt, h.InviteUserToChannel))
			channels.POST("/invitations/:id/accept", WithJWT(h.jwt, h.AcceptChannelInvitation))
			channels.POST("/invitations/:id/reject", WithJWT(h.jwt, h.RejectChannelInvitation))
			channels.GET("/invitations", WithJWT(h.jwt, h.GetChannelInvitations))
		}

		// ================================
		// 3. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		channels.GET("/:id", h.GetChannel)
		channels.PUT("/:id", WithJWT(h.jwt, h.UpdateChannel))
		channels.DELETE("/:id", WithJWT(h.jwt, h.DeleteChannel))
	}
}

// ListChannels returns all channels with optional pagination.
func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	items, total, err := h.uc.ListChannels(r.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "ok",
		"data": gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": limit,
		},
	})
}

// GetChannel returns a single channel by ID with its media items.
func (h *ChannelHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid ID"})
		return
	}

	chItem, err := h.uc.GetChannel(r.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": chItem})
}

// GetUserChannels returns channels for a specific user.
func (h *ChannelHandler) GetUserChannels(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid user ID"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 100
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	items, total, err := h.uc.ListUserChannels(r.Context(), userId, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "ok",
		"data": gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": limit,
		},
	})
}

// CreateChannel creates a new channel for the authenticated user.
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Title         string `json:"title" binding:"required,max=90"`
		Description   string `json:"description"`
		BannerLogo    string `json:"banner_logo"`
		FriendlyToken string `json:"friendly_token"`
		IsPublic      bool   `json:"is_public"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
		return
	}

	chItem := &biz.Channel{
		Title:         input.Title,
		Description:   input.Description,
		BannerLogo:    input.BannerLogo,
		FriendlyToken: input.FriendlyToken,
		IsPublic:      input.IsPublic,
		UserID:        claims.UserID,
	}

	created, err := h.uc.CreateChannel(r.Context(), chItem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": created})
}

// UpdateChannel updates a channel. Only the owner can update.
func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid ID"})
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BannerLogo  string `json:"banner_logo"`
		IsPublic    *bool  `json:"is_public"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
		return
	}

	// Get existing channel to preserve other fields
	existingChannel, err := h.uc.GetChannel(r.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	chItem := &biz.Channel{
		ID:            id,
		Title:         input.Title,
		Description:   input.Description,
		BannerLogo:    input.BannerLogo,
		IsPublic:      existingChannel.IsPublic,
		UserID:        existingChannel.UserID,
		CreatedAt:     existingChannel.CreatedAt,
	}

	// Update IsPublic if provided
	if input.IsPublic != nil {
		chItem.IsPublic = *input.IsPublic
	}

	updated, err := h.uc.UpdateChannel(
		r.Context(),
		chItem,
		claims.UserID,
		claims.IsStaff,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": updated})
}

// DeleteChannel deletes a channel. Only the owner or admin can delete.
func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid ID"})
		return
	}

	err := h.uc.DeleteChannel(r.Context(), id, claims.UserID, claims.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"message": "deleted"}})
}

// AddMedia adds a media item to a channel.
func (h *ChannelHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	var input struct {
		MediaID string `json:"media_id" binding:"required"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
		return
	}

	err := h.uc.AddMediaToChannel(
		r.Context(),
		id,
		input.MediaID,
		claims.UserID,
		claims.IsStaff,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "ok",
		"data": gin.H{
			"message":    "media added to channel",
			"channel_id": id,
			"media_id":   input.MediaID,
		},
	})
}

// RemoveMedia removes a media item from a channel (sets channel_id to null).
func (h *ChannelHandler) RemoveMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}
	mediaId := c.Param("mediaId")
	if mediaId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid media ID"})
		return
	}

	err := h.uc.RemoveMediaFromChannel(
		r.Context(),
		id,
		mediaId,
		claims.UserID,
		claims.IsStaff,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "ok",
		"data": gin.H{
			"message":    "media removed from channel",
			"channel_id": id,
			"media_id":   mediaId,
		},
	})
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (h *ChannelHandler) GetChannelSubscribers(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	// Check if count parameter is present
	if c.Query("count") == "true" {
		// Return only count
		count, err := h.uc.GetChannelSubscriberCount(r.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"count": count}})
		return
	}

	// Return subscribers list
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	subscribers, total, err := h.uc.GetChannelSubscribers(r.Context(), id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "ok",
		"data": gin.H{
			"items":     subscribers,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetSubscriptionStatus returns the subscription status for the current user.
func (h *ChannelHandler) GetSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	isSubscribed, err := h.uc.IsSubscribedToChannel(r.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"is_subscribed": isSubscribed}})
}

// SubscribeToChannel subscribes the current user to a channel.
func (h *ChannelHandler) SubscribeToChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.SubscribeToChannel(r.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "Subscribed to channel"}})
}

// UnsubscribeFromChannel unsubscribes the current user from a channel.
func (h *ChannelHandler) UnsubscribeFromChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.UnsubscribeFromChannel(r.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "Unsubscribed from channel"}})
}

// InviteUserToChannel invites a user to join a channel.
func (h *ChannelHandler) InviteUserToChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
		return
	}

	err := h.uc.InviteUserToChannel(r.Context(), id, input.UserID, claims.UserID, claims.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "User invited to channel"}})
}

// AcceptChannelInvitation accepts a channel invitation.
func (h *ChannelHandler) AcceptChannelInvitation(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.AcceptChannelInvitation(r.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "Invitation accepted"}})
}

// RejectChannelInvitation rejects a channel invitation.
func (h *ChannelHandler) RejectChannelInvitation(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.RejectChannelInvitation(r.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "Invitation rejected"}})
}

// GetChannelInvitations returns the current user's channel invitations.
func (h *ChannelHandler) GetChannelInvitations(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	invitations, err := h.uc.GetChannelInvitations(r.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"invitations": invitations}})
}
