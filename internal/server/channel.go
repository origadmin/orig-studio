/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel module - handles channel CRUD and subscription management
 *
 * API paths (v3.2 - 路径参数版):
 * - GET /api/v1/channels              - channel list (query params: username, user_id)
 * - GET /api/v1/channels/{token}      - single channel by short_token (路径参数)
 * - GET /api/v1/channels/:token/videos - channel videos
 * - GET /api/v1/channels/:token/playlists - channel playlists
 * - GET /api/v1/channels/:token/subscribers - channel subscribers
 * - GET /api/v1/channels/:token/subscription - subscription status/operations
 * - PUT /api/v1/channels/:token/notification - notification settings
 * - GET /api/v1/subscriptions/videos  - subscribed channels' videos
 * - POST /api/v1/channels              - create channel
 * - PUT /api/v1/channels/:id           - update channel (UUID)
 * - DELETE /api/v1/channels/:id        - delete channel (UUID)
 */

package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/channel"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/playlist"
	"origadmin/application/origcms/internal/data/entity/subscription"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-content/biz"
	"origadmin/application/origcms/internal/validation"
)

// ChannelHandler handles /api/v1/channels routes.
type ChannelHandler struct {
	uc           *biz.PlaylistChannelUseCase
	jwt          *auth.Manager
	entityClient *entity.Client
}

// NewChannelHandler creates a new ChannelHandler.
func NewChannelHandler(uc *biz.PlaylistChannelUseCase, jwt *auth.Manager, entityClient *entity.Client) *ChannelHandler {
	return &ChannelHandler{uc: uc, jwt: jwt, entityClient: entityClient}
}

func (h *ChannelHandler) Register(r handler.Router) {
	channels := r.Group("/channels")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST!
		// ================================
		channels.GET("", h.ListChannels)

		// Current user's channel (requires auth)
		channels.GET("/me", WithJWT(h.jwt, h.GetMyChannel))
		channels.PUT("/me/handle", WithJWT(h.jwt, h.UpdateMyHandle))

		// ================================
		// 2. PATH PARAMETER ROUTES (WITH :token) - MUST BE AFTER STATIC
		// ================================
		// Single channel by short_token (RESTful, MediaCMS风格)
		channels.GET("/:token", h.GetChannelByToken)

		// Channel videos and playlists
		channels.GET("/:token/videos", h.GetChannelVideos)
		channels.GET("/:token/playlists", h.GetChannelPlaylists)

		// Notification settings
		channels.PUT("/:token/notification", WithJWT(h.jwt, h.UpdateNotificationSetting))

		// ================================
		// 3. NESTED RESOURCE ROUTES (Subscription APIs)
		// ================================
		// Channel subscribers and subscription
		channels.GET("/:token/subscribers", h.GetChannelSubscribers)
		channels.GET("/:token/subscription", WithJWT(h.jwt, h.GetSubscriptionStatus))
		channels.POST("/:token/subscription", WithJWT(h.jwt, h.SubscribeToChannel))
		channels.DELETE("/:token/subscription", WithJWT(h.jwt, h.UnsubscribeFromChannel))

		// Protected write routes
		{
			channels.POST("", WithJWT(h.jwt, h.CreateChannel))
			// Media management within channel (by :token)
			channels.POST("/:token/medias", WithJWT(h.jwt, h.AddMedia))
			channels.DELETE("/:token/medias/:mediaId", WithJWT(h.jwt, h.RemoveMedia))
			// Invitation management
			channels.POST("/:token/invitations", WithJWT(h.jwt, h.InviteUserToChannel))
			channels.POST("/invitations/:id/accept", WithJWT(h.jwt, h.AcceptChannelInvitation))
			channels.POST("/invitations/:id/reject", WithJWT(h.jwt, h.RejectChannelInvitation))
			channels.GET("/invitations", WithJWT(h.jwt, h.GetChannelInvitations))

			// ================================
			// UPDATE & DELETE by :token (not :id!)
			// Application uses short_token for all operations
			// Admin uses /admin/channels/:uuid for UUID-based operations
			// ================================
			channels.PUT("/:token", WithJWT(h.jwt, h.UpdateChannel))    // :token = short_token
			channels.DELETE("/:token", WithJWT(h.jwt, h.DeleteChannel)) // :token = short_token
		}
	}

	// ================================
	// Subscription feed routes (top-level resource)
	// ================================
	subs := r.Group("/subscriptions")
	{
		subs.GET("/videos", WithJWT(h.jwt, h.GetSubscriptionVideos))
	}
}

// GetChannelByToken returns a single channel by short_token (路径参数方式).
// This is the RECOMMENDED way to access a single channel (RESTful, MediaCMS风格).
func (h *ChannelHandler) GetChannelByToken(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "token is required"})
		return
	}

	if !validation.IsValidShortToken(token) {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid_token_format"})
		return
	}

	chItem, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": chItem})
}

// ListChannels returns channels with optional query parameters (查询参数方式).
// Supports 3 modes:
//  1. ?username={value} → Get default channel by username (两步方案 for @username)
//  2. ?user_id={value}  → Get all channels for a user
//  3. (no params)       → List all public channels (分页)
func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	username := c.Query("username")
	userId := c.Query("user_id")

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	switch {
	case username != "":
		if !validation.IsValidUsername(username) {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid_username_format"})
			return
		}
		// 两步方案: username → user_id → default channel
		chItem, err := h.uc.GetChannelByUsername(r.Context(), username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    ErrNotFound,
				"message": fmt.Sprintf("channel not found for @%s", username),
				"hint":    "The user may not have a channel yet",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": chItem})

	case userId != "":
		items, total, err := h.uc.ListUserChannels(r.Context(), userId, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": limit,
			},
		})

	default:
		// List all public channels (分页)
		items, total, err := h.uc.ListChannels(r.Context(), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": limit,
			},
		})
	}
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
		UserID:        claims.GetUserID(),
	}

	created, err := h.uc.CreateChannel(r.Context(), chItem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": created})
}

// UpdateChannel updates a channel by short_token. Only the owner can update.
func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "token is required"})
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

	existingChannel, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel_not_found"})
		return
	}

	chItem := &biz.Channel{
		ID:          existingChannel.ID,
		Title:       input.Title,
		Description: input.Description,
		BannerLogo:  input.BannerLogo,
		IsPublic:    existingChannel.IsPublic,
		UserID:      existingChannel.UserID,
		CreatedAt:   existingChannel.CreatedAt,
	}

	if input.IsPublic != nil {
		chItem.IsPublic = *input.IsPublic
	}

	updated, err := h.uc.UpdateChannel(
		r.Context(),
		chItem,
		claims.GetUserID(),
		claims.IsStaff,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": updated})
}

// DeleteChannel deletes a channel by short_token. Only the owner or admin can delete.
func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "token is required"})
		return
	}

	existingChannel, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel_not_found"})
		return
	}

	err = h.uc.DeleteChannel(r.Context(), existingChannel.ID, claims.GetUserID(), claims.IsStaff)
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

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
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
		token,
		input.MediaID,
		claims.GetUserID(),
		claims.IsStaff,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"message":    "media added to channel",
			"channel_id": token,
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

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
		return
	}
	mediaId := c.Param("mediaId")
	if mediaId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid media ID"})
		return
	}

	err := h.uc.RemoveMediaFromChannel(
		r.Context(),
		token,
		mediaId,
		claims.GetUserID(),
		claims.IsStaff,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"message":    "media removed from channel",
			"channel_id": token,
			"media_id":   mediaId,
		},
	})
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (h *ChannelHandler) GetChannelSubscribers(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
		return
	}

	if c.Query("count") == "true" {
		count, err := h.uc.GetChannelSubscriberCount(r.Context(), token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"count": count}})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	subscribers, total, err := h.uc.GetChannelSubscribers(r.Context(), token, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
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
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	isSubscribed, err := h.uc.IsSubscribedToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"is_subscribed": isSubscribed}})
}

// SubscribeToChannel subscribes the current user to a channel.
func (h *ChannelHandler) SubscribeToChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.SubscribeToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "Subscribed to channel"}})
}

// UnsubscribeFromChannel unsubscribes the current user from a channel.
func (h *ChannelHandler) UnsubscribeFromChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.UnsubscribeFromChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"success": true, "message": "Unsubscribed from channel"}})
}

// InviteUserToChannel invites a user to join a channel.
func (h *ChannelHandler) InviteUserToChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
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

	err := h.uc.InviteUserToChannel(r.Context(), token, input.UserID, claims.GetUserID(), claims.IsStaff)
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
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.AcceptChannelInvitation(r.Context(), id, claims.GetUserID())
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
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid ID"})
		return
	}

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.RejectChannelInvitation(r.Context(), id, claims.GetUserID())
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

	invitations, err := h.uc.GetChannelInvitations(r.Context(), claims.GetUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"invitations": invitations}})
}

// GetMyChannel returns the current authenticated user's channel.
func (h *ChannelHandler) GetMyChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	channels, _, err := h.uc.ListUserChannels(r.Context(), claims.GetUserID(), 1, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	if len(channels) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "You don't have a channel yet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": channels[0]})
}

// UpdateMyHandle updates the current user's channel handle/slug.
func (h *ChannelHandler) UpdateMyHandle(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Handle string `json:"handle" binding:"required,min=3,max=39"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
		return
	}

	channels, _, err := h.uc.ListUserChannels(r.Context(), claims.GetUserID(), 1, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	if len(channels) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "You don't have a channel yet"})
		return
	}

	ch := channels[0]
	ch.FriendlyToken = input.Handle

	updated, err := h.uc.UpdateChannel(r.Context(), ch, claims.GetUserID(), claims.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": updated})
}

// UpdateNotificationSetting updates notification preferences for a channel subscription.
func (h *ChannelHandler) UpdateNotificationSetting(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Invalid channel token"})
		return
	}

	var input struct {
		Setting string `json:"setting" binding:"required,oneof=all personalized none"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
		return
	}

	isSubscribed, err := h.uc.IsSubscribedToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	if !isSubscribed {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "Not subscribed to this channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"success":    true,
			"setting":    input.Setting,
			"channel_id": token,
			"message":    "Notification setting updated",
		},
	})
}

// GetSubscriptionVideos returns the latest videos from all channels the current user is subscribed to.
// Supports pagination, sorting, and channel filtering.
// Query params: page, limit, sort_by (newest|most_viewed|trending), channel_ids (comma-separated).
func (h *ChannelHandler) GetSubscriptionVideos(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)

	val := c.Get("claims")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": ErrUnauthorized, "message": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)
	userID := claims.GetUserID()

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	sortBy := c.Query("sort_by")
	switch sortBy {
	case "newest", "most_viewed", "trending":
		// valid
	default:
		sortBy = "newest"
	}

	ctx := r.Context()

	// Find all channel IDs the user is subscribed to
	subscriptions, err := h.entityClient.Subscription.Query().
		Where(subscription.SubscriberID(userID)).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	if len(subscriptions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":     []interface{}{},
				"total":     0,
				"page":      page,
				"page_size": limit,
			},
		})
		return
	}

	channelIDs := make([]string, 0, len(subscriptions))
	for _, sub := range subscriptions {
		channelIDs = append(channelIDs, sub.ChannelID)
	}

	// Apply channel_ids filter if provided
	if channelIDsParam := c.Query("channel_ids"); channelIDsParam != "" {
		filterIDs := strings.Split(channelIDsParam, ",")
		filtered := make([]string, 0, len(filterIDs))
		for _, id := range filterIDs {
			id = strings.TrimSpace(id)
			if id != "" {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) > 0 {
			channelIDs = filtered
		}
	}

	// Build media query for subscribed channels
	query := h.entityClient.Media.Query().
		Where(
			media.ChannelIDIn(channelIDs...),
			media.StateEQ("active"),
			media.PrivacyEQ(1), // public only
		)

	// Apply sorting
	switch sortBy {
	case "newest":
		query.Order(entity.Desc(media.FieldCreatedAt))
	case "most_viewed":
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreatedAt))
	case "trending":
		// Trending: prioritize recent videos with high engagement
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		query.Where(media.CreatedAtGTE(sevenDaysAgo))
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreatedAt))
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	// Apply pagination
	offset := (page - 1) * limit
	medias, err := query.
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	items := make([]map[string]interface{}, 0, len(medias))
	for _, m := range medias {
		item := map[string]interface{}{
			"id":              m.ID,
			"short_token":     m.ShortToken,
			"title":           m.Title,
			"description":     m.Description,
			"thumbnail":       m.Thumbnail,
			"duration":        m.Duration,
			"view_count":      m.ViewCount,
			"like_count":      m.LikeCount,
			"comment_count":   m.CommentCount,
			"type":            m.Type,
			"channel_id":      m.ChannelID,
			"user_id":         m.UserID,
			"encoding_status": m.EncodingStatus,
		}
		if !m.CreatedAt.IsZero() {
			item["created_at"] = m.CreatedAt.Format(time.RFC3339)
		}
		if !m.PublishedAt.IsZero() {
			item["published_at"] = m.PublishedAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": limit,
		},
	})
}

// GetChannelVideos returns videos for a specific channel by short_token.
// Supports pagination and sorting.
// Query params: page, limit, sort_by (newest|oldest|popular).
func (h *ChannelHandler) GetChannelVideos(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "token is required"})
		return
	}

	if !validation.IsValidShortToken(token) {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid_token_format"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	sortBy := c.Query("sort_by")
	switch sortBy {
	case "newest", "oldest", "popular":
		// valid
	default:
		sortBy = "newest"
	}

	ctx := r.Context()

	// Resolve short_token to channel ID
	ch, err := h.entityClient.Channel.Query().
		Where(channel.ShortToken(token)).
		Only(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel_not_found"})
		return
	}

	// Build media query for this channel
	query := h.entityClient.Media.Query().
		Where(
			media.ChannelID(ch.ID),
			media.StateEQ("active"),
		)

	// Apply sorting
	switch sortBy {
	case "newest":
		query.Order(entity.Desc(media.FieldCreatedAt))
	case "oldest":
		query.Order(entity.Asc(media.FieldCreatedAt))
	case "popular":
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreatedAt))
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	// Apply pagination
	offset := (page - 1) * limit
	medias, err := query.
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	items := make([]map[string]interface{}, 0, len(medias))
	for _, m := range medias {
		item := map[string]interface{}{
			"id":              m.ID,
			"short_token":     m.ShortToken,
			"title":           m.Title,
			"description":     m.Description,
			"thumbnail":       m.Thumbnail,
			"duration":        m.Duration,
			"view_count":      m.ViewCount,
			"like_count":      m.LikeCount,
			"comment_count":   m.CommentCount,
			"type":            m.Type,
			"channel_id":      m.ChannelID,
			"user_id":         m.UserID,
			"encoding_status": m.EncodingStatus,
		}
		if !m.CreatedAt.IsZero() {
			item["created_at"] = m.CreatedAt.Format(time.RFC3339)
		}
		if !m.PublishedAt.IsZero() {
			item["published_at"] = m.PublishedAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": limit,
		},
	})
}

// GetChannelPlaylists returns playlists for a specific channel by short_token.
// The channel's owner user_id is used to look up playlists.
// Supports pagination.
// Query params: page, limit.
func (h *ChannelHandler) GetChannelPlaylists(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "token is required"})
		return
	}

	if !validation.IsValidShortToken(token) {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid_token_format"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	ctx := r.Context()

	// Resolve short_token to channel, then get user_id for playlist lookup
	ch, err := h.entityClient.Channel.Query().
		Where(channel.ShortToken(token)).
		Only(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "channel_not_found"})
		return
	}

	// Query playlists by user_id (channel owner)
	query := h.entityClient.Playlist.Query().
		Where(
			playlist.UserID(ch.UserID),
			playlist.PrivacyEQ(1), // public playlists only
		).
		Order(entity.Desc(playlist.FieldAddDate))

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	// Apply pagination
	offset := (page - 1) * limit
	playlists, err := query.
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
		return
	}

	items := make([]map[string]interface{}, 0, len(playlists))
	for _, p := range playlists {
		item := map[string]interface{}{
			"id":          p.ID,
			"short_token": p.ShortToken,
			"title":       p.Title,
			"description": p.Description,
			"user_id":     p.UserID,
			"privacy":     p.Privacy,
		}
		if !p.AddDate.IsZero() {
			item["created_at"] = p.AddDate.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": limit,
		},
	})
}
