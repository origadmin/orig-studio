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

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/server/validation"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"

	"github.com/gin-gonic/gin"
)

// ChannelHandler handles /api/v1/channels routes.
type ChannelHandler struct {
	uc        *biz.PlaylistChannelUseCase
	jwt       *auth.Manager
	settingUC *systembiz.SettingUseCase
}

// NewChannelHandler creates a new ChannelHandler.
func NewChannelHandler(uc *biz.PlaylistChannelUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase) *ChannelHandler {
	return &ChannelHandler{uc: uc, jwt: jwt, settingUC: settingUC}
}

func (h *ChannelHandler) RegisterRoutes(r http2.Router) {
	channelsGroup := r.Group("/channels")
	channelsGroup.Use(systemservice.ModuleGuardCtx(h.settingUC, "module_videos"))

	channels := channelsGroup.Group("")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST!
		// ================================
		channels.GET("", httpToHandlerFunc(h.ListChannels))

		// Current user's channels (requires auth)
		channels.GET("/me", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.GetMyChannels)))
		channels.PUT("/me/handle", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.UpdateMyHandle)))

		// Handle validation (public)
		channels.GET("/validate-handle", httpToHandlerFunc(h.ValidateHandle))

		// ================================
		// 2. PATH PARAMETER ROUTES (WITH :token) - MUST BE AFTER STATIC
		// ================================
		// Single channel by short_token (RESTful, MediaCMS style)
		channels.GET("/:token", httpToHandlerFunc(h.GetChannelByToken))

		// Channel videos and playlists
		channels.GET("/:token/videos", httpToHandlerFunc(h.GetChannelVideos))
		channels.GET("/:token/playlists", httpToHandlerFunc(h.GetChannelPlaylists))

		// Notification settings
		channels.PUT("/:token/notification", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.UpdateNotificationSetting)))

		// ================================
		// 3. NESTED RESOURCE ROUTES (Subscription APIs)
		// ================================
		// Channel subscribers and subscription
		channels.GET("/:token/subscribers", httpToHandlerFunc(h.GetChannelSubscribers))
		channels.GET("/:token/subscription", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.GetSubscriptionStatus)))
		channels.POST("/:token/subscription", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.SubscribeToChannel)))
		channels.DELETE("/:token/subscription", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.UnsubscribeFromChannel)))

		// Protected write routes
		{
			channels.POST("", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.CreateChannel)))
			// Media management within channel (by :token)
			channels.POST("/:token/medias", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.AddMedia)))
			channels.DELETE("/:token/medias/:mediaId", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.RemoveMedia)))
			// Invitation management
			channels.POST("/:token/invitations", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.InviteUserToChannel)))
			channels.POST("/invitations/:id/accept", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.AcceptChannelInvitation)))
			channels.POST("/invitations/:id/reject", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.RejectChannelInvitation)))
			channels.GET("/invitations", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.GetChannelInvitations)))

			// ================================
			// UPDATE & DELETE by :token (not :id!)
			// Application uses short_token for all operations
			// Admin uses /admin/channels/:uuid for UUID-based operations
			// ================================
			channels.PUT("/:token", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.UpdateChannel)))    // :token = short_token
			channels.DELETE("/:token", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.DeleteChannel))) // :token = short_token
		}
	}

	// ================================
	// Handle resolution route (top-level, NOT under /channels)
	// ================================
	resolveGroup := r.Group("/resolve")
	{
		resolveGroup.GET("/@:handle", httpToHandlerFunc(h.ResolveHandle))
	}

	// ================================
	// System config routes (top-level, NOT under /channels)
	// ================================
	configGroup := r.Group("/system/config")
	{
		configGroup.GET("/channel-limits", httpToHandlerFunc(h.GetChannelLimits))
	}

	// ================================
	// Subscription feed routes (top-level, NOT under /channels)
	// ================================
	subsGroup := r.Group("/subscriptions")
	{
		subsGroup.GET("/videos", server.WithJWTCtx(h.jwt, httpToHandlerFunc(h.GetSubscriptionVideos)))
	}
}

// GetChannelByToken returns a single channel by short_token (path parameter).
// This is the RECOMMENDED way to access a single channel (RESTful, MediaCMS style).
func (h *ChannelHandler) GetChannelByToken(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	if !validation.IsValidShortToken(token) {
		server.Fail(gc, server.ErrBadRequest, "invalid_token_format: token must be 6-12 alphanumeric characters")
		return
	}

	chItem, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "channel_not_found: no channel exists with the given token")
		return
	}

	server.OK(gc, &pb.GetChannelResponse{
		Channel: bizChannelToProto(chItem),
	})
}

// ListChannels returns channels with optional query parameters.
// Supports 3 modes:
//  1. ?username={value} -> Get default channel by username
//  2. ?user_id={value}  -> Get all channels for a user
//  3. (no params)       -> List all public channels (paginated)
func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	username := gc.Query("username")
	userId := gc.Query("user_id")

	limit, _ := strconv.Atoi(gc.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	page, _ := strconv.Atoi(gc.Query("page"))
	if page == 0 {
		page = 1
	}
	// Normalize pagination parameters
	page, limit = repotypes.NormalizeHTTPPagination(page, limit)

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
		server.OK(gc, &pb.GetChannelResponse{
			Channel: bizChannelToProto(chItem),
		})

	case userId != "":
		items, total, err := h.uc.ListUserChannels(r.Context(), userId, page, limit)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		pbChannels := bizChannelsToProto(items)
		server.OK(gc, &pb.ListChannelsResponse{
			Items:     pbChannels,
			Total:     int32(total),
			Page:      int32(page),
			PageSize:  int32(limit),
		})

	default:
		// List all public channels (paginated)
		items, total, err := h.uc.ListChannels(r.Context(), page, limit)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		pbChannels := bizChannelsToProto(items)
		server.OK(gc, &pb.ListChannelsResponse{
			Items:     pbChannels,
			Total:     int32(total),
			Page:      int32(page),
			PageSize:  int32(limit),
		})
	}
}

// CreateChannel creates a new channel for the authenticated user.
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Name          string   `json:"name" binding:"required,min=3,max=150"`
		Handle        string   `json:"handle" binding:"required,min=3,max=39"`
		Description   string   `json:"description"`
		Avatar        string   `json:"avatar"`
		Banner        string   `json:"banner"`
		BannerLogo    string   `json:"banner_logo"`
		Privacy       string   `json:"privacy"`
		Tags          []string `json:"tags"`
		TagIDs        []int    `json:"tag_ids"`
		CategoryID    *int64   `json:"category_id"`
		ShortToken    string   `json:"short_token"`
	}
	if err := gc.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	// Validate handle format
	if !validation.IsValidHandle(input.Handle) {
		server.Fail(gc, server.ErrBadRequest, "invalid_handle_format: must be 3-39 chars, alphanumeric and hyphens only")
		return
	}

	// Auto-generate slug from name
	slug := generateSlug(input.Name)

	chItem := &biz.Channel{
		Name:        input.Name,
		Title:       input.Name, // Title defaults to name
		Slug:        slug,
		Handle:      input.Handle,
		Description: input.Description,
		Avatar:      input.Avatar,
		Banner:      input.Banner,
		BannerLogo:  input.BannerLogo,
		// Only set ShortToken when FriendlyToken is provided and non-empty;
		// otherwise let ent schema's DefaultFunc (idutil.GenShortID) auto-generate one.
		Privacy:    input.Privacy,
		Tags:       input.Tags,
		TagIDs:     input.TagIDs,
		CategoryID: input.CategoryID,
		Status:     "ACTIVE",
		UserID:     claims.GetUserID(),
	}
	if input.ShortToken != "" {
		chItem.ShortToken = input.ShortToken
	}

	if chItem.Privacy == "" {
		chItem.Privacy = "PUBLIC"
	}

	created, err := h.uc.CreateChannel(r.Context(), chItem)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "channel_limit_reached") {
			server.Fail(gc, server.ErrBadRequest, errMsg)
			return
		}
		if strings.Contains(errMsg, "handle_already_taken") {
			server.Fail(gc, server.ErrConflict, errMsg)
			return
		}
		if strings.Contains(errMsg, "too_many_tags") {
			server.Fail(gc, server.ErrBadRequest, errMsg)
			return
		}
		server.Fail(gc, server.ErrInternal, errMsg)
		return
	}

	server.Created(gc, &pb.CreateChannelResponse{
		Channel: bizChannelToProto(created),
	})
}

// UpdateChannel updates a channel by short_token. Only the owner can update.
func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	var input struct {
		Name        *string  `json:"name"`
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Avatar      *string  `json:"avatar"`
		Banner      *string  `json:"banner"`
		BannerLogo  *string  `json:"banner_logo"`
		Privacy     *string  `json:"privacy"`
		Status      *string  `json:"status"`
		Tags        []string `json:"tags"`
		TagIDs      []int    `json:"tag_ids"`
		CategoryID  *int64   `json:"category_id"`
		Links       []struct {
			Type     string `json:"type"`
			Platform string `json:"platform"`
			URL      string `json:"url"`
			Title    string `json:"title"`
		} `json:"links"`
	}
	if err := gc.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	existingChannel, err := h.uc.GetByShortToken(r.Context(), token)
	if err != nil {
		server.Fail(gc, server.ErrNotFound, "channel_not_found")
		return
	}

	// Apply partial updates
	chItem := &biz.Channel{
		ID:              existingChannel.ID,
		Name:            existingChannel.Name,
		Title:           existingChannel.Title,
		Slug:            existingChannel.Slug,
		Handle:          existingChannel.Handle,
		Description:     existingChannel.Description,
		Avatar:          existingChannel.Avatar,
		Banner:          existingChannel.Banner,
		BannerLogo:      existingChannel.BannerLogo,
		ShortToken:      existingChannel.ShortToken,
		Status:          existingChannel.Status,
		Privacy:         existingChannel.Privacy,
		IsVerified:      existingChannel.IsVerified,
		Tags:            existingChannel.Tags,
		TagIDs:          existingChannel.TagIDs,
		CategoryID:      existingChannel.CategoryID,
		SubscriberCount: existingChannel.SubscriberCount,
		MediaCount:      existingChannel.MediaCount,
		ArticleCount:    existingChannel.ArticleCount,
		TotalViews:      existingChannel.TotalViews,
		Links:           existingChannel.Links,
		UserID:          existingChannel.UserID,
		CreateTime:      existingChannel.CreateTime,
		UpdateTime:      existingChannel.UpdateTime,
	}

	if input.Name != nil {
		chItem.Name = *input.Name
		chItem.Slug = generateSlug(*input.Name) // Regenerate slug on name change
	}
	if input.Title != nil {
		chItem.Title = *input.Title
	}
	if input.Description != nil {
		chItem.Description = *input.Description
	}
	if input.Avatar != nil {
		chItem.Avatar = *input.Avatar
	}
	if input.Banner != nil {
		chItem.Banner = *input.Banner
	}
	if input.BannerLogo != nil {
		chItem.BannerLogo = *input.BannerLogo
	}
	if input.Privacy != nil {
		chItem.Privacy = *input.Privacy
	}
	if input.Status != nil {
		chItem.Status = *input.Status
	}
	if input.Tags != nil {
		chItem.Tags = input.Tags
	}
	if input.TagIDs != nil {
		chItem.TagIDs = input.TagIDs
	}
	if input.CategoryID != nil {
		chItem.CategoryID = input.CategoryID
	}
	if input.Links != nil {
		chItem.Links = make([]biz.ChannelLink, len(input.Links))
		for i, l := range input.Links {
			chItem.Links[i] = biz.ChannelLink{
				Type:     l.Type,
				Platform: l.Platform,
				URL:      l.URL,
				Title:    l.Title,
			}
		}
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

	server.OK(gc, &pb.UpdateChannelResponse{
		Channel: bizChannelToProto(updated),
	})
}

// DeleteChannel deletes a channel by short_token. Only the owner or admin can delete.
func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := gc.Param("token")
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

	server.OK(gc, &pb.DeleteChannelResponse{})
}

// AddMedia adds a media item to a channel.
func (h *ChannelHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	var input struct {
		MediaID string `json:"media_id" binding:"required"`
	}
	if err := gc.Bind(&input); err != nil {
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

	server.OK(gc, &pb.AddChannelMediaResponse{
		Success: true,
	})
}

// RemoveMedia removes a media item from a channel (sets channel_id to null).
func (h *ChannelHandler) RemoveMedia(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}
	mediaId := gc.Param("mediaId")
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

	server.OK(gc, &pb.RemoveChannelMediaResponse{
		Success: true,
	})
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (h *ChannelHandler) GetChannelSubscribers(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	if gc.Query("count") == "true" {
		count, err := h.uc.GetChannelSubscriberCount(r.Context(), token)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, &pb.GetChannelSubscribersResponse{
			Count: int32(count),
		})
		return
	}

	page, _ := strconv.Atoi(gc.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(gc.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}
	// Normalize pagination parameters
	page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

	subscribers, total, err := h.uc.GetChannelSubscribers(r.Context(), token, page, pageSize)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	pbSubscribers := make([]*types.User, len(subscribers))
	// subscribers from biz layer are interface{}; convert to proto User where possible
	_ = pbSubscribers // placeholder until proper conversion is available

	server.OK(gc, &pb.GetChannelSubscribersResponse{
		Subscribers: pbSubscribers,
		Total:       int32(total),
		Page:        int32(page),
		PageSize:    int32(pageSize),
	})
}

// GetSubscriptionStatus returns the subscription status for the current user.
func (h *ChannelHandler) GetSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	isSubscribed, err := h.uc.IsSubscribedToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.GetChannelSubscriptionResponse{
		IsSubscribed: isSubscribed,
	})
}

// SubscribeToChannel subscribes the current user to a channel.
func (h *ChannelHandler) SubscribeToChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.SubscribeToChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		errMsg := err.Error()
		if errMsg == "cannot_subscribe_own_channel" {
			server.Fail(gc, server.ErrBadRequest, errMsg)
			return
		}
		server.Fail(gc, server.ErrInternal, errMsg)
		return
	}

	server.OK(gc, &pb.SubscribeChannelResponse{
		Success:      true,
		IsSubscribed: true,
	})
}

// UnsubscribeFromChannel unsubscribes the current user from a channel.
func (h *ChannelHandler) UnsubscribeFromChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.UnsubscribeFromChannel(r.Context(), token, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.UnsubscribeChannelResponse{
		Success:      true,
		IsSubscribed: false,
	})
}

// InviteUserToChannel invites a user to join a channel.
func (h *ChannelHandler) InviteUserToChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := gc.Bind(&input); err != nil {
		server.Fail(gc, server.ErrBadRequest, err.Error())
		return
	}

	err := h.uc.InviteUserToChannel(r.Context(), token, input.UserID, claims.GetUserID(), claims.IsStaff)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.SubscribeChannelResponse{
		Success:      true,
		IsSubscribed: true,
	})
}

// AcceptChannelInvitation accepts a channel invitation.
func (h *ChannelHandler) AcceptChannelInvitation(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	id := gc.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid ID")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.AcceptChannelInvitation(r.Context(), id, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.SubscribeChannelResponse{
		Success:      true,
		IsSubscribed: true,
	})
}

// RejectChannelInvitation rejects a channel invitation.
func (h *ChannelHandler) RejectChannelInvitation(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	id := gc.Param("id")
	if id == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid ID")
		return
	}

	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.RejectChannelInvitation(r.Context(), id, claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.UnsubscribeChannelResponse{
		Success:      true,
		IsSubscribed: false,
	})
}

// GetChannelInvitations returns the current user's channel invitations.
func (h *ChannelHandler) GetChannelInvitations(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	invitations, err := h.uc.GetChannelInvitations(r.Context(), claims.GetUserID())
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	// TODO: Define ChannelInvitation proto type and convert properly
	server.OK(gc, &pb.ListChannelsResponse{
		Items: []*types.Channel{},
	})
	_ = invitations
}

// GetMyChannel returns the current authenticated user's channel.
func (h *ChannelHandler) GetMyChannel(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
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
		// Return 200 with null channel instead of 404;
		// 404 should mean "endpoint not found", not "user has no channel yet".
		server.OK(gc, &pb.GetChannelResponse{
			Channel: nil,
		})
		return
	}

	server.OK(gc, &pb.GetChannelResponse{
		Channel: bizChannelToProto(channels[0]),
	})
}

// UpdateMyHandle updates the current user's channel handle/slug.
func (h *ChannelHandler) UpdateMyHandle(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Handle string `json:"handle" binding:"required,min=3,max=39"`
	}
	if err := gc.Bind(&input); err != nil {
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
	ch.Handle = input.Handle

	updated, err := h.uc.UpdateChannel(r.Context(), ch, claims.GetUserID(), claims.IsStaff)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &pb.UpdateChannelResponse{
		Channel: bizChannelToProto(updated),
	})
}

// UpdateNotificationSetting updates notification preferences for a channel subscription.
func (h *ChannelHandler) UpdateNotificationSetting(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "Invalid channel token")
		return
	}

	var input struct {
		Setting string `json:"setting" binding:"required,oneof=all personalized none"`
	}
	if err := gc.Bind(&input); err != nil {
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

	server.OK(gc, &pb.SubscribeChannelResponse{
		Success:      true,
		IsSubscribed: true,
	})
}

// GetSubscriptionVideos returns the latest videos from all channels the current user is subscribed to.
// Supports pagination, sorting, and channel filtering.
// Query params: page, limit, sort_by (newest|most_viewed|trending), channel_ids (comma-separated).
func (h *ChannelHandler) GetSubscriptionVideos(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)
	userID := claims.GetUserID()

	page, _ := strconv.Atoi(gc.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(gc.Query("limit"))
	if limit < 1 {
		limit = 20
	}
	// Normalize pagination parameters
	page, limit = repotypes.NormalizeHTTPPagination(page, limit)

	sortBy := gc.Query("sort_by")
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
		server.OK(gc, &pb.GetChannelMediasResponse{
			Items:    []*types.Media{},
			Total:    0,
			Page:     int32(page),
			PageSize: int32(limit),
		})
		return
	}

	// Apply channel_ids filter if provided
	if channelIDsParam := gc.Query("channel_ids"); channelIDsParam != "" {
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

	// TODO: Convert biz items to proto Media objects properly
	pbMedias := make([]*types.Media, 0)
	_ = items

	server.OK(gc, &pb.GetChannelMediasResponse{
		Items:    pbMedias,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(limit),
	})
}

// GetChannelVideos returns videos for a specific channel by short_token.
// Supports pagination and sorting.
// Query params: page, limit, sort_by (newest|oldest|popular).
func (h *ChannelHandler) GetChannelVideos(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	if !validation.IsValidShortToken(token) {
		server.Fail(gc, server.ErrBadRequest, "invalid_token_format")
		return
	}

	page, _ := strconv.Atoi(gc.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(gc.Query("limit"))
	if limit < 1 {
		limit = 20
	}
	// Normalize pagination parameters
	page, limit = repotypes.NormalizeHTTPPagination(page, limit)

	sortBy := gc.Query("sort_by")
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

	// TODO: Convert biz items to proto Media objects properly
	pbMedias := make([]*types.Media, 0)
	_ = items

	server.OK(gc, &pb.GetChannelMediasResponse{
		Items:    pbMedias,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(limit),
	})
}

// GetChannelPlaylists returns playlists for a specific channel by short_token.
// The channel's owner user_id is used to look up playlists.
// Supports pagination.
// Query params: page, limit.
func (h *ChannelHandler) GetChannelPlaylists(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	token := gc.Param("token")
	if token == "" {
		server.Fail(gc, server.ErrBadRequest, "token is required")
		return
	}

	if !validation.IsValidShortToken(token) {
		server.Fail(gc, server.ErrBadRequest, "invalid_token_format")
		return
	}

	page, _ := strconv.Atoi(gc.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(gc.Query("limit"))
	if limit < 1 {
		limit = 20
	}
	// Normalize pagination parameters
	page, limit = repotypes.NormalizeHTTPPagination(page, limit)

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

	// TODO: Convert biz items to proto Playlist objects properly
	pbPlaylists := make([]*types.Playlist, 0)
	_ = items

	server.OK(gc, &pb.GetPlaylistsResponse{
		Items:     pbPlaylists,
		Total:     int32(total),
		Page:      int32(page),
		PageSize:  int32(limit),
	})
}

// bizChannelToProto converts a biz.Channel to a proto types.Channel.
func bizChannelToProto(ch *biz.Channel) *types.Channel {
	if ch == nil {
		return nil
	}

	privacy := types.Privacy_PRIVACY_PUBLIC
	switch ch.Privacy {
	case "PRIVATE":
		privacy = types.Privacy_PRIVACY_PRIVATE
	case "UNLISTED":
		privacy = types.Privacy_PRIVACY_UNLISTED
	case "PAID":
		privacy = types.Privacy_PRIVACY_PAID
	case "SUBSCRIBERS_ONLY":
		privacy = types.Privacy_PRIVACY_SUBSCRIBERS_ONLY
	}

	status := types.ChannelStatus_CHANNEL_STATUS_ACTIVE
	switch ch.Status {
	case "INACTIVE":
		status = types.ChannelStatus_CHANNEL_STATUS_INACTIVE
	case "SUSPENDED":
		status = types.ChannelStatus_CHANNEL_STATUS_SUSPENDED
	case "PENDING_REVIEW":
		status = types.ChannelStatus_CHANNEL_STATUS_PENDING_REVIEW
	}

	pb := &types.Channel{
		Id:              ch.ID,
		Name:            ch.Name,
		Title:           ch.Title,
		Slug:            ch.Slug,
		Handle:          ch.Handle,
		Description:     ch.Description,
		Avatar:          ch.Avatar,
		Banner:          ch.Banner,
		BannerLogo:      ch.BannerLogo,
		ShortToken:      ch.ShortToken,
		Status:          status,
		Privacy:         privacy,
		IsVerified:      ch.IsVerified,
		Tags:            ch.Tags,
		SubscriberCount: ch.SubscriberCount,
		MediaCount:      int64(ch.MediaCount),
		ArticleCount:    int32(ch.ArticleCount),
		TotalViews:      ch.TotalViews,
		UserId:          ch.UserID,
		IsOwner:         ch.IsOwner,
		IsSubscribed:    ch.IsSubscribed,
	}

	if ch.CategoryID != nil {
		pb.CategoryId = *ch.CategoryID
	}

	if ch.Links != nil {
		pb.Links = make([]*types.ChannelLink, len(ch.Links))
		for i, l := range ch.Links {
			pb.Links[i] = &types.ChannelLink{
				Type:     l.Type,
				Platform: l.Platform,
				Url:      l.URL,
				Title:    l.Title,
			}
		}
	}

	if !ch.CreateTime.IsZero() {
		pb.CreateTime = timestamppb.New(ch.CreateTime)
	}
	if !ch.UpdateTime.IsZero() {
		pb.UpdateTime = timestamppb.New(ch.UpdateTime)
	}
	return pb
}

// bizChannelsToProto converts a slice of biz.Channel to proto types.Channel.
func bizChannelsToProto(channels []*biz.Channel) []*types.Channel {
	result := make([]*types.Channel, len(channels))
	for i, ch := range channels {
		result[i] = bizChannelToProto(ch)
	}
	return result
}

// generateSlug creates a URL-safe slug from a name.
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	// Collapse multiple hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) > 150 {
		slug = slug[:150]
	}
	return slug
}

// ResolveHandle resolves a @handle to a channel or user.
// GET /api/v1/resolve/@{handle}
func (h *ChannelHandler) ResolveHandle(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	handle := gc.Param("handle")
	if handle == "" {
		server.Fail(gc, server.ErrBadRequest, "handle is required")
		return
	}

	// Strip leading @ if present
	handle = strings.TrimPrefix(handle, "@")

	result, err := h.uc.ResolveHandle(r.Context(), handle)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	resolutionType := types.HandleResolution_RESOLUTION_TYPE_NOT_FOUND
	switch result.Type {
	case "channel":
		resolutionType = types.HandleResolution_RESOLUTION_TYPE_CHANNEL
	case "user":
		resolutionType = types.HandleResolution_RESOLUTION_TYPE_USER
	}

	pbResult := &types.HandleResolution{
		Type: resolutionType,
	}

	if result.Channel != nil {
		pbResult.Channel = bizChannelToProto(result.Channel)
	}
	if result.User != nil {
		pbResult.User = &types.User{
			Id:       result.User.ID,
			Username: result.User.Username,
			Name:     result.User.Name,
			Logo:     result.User.Logo,
		}
	}

	server.OK(gc, pbResult)
}

// GetMyChannels returns all channels for the authenticated user.
// GET /api/v1/channels/me
func (h *ChannelHandler) GetMyChannels(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, exists := gc.Get("claims")
	if !exists || val == nil {
		server.Fail(gc, server.ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	page, _ := strconv.Atoi(gc.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(gc.Query("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}
	page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

	channels, total, err := h.uc.ListUserChannels(r.Context(), claims.GetUserID(), page, pageSize)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	// Mark is_owner for all channels
	for _, ch := range channels {
		ch.IsOwner = true
	}

	server.OK(gc, &pb.ListChannelsResponse{
		Items:    bizChannelsToProto(channels),
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
}

// GetChannelLimits returns channel creation limits for the current user.
// GET /api/v1/system/config/channel-limits
func (h *ChannelHandler) GetChannelLimits(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	val, _ := gc.Get("claims")
	isAdmin := false
	var userID string
	if val != nil {
		claims := val.(*auth.Claims)
		userID = claims.GetUserID()
		isAdmin = claims.IsStaff
	}

	maxChannels, currentCount, canCreate, err := h.uc.GetChannelLimits(r.Context(), userID, isAdmin)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, &types.ChannelLimits{
		MaxChannels:  int32(maxChannels),
		CurrentCount: int32(currentCount),
		CanCreate:    canCreate,
	})
}

// ValidateHandle checks if a handle is available.
// GET /api/v1/channels/validate-handle?handle=xxx
func (h *ChannelHandler) ValidateHandle(w http.ResponseWriter, r *http.Request) {
	gc := ginadapter.GetGinContext(r)


	handle := gc.Query("handle")
	if handle == "" {
		server.Fail(gc, server.ErrBadRequest, "handle query parameter is required")
		return
	}

	if !validation.IsValidHandle(handle) {
		server.OK(gc, gin.H{
			"available": false,
			"reason":    "invalid_format",
		})
		return
	}

	available, err := h.uc.ValidateHandle(r.Context(), handle)
	if err != nil {
		server.Fail(gc, server.ErrInternal, err.Error())
		return
	}

	server.OK(gc, gin.H{
		"available": available,
	})
}
