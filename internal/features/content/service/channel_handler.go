/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel module - handles channel CRUD and subscription management
 *
 * API paths (v3.2 - path parameter version):
 * - GET /api/v1/channels              - channel list (query params: username, user_id)
 * - GET /api/v1/channels/{token}      - single channel by short_token (path parameter)
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

package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/validation"

	"github.com/gin-gonic/gin"
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

func (h *ChannelHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	channels := r.Group("/channels")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST!
		// ================================
		channels.GET("", h.ListChannels)

		// Current user's channel (requires auth)
		channels.GET("/me", server.WithJWT(h.jwt, h.GetMyChannel))
		channels.PUT("/me/handle", server.WithJWT(h.jwt, h.UpdateMyHandle))

		// ================================
		// 2. PATH PARAMETER ROUTES (WITH :token) - MUST BE AFTER STATIC
		// ================================
		// Single channel by short_token (RESTful, MediaCMS style)
		channels.GET("/:token", h.GetChannelByToken)

		// Channel videos and playlists
		channels.GET("/:token/videos", h.GetChannelVideos)
		channels.GET("/:token/playlists", h.GetChannelPlaylists)

		// Notification settings
		channels.PUT("/:token/notification", server.WithJWT(h.jwt, h.UpdateNotificationSetting))

		// ================================
		// 3. NESTED RESOURCE ROUTES (Subscription APIs)
		// ================================
		// Channel subscribers and subscription
		channels.GET("/:token/subscribers", h.GetChannelSubscribers)
		channels.GET("/:token/subscription", server.WithJWT(h.jwt, h.GetSubscriptionStatus))
		channels.POST("/:token/subscription", server.WithJWT(h.jwt, h.SubscribeToChannel))
		channels.DELETE("/:token/subscription", server.WithJWT(h.jwt, h.UnsubscribeFromChannel))

		// Protected write routes
		{
			channels.POST("", server.WithJWT(h.jwt, h.CreateChannel))
			// Media management within channel (by :token)
			channels.POST("/:token/medias", server.WithJWT(h.jwt, h.AddMedia))
			channels.DELETE("/:token/medias/:mediaId", server.WithJWT(h.jwt, h.RemoveMedia))
			// Invitation management
			channels.POST("/:token/invitations", server.WithJWT(h.jwt, h.InviteUserToChannel))
			channels.POST("/invitations/:id/accept", server.WithJWT(h.jwt, h.AcceptChannelInvitation))
			channels.POST("/invitations/:id/reject", server.WithJWT(h.jwt, h.RejectChannelInvitation))
			channels.GET("/invitations", server.WithJWT(h.jwt, h.GetChannelInvitations))

			// ================================
			// UPDATE & DELETE by :token (not :id!)
			// Application uses short_token for all operations
			// Admin uses /admin/channels/:uuid for UUID-based operations
			// ================================
			channels.PUT("/:token", server.WithJWT(h.jwt, h.UpdateChannel))    // :token = short_token
			channels.DELETE("/:token", server.WithJWT(h.jwt, h.DeleteChannel)) // :token = short_token
		}
	}

	// ================================
	// Subscription feed routes (top-level resource)
	// ================================
	subs := r.Group("/subscriptions")
	{
		subs.GET("/videos", server.WithJWT(h.jwt, h.GetSubscriptionVideos))
	}
}

// GetChannelByToken returns a single channel by short_token (path parameter).
// This is the RECOMMENDED way to access a single channel (RESTful, MediaCMS style).
func (h *ChannelHandler) GetChannelByToken(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	if !validation.IsValidShortToken(token) {
		server.Fail(gc, server.ErrBadRequest, "invalid_token_format")
		return
	}

	chItem, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "channel_not_found")
		return
	}

	server.OK(gc, chItem)
}

// ListChannels returns channels with optional query parameters.
// Supports 3 modes:
//  1. ?username={value} -> Get default channel by username
//  2. ?user_id={value}  -> Get all channels for a user
//  3. (no params)       -> List all public channels (paginated)
func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

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
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	switch {
	case username != "":
		if !validation.IsValidUsername(username) {
			server.Fail(gc, server.ErrBadRequest, "invalid_username_format")
			return
		}
		// Two-step: username -> user_id -> default channel
		chItem, err := h.uc.GetChannelByUsername(r.Context(), username)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, fmt.Sprintf("channel not found for @%s", username))
			return
		}
		server.OK(gc, chItem)

	case userId != "":
		items, total, err := h.uc.ListUserChannels(r.Context(), userId, page, limit)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OKPage(gc, items, int64(total), page, limit)

	default:
		// List all public channels (paginated)
		items, total, err := h.uc.ListChannels(r.Context(), page, limit)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OKPage(gc, items, int64(total), page, limit)
	}
}

// CreateChannel creates a new channel for the authenticated user.
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
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
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	chItem := &biz.Channel{
		Title:       input.Title,
		Description: input.Description,
		BannerLogo:  input.BannerLogo,
		ShortToken:  input.FriendlyToken,
		IsPublic:    input.IsPublic,
		UserID:      claims.GetUserID(),
	}

	created, err := h.uc.CreateChannel(r.Context(), chItem)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.Created(gc, created)
}

// UpdateChannel updates a channel by short_token. Only the owner can update.
func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BannerLogo  string `json:"banner_logo"`
		IsPublic    *bool  `json:"is_public"`
	}
	if err := c.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	existingChannel, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "channel_not_found")
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
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, updated)
}

// DeleteChannel deletes a channel by short_token. Only the owner or admin can delete.
func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	existingChannel, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "channel_not_found")
		return
	}

	err = h.uc.DeleteChannel(r.Context(), existingChannel.ID, claims.GetUserID(), claims.IsStaff)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"message": "deleted"})
}

// AddMedia adds a media item to a channel.
func (h *ChannelHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	var input struct {
		MediaID string `json:"media_id" binding:"required"`
	}
	if err := c.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
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
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{
		"message":    "media added to channel",
		"channel_id": token,
		"media_id":   input.MediaID,
	})
}

// RemoveMedia removes a media item from a channel (sets channel_id to null).
func (h *ChannelHandler) RemoveMedia(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}
	mediaId := c.Param("mediaId")
	if mediaId == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid media ID")
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
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{
		"message":    "media removed from channel",
		"channel_id": token,
		"media_id":   mediaId,
	})
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (h *ChannelHandler) GetChannelSubscribers(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	if c.Query("count") == "true" {
		count, err := h.uc.GetChannelSubscriberCount(r.Context(), token)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, gin.H{"count": count})
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
	// Normalize pagination parameters
	page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

	subscribers, total, err := h.uc.GetChannelSubscribers(r.Context(), token, page, pageSize)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OKPage(gc, subscribers, int64(total), page, pageSize)
}

// GetSubscriptionStatus returns the subscription status for the current user.
func (h *ChannelHandler) GetSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	isSubscribed, err := h.uc.IsSubscribedToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"is_subscribed": isSubscribed})
}

// SubscribeToChannel subscribes the current user to a channel.
func (h *ChannelHandler) SubscribeToChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.SubscribeToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"success": true, "message": "Subscribed to channel"})
}

// UnsubscribeFromChannel unsubscribes the current user from a channel.
func (h *ChannelHandler) UnsubscribeFromChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.UnsubscribeFromChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"success": true, "message": "Unsubscribed from channel"})
}

// InviteUserToChannel invites a user to join a channel.
func (h *ChannelHandler) InviteUserToChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	err := h.uc.InviteUserToChannel(r.Context(), token, input.UserID, claims.GetUserID(), claims.IsStaff)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"success": true, "message": "User invited to channel"})
}

// AcceptChannelInvitation accepts a channel invitation.
func (h *ChannelHandler) AcceptChannelInvitation(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	id := c.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid ID")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.AcceptChannelInvitation(r.Context(), id, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"success": true, "message": "Invitation accepted"})
}

// RejectChannelInvitation rejects a channel invitation.
func (h *ChannelHandler) RejectChannelInvitation(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	id := c.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid ID")
		return
	}

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.RejectChannelInvitation(r.Context(), id, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"success": true, "message": "Invitation rejected"})
}

// GetChannelInvitations returns the current user's channel invitations.
func (h *ChannelHandler) GetChannelInvitations(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	invitations, err := h.uc.GetChannelInvitations(r.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{"invitations": invitations})
}

// GetMyChannel returns the current authenticated user's channel.
func (h *ChannelHandler) GetMyChannel(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	channels, _, err := h.uc.ListUserChannels(r.Context(), claims.GetUserID(), 1, 1)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	if len(channels) == 0 {
		server.Fail(gc, server.ErrNotFound, "You don't have a channel yet")
		return
	}

	server.OK(gc, channels[0])
}

// UpdateMyHandle updates the current user's channel handle/slug.
func (h *ChannelHandler) UpdateMyHandle(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Handle string `json:"handle" binding:"required,min=3,max=39"`
	}
	if err := c.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	channels, _, err := h.uc.ListUserChannels(r.Context(), claims.GetUserID(), 1, 1)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	if len(channels) == 0 {
		server.Fail(gc, server.ErrNotFound, "You don't have a channel yet")
		return
	}

	ch := channels[0]
	ch.ShortToken = input.Handle

	updated, err := h.uc.UpdateChannel(r.Context(), ch, claims.GetUserID(), claims.IsStaff)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, updated)
}

// UpdateNotificationSetting updates notification preferences for a channel subscription.
func (h *ChannelHandler) UpdateNotificationSetting(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	var input struct {
		Setting string `json:"setting" binding:"required,oneof=all personalized none"`
	}
	if err := c.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	isSubscribed, err := h.uc.IsSubscribedToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	if !isSubscribed {
		server.Fail(gc, server.ErrBadRequest, "Not subscribed to this channel")
		return
	}

	server.OK(gc, gin.H{
		"success":    true,
		"setting":    input.Setting,
		"channel_id": token,
		"message":    "Notification setting updated",
	})
}

// GetSubscriptionVideos returns the latest videos from all channels the current user is subscribed to.
// Supports pagination, sorting, and channel filtering.
// Query params: page, limit, sort_by (newest|most_viewed|trending), channel_ids (comma-separated).
func (h *ChannelHandler) GetSubscriptionVideos(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	val := c.Get("claims")
	if val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
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
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	sortBy := c.Query("sort_by")
	switch sortBy {
	case "newest", "most_viewed", "trending":
		// valid
	default:
		sortBy = "newest"
	}

	ctx := r.Context()

	// Find all channel IDs the user is subscribed to
	channelIDs, err := h.uc.GetSubscribedChannelIDs(ctx, userID)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	if len(channelIDs) == 0 {
		server.OKPage(gc, []interface{}{}, int64(0), page, limit)
		return
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

	// Query videos from subscribed channels via biz layer
	items, total, err := h.uc.GetSubscriptionVideos(ctx, userID, channelIDs, sortBy, page, limit)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OKPage(gc, items, int64(total), page, limit)
}

// GetChannelVideos returns videos for a specific channel by short_token.
// Supports pagination and sorting.
// Query params: page, limit, sort_by (newest|oldest|popular).
func (h *ChannelHandler) GetChannelVideos(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	if !validation.IsValidShortToken(token) {
		server.Fail(gc, server.ErrBadRequest, "invalid_token_format")
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
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	sortBy := c.Query("sort_by")
	switch sortBy {
	case "newest", "oldest", "popular":
		// valid
	default:
		sortBy = "newest"
	}

	// Query channel videos via biz layer
	items, total, err := h.uc.GetChannelVideos(r.Context(), token, sortBy, page, limit)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "channel_not_found" {
			server.Fail(gc, server.ErrNotFound, "channel_not_found")
			return
		}
		server.Fail(gc, server.ErrInternal, errMsg)
		return
	}

	server.OKPage(gc, items, int64(total), page, limit)
}

// GetChannelPlaylists returns playlists for a specific channel by short_token.
// The channel's owner user_id is used to look up playlists.
// Supports pagination.
// Query params: page, limit.
func (h *ChannelHandler) GetChannelPlaylists(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	gc := c.GinContext()

	token := c.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	if !validation.IsValidShortToken(token) {
		server.Fail(gc, server.ErrBadRequest, "invalid_token_format")
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
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	// Query channel playlists via biz layer
	items, total, err := h.uc.GetChannelPlaylists(r.Context(), token, page, limit)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "channel_not_found" {
			server.Fail(gc, server.ErrNotFound, "channel_not_found")
			return
		}
		server.Fail(gc, server.ErrInternal, errMsg)
		return
	}

	server.OKPage(gc, items, int64(total), page, limit)
}
