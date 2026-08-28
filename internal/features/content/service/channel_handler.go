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
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/server/validation"
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
		channels.GET("", h.ListChannels())

		// Current user's channels (requires auth)
		channels.GET("/me", server.WithJWTCtx(h.jwt, h.GetMyChannels()))
		channels.PUT("/me/handle", server.WithJWTCtx(h.jwt, h.UpdateMyHandle()))

		// Handle validation (public)
		channels.GET("/validate-handle", h.ValidateHandle())

		// ================================
		// 2. PATH PARAMETER ROUTES (WITH :token) - MUST BE AFTER STATIC
		// ================================
		// Single channel by short_token (RESTful, MediaCMS style)
		channels.GET("/:token", h.GetChannelByToken())

		// Channel videos and playlists
		channels.GET("/:token/videos", h.GetChannelVideos())
		channels.GET("/:token/playlists", h.GetChannelPlaylists())

		// Notification settings
		channels.PUT("/:token/notification", server.WithJWTCtx(h.jwt, h.UpdateNotificationSetting()))

		// ================================
		// 3. NESTED RESOURCE ROUTES (Subscription APIs)
		// ================================
		// Channel subscribers and subscription
		channels.GET("/:token/subscribers", h.GetChannelSubscribers())
		channels.GET("/:token/subscription", server.WithJWTCtx(h.jwt, h.GetSubscriptionStatus()))
		channels.POST("/:token/subscription", server.WithJWTCtx(h.jwt, h.SubscribeToChannel()))
		channels.DELETE("/:token/subscription", server.WithJWTCtx(h.jwt, h.UnsubscribeFromChannel()))

		// Protected write routes
		{
			channels.POST("", server.WithJWTCtx(h.jwt, h.CreateChannel()))
			// Media management within channel (by :token)
			channels.POST("/:token/medias", server.WithJWTCtx(h.jwt, h.AddMedia()))
			channels.DELETE("/:token/medias/:media_id", server.WithJWTCtx(h.jwt, h.RemoveMedia()))
			// Invitation management
			channels.POST("/:token/invitations", server.WithJWTCtx(h.jwt, h.InviteUserToChannel()))
			channels.POST("/invitations/:id/accept", server.WithJWTCtx(h.jwt, h.AcceptChannelInvitation()))
			channels.POST("/invitations/:id/reject", server.WithJWTCtx(h.jwt, h.RejectChannelInvitation()))
			channels.GET("/invitations", server.WithJWTCtx(h.jwt, h.GetChannelInvitations()))

			// ================================
			// UPDATE & DELETE by :token (not :id!)
			// Application uses short_token for all operations
			// Admin uses /admin/channels/:uuid for UUID-based operations
			// ================================
			channels.PUT("/:token", server.WithJWTCtx(h.jwt, h.UpdateChannel()))    // :token = short_token
			channels.DELETE("/:token", server.WithJWTCtx(h.jwt, h.DeleteChannel())) // :token = short_token
		}
	}

	// ================================
	// Handle resolution route (top-level, NOT under /channels)
	// ================================
	resolveGroup := r.Group("/resolve")
	{
		resolveGroup.GET("", h.ResolveHandle())
	}

	// ================================
	// System config routes (top-level, NOT under /channels)
	// ================================
	configGroup := r.Group("/system/config")
	{
		configGroup.GET("/channel-limits", h.GetChannelLimits())
	}

	// ================================
	// Admin channel list: REMOVED. BUG-237 stopgap — AdminHandler now owns
	// GET /admin/channels (full CRUD); duplicate registration panicked the
	// monolith (cmd/server) with "handlers are already registered".
	// ================================

	// ================================
	// Subscription feed routes (top-level, NOT under /channels)
	// ================================
	subsGroup := r.Group("/subscriptions")
	{
		subsGroup.GET("/videos", server.WithJWTCtx(h.jwt, h.GetSubscriptionVideos()))
	}
}

// GetChannelByToken returns a single channel by short_token (path parameter).
// This is the RECOMMENDED way to access a single channel (RESTful, MediaCMS style).
func (h *ChannelHandler) GetChannelByToken() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "token is required")
		}

		// Resolve by short_token, UUID id, handle, or slug (BUG-237/2026-08-20:
		// slug follows the tag slug rule — Base58 for non-ASCII names, hyphenated
		// ASCII — so lookups must not be restricted to 6-12 alphanumeric tokens).
		chItem, err := h.uc.ResolveChannel(ctx, token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found: no channel exists with the given token/slug")
		}

		return server.OKCtx(ctx, &pb.GetChannelResponse{
			Channel: bizChannelToProto(chItem),
		})
	}
}

// ListChannels returns channels with optional query parameters.
// Supports 3 modes:
//  1. ?username={value} -> Get default channel by username
//  2. ?user_id={value}  -> Get all channels for a user
//  3. (no params)       -> List all public channels (paginated)
func (h *ChannelHandler) ListChannels() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		username := ctx.QueryVar("username")
		userId := ctx.QueryVar("user_id")

		limit, _ := strconv.Atoi(ctx.QueryVar("limit"))
		if limit == 0 {
			limit = 20
		}
		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page == 0 {
			page = 1
		}
		// Normalize pagination parameters
		page, limit = repotypes.NormalizeHTTPPagination(page, limit)

		switch {
		case username != "":
			if !validation.IsValidUsername(username) {
				return server.FailCtx(ctx, server.ErrBadRequest, "invalid_username_format")
			}
			// Two-step: username -> user_id -> default channel
			chItem, err := h.uc.GetChannelByUsername(ctx, username)
			if err != nil {
				return server.FailCtx(ctx, server.ErrNotFound, fmt.Sprintf("channel not found for @%s", username))
			}
			return server.OKCtx(ctx, &pb.GetChannelResponse{
				Channel: bizChannelToProto(chItem),
			})

		case userId != "":
			items, total, err := h.uc.ListUserChannels(ctx, userId, page, limit)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
			pbChannels := bizChannelsToProto(items)
			return server.OKCtx(ctx, &pb.ListChannelsResponse{
				Items:    pbChannels,
				Total:    int32(total),
				Page:     int32(page),
				PageSize: int32(limit),
			})

		default:
			// List all public channels (paginated)
			items, total, err := h.uc.ListChannels(ctx, page, limit)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
			pbChannels := bizChannelsToProto(items)
			return server.OKCtx(ctx, &pb.ListChannelsResponse{
				Items:    pbChannels,
				Total:    int32(total),
				Page:     int32(page),
				PageSize: int32(limit),
			})
		}
	}
}

// AdminListChannels removed — dead code after the duplicate /admin/channels
// registration was removed (AdminHandler owns that route; see RegisterRoutes).

// CreateChannel creates a new channel for the authenticated user.
func (h *ChannelHandler) CreateChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		var input struct {
			Name        string   `json:"name" binding:"required,min=3,max=150"`
			Handle      string   `json:"handle"`
			Description string   `json:"description"`
			Avatar      string   `json:"avatar"`
			Banner      string   `json:"banner"`
			BannerLogo  string   `json:"banner_logo"`
			Privacy     string   `json:"privacy"`
			Tags        []string `json:"tags"`
			TagIDs      []int    `json:"tag_ids"`
			CategoryID  *int64   `json:"category_id"`
			ShortToken  string   `json:"short_token"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if input.ShortToken != "" && !validation.IsValidShortToken(input.ShortToken) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_short_token_format: must be 6-12 chars, alphanumeric, underscores and hyphens only")
		}

		slug := hashtag.GenerateTagSlug(input.Name)

		handle := input.Handle
		if handle == "" {
			handle = slug
		}

		chItem := &biz.Channel{
			Name:        input.Name,
			Title:       input.Name,
			Slug:        slug,
			Handle:      handle,
			Description: input.Description,
			Avatar:      input.Avatar,
			Banner:      input.Banner,
			BannerLogo:  input.BannerLogo,
			Privacy:     input.Privacy,
			Tags:        input.Tags,
			TagIDs:      input.TagIDs,
			CategoryID:  input.CategoryID,
			Status:      "ACTIVE",
			UserID:      claims.GetUserID(),
		}
		if input.ShortToken != "" {
			chItem.ShortToken = input.ShortToken
		}

		if chItem.Privacy == "" {
			chItem.Privacy = "PUBLIC"
		}

		created, err := h.uc.CreateChannel(ctx, chItem)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "channel_limit_reached") {
				return server.FailCtx(ctx, server.ErrBadRequest, errMsg)
			}
			if strings.Contains(errMsg, "short_token_already_taken") || strings.Contains(errMsg, "handle_already_taken") {
				return server.FailCtx(ctx, server.ErrConflict, "short_token_already_taken")
			}
			if strings.Contains(errMsg, "too_many_tags") {
				return server.FailCtx(ctx, server.ErrBadRequest, errMsg)
			}
			return server.FailCtx(ctx, server.ErrInternal, errMsg)
		}

		return server.CreatedCtx(ctx, &pb.CreateChannelResponse{
			Channel: bizChannelToProto(created),
		})
	}
}

// UpdateChannel updates a channel by short_token. Only the owner can update.
func (h *ChannelHandler) UpdateChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "token is required")
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
			CategoryID  *int64   `json:"category_id"`
			Tags        []string `json:"tags"`
			TagIDs      []int    `json:"tag_ids"`
			Links       []struct {
				Type     string `json:"type"`
				Platform string `json:"platform"`
				URL      string `json:"url"`
				Title    string `json:"title"`
			} `json:"links"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existingChannel, err := h.uc.GetByShortToken(ctx, token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found")
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
			chItem.Slug = hashtag.GenerateTagSlug(*input.Name) // Regenerate slug on name change
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
			ctx,
			chItem,
			claims.GetUserID(),
			claims.IsAdmin(),
		)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateChannelResponse{
			Channel: bizChannelToProto(updated),
		})
	}
}

// DeleteChannel deletes a channel by short_token. Only the owner or admin can delete.
func (h *ChannelHandler) DeleteChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "token is required")
		}

		existingChannel, err := h.uc.GetByShortToken(ctx, token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found")
		}

		err = h.uc.DeleteChannel(ctx, existingChannel.ID, claims.GetUserID(), claims.IsAdmin())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.DeleteChannelResponse{})
	}
}

// AddMedia adds a media item to a channel.
func (h *ChannelHandler) AddMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		var input struct {
			MediaID string `json:"media_id" binding:"required"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		err := h.uc.AddMediaToChannel(
			ctx,
			token,
			input.MediaID,
			claims.GetUserID(),
			claims.IsAdmin(),
		)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.AddChannelMediaResponse{
			Success: true,
		})
	}
}

// RemoveMedia removes a media item from a channel (sets channel_id to null).
func (h *ChannelHandler) RemoveMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}
		mediaId := ctx.Var("media_id")
		if mediaId == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid media ID")
		}

		err := h.uc.RemoveMediaFromChannel(
			ctx,
			token,
			mediaId,
			claims.GetUserID(),
			claims.IsAdmin(),
		)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.RemoveChannelMediaResponse{
			Success: true,
		})
	}
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (h *ChannelHandler) GetChannelSubscribers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		if ctx.QueryVar("count") == "true" {
			count, err := h.uc.GetChannelSubscriberCount(ctx, token)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
			return server.OKCtx(ctx, &pb.GetChannelSubscribersResponse{
				Count: int32(count),
			})
		}

		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(ctx.QueryVar("page_size"))
		if pageSize == 0 {
			pageSize = 20
		}
		// Normalize pagination parameters
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		subscribers, total, err := h.uc.GetChannelSubscribers(ctx, token, page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		pbSubscribers := make([]*types.User, len(subscribers))
		// subscribers from biz layer are interface{}; convert to proto User where possible
		_ = pbSubscribers // placeholder until proper conversion is available

		return server.OKCtx(ctx, &pb.GetChannelSubscribersResponse{
			Subscribers: pbSubscribers,
			Total:       int32(total),
			Page:        int32(page),
			PageSize:    int32(pageSize),
		})
	}
}

// GetSubscriptionStatus returns the subscription status for the current user.
func (h *ChannelHandler) GetSubscriptionStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		isSubscribed, err := h.uc.IsSubscribedToChannel(ctx, token, claims.GetUserID())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		// BUG-198: reflect the subscriber's notification preference on the button.
		pref := "all"
		if isSubscribed {
			pref, err = h.uc.GetSubscriptionNotificationPreference(ctx, token, claims.GetUserID())
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
		}

		return server.OKCtx(ctx, &pb.GetChannelSubscriptionResponse{
			IsSubscribed:         isSubscribed,
			NotificationPreference: pref,
		})
	}
}

// SubscribeToChannel subscribes the current user to a channel.
func (h *ChannelHandler) SubscribeToChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		err := h.uc.SubscribeToChannel(ctx, token, claims.GetUserID())
		if err != nil {
			errMsg := err.Error()
			if errMsg == "cannot_subscribe_own_channel" {
				return server.FailCtx(ctx, server.ErrBadRequest, errMsg)
			}
			return server.FailCtx(ctx, server.ErrInternal, errMsg)
		}

		return server.OKCtx(ctx, &pb.SubscribeChannelResponse{
			Success:      true,
			IsSubscribed: true,
		})
	}
}

// UnsubscribeFromChannel unsubscribes the current user from a channel.
func (h *ChannelHandler) UnsubscribeFromChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		err := h.uc.UnsubscribeFromChannel(ctx, token, claims.GetUserID())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UnsubscribeChannelResponse{
			Success:      true,
			IsSubscribed: false,
		})
	}
}

// InviteUserToChannel invites a user to join a channel.
func (h *ChannelHandler) InviteUserToChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		var input struct {
			UserID string `json:"user_id" binding:"required"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		err := h.uc.InviteUserToChannel(ctx, token, input.UserID, claims.GetUserID(), claims.IsAdmin())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.SubscribeChannelResponse{
			Success:      true,
			IsSubscribed: true,
		})
	}
}

// AcceptChannelInvitation accepts a channel invitation.
func (h *ChannelHandler) AcceptChannelInvitation() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid ID")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		err := h.uc.AcceptChannelInvitation(ctx, id, claims.GetUserID())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.SubscribeChannelResponse{
			Success:      true,
			IsSubscribed: true,
		})
	}
}

// RejectChannelInvitation rejects a channel invitation.
func (h *ChannelHandler) RejectChannelInvitation() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid ID")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		err := h.uc.RejectChannelInvitation(ctx, id, claims.GetUserID())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UnsubscribeChannelResponse{
			Success:      true,
			IsSubscribed: false,
		})
	}
}

// GetChannelInvitations returns the current user's channel invitations.
func (h *ChannelHandler) GetChannelInvitations() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		invitations, err := h.uc.GetChannelInvitations(ctx, claims.GetUserID())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		// TODO: Define ChannelInvitation proto type and convert properly
		_ = invitations
		return server.OKCtx(ctx, &pb.ListChannelsResponse{
			Items: []*types.Channel{},
		})
	}
}

// GetMyChannel returns the current authenticated user's channel.
// If the user has no channel, a default public channel is automatically created.
func (h *ChannelHandler) GetMyChannel() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		userID := claims.GetUserID()
		channels, _, err := h.uc.ListUserChannels(ctx, userID, 1, 1)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		if len(channels) == 0 {
			// Auto-create default channel for new users
			defaultName := claims.Username
			if len(defaultName) < 3 {
				defaultName = "My Channel"
			}
			slug := hashtag.GenerateTagSlug(defaultName)
			if slug == "" {
				slug = "my-channel"
			}

			defaultChannel := &biz.Channel{
				Name:    defaultName,
				Title:   defaultName,
				Slug:    slug,
				Handle:  slug,
				Privacy: "PUBLIC",
				Status:  "ACTIVE",
				Tags:    []string{},
				UserID:  userID,
			}

			created, createErr := h.uc.CreateChannel(ctx, defaultChannel)
			if createErr != nil {
				return server.OKCtx(ctx, &pb.GetChannelResponse{
					Channel: nil,
				})
			}

			return server.OKCtx(ctx, &pb.GetChannelResponse{
				Channel: bizChannelToProto(created),
			})
		}

		return server.OKCtx(ctx, &pb.GetChannelResponse{
			Channel: bizChannelToProto(channels[0]),
		})
	}
}

// UpdateMyHandle updates the current user's channel handle/slug.
func (h *ChannelHandler) UpdateMyHandle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		var input struct {
			Handle string `json:"handle" binding:"required,min=3,max=39"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		channels, _, err := h.uc.ListUserChannels(ctx, claims.GetUserID(), 1, 1)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		if len(channels) == 0 {
			return server.FailCtx(ctx, server.ErrNotFound, "You don't have a channel yet")
		}

		ch := channels[0]
		ch.Handle = input.Handle

		updated, err := h.uc.UpdateChannel(ctx, ch, claims.GetUserID(), claims.IsAdmin())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateChannelResponse{
			Channel: bizChannelToProto(updated),
		})
	}
}

// UpdateNotificationSetting updates notification preferences for a channel subscription.
func (h *ChannelHandler) UpdateNotificationSetting() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "Invalid channel token")
		}

		var input struct {
			Setting string `json:"setting" binding:"required,oneof=all personalized none"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		isSubscribed, err := h.uc.IsSubscribedToChannel(ctx, token, claims.GetUserID())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		if !isSubscribed {
			return server.FailCtx(ctx, server.ErrBadRequest, "Not subscribed to this channel")
		}

		// BUG-198: persist the real notification preference instead of a fake control.
		if err := h.uc.UpdateSubscriptionNotificationPreference(ctx, token, claims.GetUserID(), input.Setting); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateChannelNotificationResponse{
			Success: true,
		})
	}
}

// GetSubscriptionVideos returns the latest videos from all channels the current user is subscribed to.
// Supports pagination, sorting, and channel filtering.
// Query params: page, limit, sort_by (newest|most_viewed|trending), channel_ids (comma-separated).
func (h *ChannelHandler) GetSubscriptionVideos() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID := claims.GetUserID()

		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(ctx.QueryVar("limit"))
		if limit < 1 {
			limit = 20
		}
		// Normalize pagination parameters
		page, limit = repotypes.NormalizeHTTPPagination(page, limit)

		sortBy := ctx.QueryVar("sort_by")
		switch sortBy {
		case "newest", "most_viewed", "trending":
			// valid
		default:
			sortBy = "newest"
		}

		// Find all channel IDs the user is subscribed to
		channelIDs, err := h.uc.GetSubscribedChannelIDs(ctx, userID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		if len(channelIDs) == 0 {
			return server.OKCtx(ctx, &pb.GetChannelMediasResponse{
				Items:    []*types.Media{},
				Total:    0,
				Page:     int32(page),
				PageSize: int32(limit),
			})
		}

		// Apply channel_ids filter if provided
		if channelIDsParam := ctx.QueryVar("channel_ids"); channelIDsParam != "" {
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
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		pbMedias := bizSubscriptionVideosToMedias(items)

		return server.OKCtx(ctx, &pb.GetChannelMediasResponse{
			Items:    pbMedias,
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(limit),
		})
	}
}

// GetChannelVideos returns videos for a specific channel by short_token.
// Supports pagination and sorting.
// Query params: page, limit, sort_by (newest|oldest|popular).
func (h *ChannelHandler) GetChannelVideos() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "token is required")
		}

		if !validation.IsValidShortToken(token) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_token_format")
		}

		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(ctx.QueryVar("limit"))
		if limit < 1 {
			limit = 20
		}
		// Normalize pagination parameters
		page, limit = repotypes.NormalizeHTTPPagination(page, limit)

		sortBy := ctx.QueryVar("sort_by")
		switch sortBy {
		case "newest", "oldest", "popular":
			// valid
		default:
			sortBy = "newest"
		}

		// Query channel videos via biz layer
		items, total, err := h.uc.GetChannelVideos(ctx, token, sortBy, page, limit)
		if err != nil {
			errMsg := err.Error()
			if errMsg == "channel_not_found" {
				return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found")
			}
			return server.FailCtx(ctx, server.ErrInternal, errMsg)
		}

		pbMedias := bizSubscriptionVideosToMedias(items)

		return server.OKCtx(ctx, &pb.GetChannelMediasResponse{
			Items:    pbMedias,
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(limit),
		})
	}
}

// GetChannelPlaylists returns playlists for a specific channel by short_token.
// The channel's owner user_id is used to look up playlists.
// Supports pagination.
// Query params: page, limit.
func (h *ChannelHandler) GetChannelPlaylists() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "token is required")
		}

		if !validation.IsValidShortToken(token) {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid_token_format")
		}

		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(ctx.QueryVar("limit"))
		if limit < 1 {
			limit = 20
		}
		// Normalize pagination parameters
		page, limit = repotypes.NormalizeHTTPPagination(page, limit)

		// Query channel playlists via biz layer
		items, total, err := h.uc.GetChannelPlaylists(ctx, token, page, limit)
		if err != nil {
			errMsg := err.Error()
			if errMsg == "channel_not_found" {
				return server.FailCtx(ctx, server.ErrNotFound, "channel_not_found")
			}
			return server.FailCtx(ctx, server.ErrInternal, errMsg)
		}

		pbPlaylists := bizChannelPlaylistsToProto(items)

		return server.OKCtx(ctx, &pb.GetPlaylistsResponse{
			Items:    pbPlaylists,
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(limit),
		})
	}
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

// ResolveHandle resolves a @handle to a channel or user.
// GET /api/v1/resolve?handle=xxx
func (h *ChannelHandler) ResolveHandle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		handle := ctx.QueryVar("handle")
		if handle == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "handle query parameter is required")
		}

		// Strip leading @ if present
		handle = strings.TrimPrefix(handle, "@")

		result, err := h.uc.ResolveHandle(ctx, handle)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
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

		return server.OKCtx(ctx, pbResult)
	}
}

// GetMyChannels returns all channels for the authenticated user.
// GET /api/v1/channels/me
func (h *ChannelHandler) GetMyChannels() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}

		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(ctx.QueryVar("page_size"))
		if pageSize < 1 {
			pageSize = 20
		}
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		channels, total, err := h.uc.ListUserChannels(ctx, claims.GetUserID(), page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		// Mark is_owner for all channels
		for _, ch := range channels {
			ch.IsOwner = true
		}

		return server.OKCtx(ctx, &pb.ListChannelsResponse{
			Items:    bizChannelsToProto(channels),
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
	}
}

// GetChannelLimits returns channel creation limits for the current user.
// GET /api/v1/system/config/channel-limits
func (h *ChannelHandler) GetChannelLimits() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		isAdmin := false
		var userID string
		if claims, ok := server.GetClaimsCtx(ctx); ok {
			userID = claims.GetUserID()
			isAdmin = claims.IsAdmin()
		}

		maxChannels, currentCount, canCreate, err := h.uc.GetChannelLimits(ctx, userID, isAdmin)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &types.ChannelLimits{
			MaxChannels:  int32(maxChannels),
			CurrentCount: int32(currentCount),
			CanCreate:    canCreate,
		})
	}
}

// ValidateHandle checks if a handle is available.
// GET /api/v1/channels/validate-handle?handle=xxx
func (h *ChannelHandler) ValidateHandle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		handle := ctx.QueryVar("handle")
		if handle == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "handle query parameter is required")
		}

		if !validation.IsValidHandle(handle) {
			return server.OKCtx(ctx, map[string]any{
				"available": false,
				"reason":    "invalid_format",
			})
		}

		available, err := h.uc.ValidateHandle(ctx, handle)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"available": available,
		})
	}
}
