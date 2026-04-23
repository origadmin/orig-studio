/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel module - handles channel CRUD and subscription management
 *
 * API paths (v3.2 - 路径参数版):
 * - GET /api/v1/channels              - channel list (query params: username, user_id)
 * - GET /api/v1/channels/{token}      - single channel by short_token (路径参数)
 * - GET /api/v1/channels/:token/subscribers - channel subscribers
 * - GET /api/v1/channels/:token/subscription - subscription status/operations
 * - PUT /api/v1/channels/:token/notification - notification settings
 * - POST /api/v1/channels              - create channel
 * - PUT /api/v1/channels/:id           - update channel (UUID)
 * - DELETE /api/v1/channels/:id        - delete channel (UUID)
 */

package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-content/biz"
	"origadmin/application/origcms/internal/validation"
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
		// Single channel by short_token (RESTful, MediaCMS风格) ⭐
		channels.GET("/:token", h.GetChannelByToken)

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
			channels.PUT("/:token", WithJWT(h.jwt, h.UpdateChannel))     // :token = short_token
			channels.DELETE("/:token", WithJWT(h.jwt, h.DeleteChannel))  // :token = short_token
		}
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
//   1. ?username={value} → Get default channel by username (两步方案 for @username)
//   2. ?user_id={value}  → Get all channels for a user
//   3. (no params)       → List all public channels (分页)
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
			"code": 0,
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
		ID:            existingChannel.ID,
		Title:         input.Title,
		Description:   input.Description,
		BannerLogo:    input.BannerLogo,
		IsPublic:      existingChannel.IsPublic,
		UserID:        existingChannel.UserID,
		CreatedAt:     existingChannel.CreatedAt,
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
		"code": 0,
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
		"code": 0,
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
			"success":   true,
			"setting":  input.Setting,
			"channel_id": token,
			"message":   "Notification setting updated",
		},
	})
}
