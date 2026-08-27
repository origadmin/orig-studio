/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel module - gRPC implementation of ChannelServiceServer
 */

package service

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server/validation"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

// ChannelServiceServer implements the gRPC ChannelServiceServer interface.
type ChannelServiceServer struct {
	pb.UnimplementedChannelServiceServer
	uc        *biz.PlaylistChannelUseCase
	jwt       *auth.Manager
	settingUC *systembiz.SettingUseCase
	log       *log.Helper
}

// NewChannelServiceServer creates a new ChannelServiceServer.
func NewChannelServiceServer(uc *biz.PlaylistChannelUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase, logger log.Logger) *ChannelServiceServer {
	return &ChannelServiceServer{
		uc:        uc,
		jwt:       jwt,
		settingUC: settingUC,
		log:       log.NewHelper(log.With(logger, "module", "service/channel_grpc")),
	}
}

// extractUserID extracts the user ID from the gRPC context.
// It first checks the context value, then falls back to JWT parsing from metadata.
func (s *ChannelServiceServer) extractUserID(ctx context.Context) string {
	if id, ok := ctx.Value("user_id").(string); ok && id != "" {
		return id
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("grpcgateway-authorization")
	}
	for _, header := range authHeaders {
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			continue
		}
		claims, err := s.jwt.Parse(token)
		if err != nil {
			continue
		}
		return claims.GetUserID()
	}
	return ""
}

// extractClaims extracts the JWT claims from the gRPC context metadata.
func (s *ChannelServiceServer) extractClaims(ctx context.Context) *auth.Claims {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("grpcgateway-authorization")
	}
	for _, header := range authHeaders {
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			continue
		}
		claims, err := s.jwt.Parse(token)
		if err != nil {
			continue
		}
		return claims
	}
	return nil
}

// protoPrivacyToBiz converts a proto Privacy enum to the biz layer string.
func protoPrivacyToBiz(p types.Privacy) string {
	switch p {
	case types.Privacy_PRIVACY_PRIVATE:
		return "PRIVATE"
	case types.Privacy_PRIVACY_UNLISTED:
		return "UNLISTED"
	case types.Privacy_PRIVACY_PAID:
		return "PAID"
	case types.Privacy_PRIVACY_SUBSCRIBERS_ONLY:
		return "SUBSCRIBERS_ONLY"
	default:
		return "PUBLIC"
	}
}

// protoStatusToBiz converts a proto ChannelStatus enum to the biz layer string.
func protoStatusToBiz(s types.ChannelStatus) string {
	switch s {
	case types.ChannelStatus_CHANNEL_STATUS_INACTIVE:
		return "INACTIVE"
	case types.ChannelStatus_CHANNEL_STATUS_SUSPENDED:
		return "SUSPENDED"
	case types.ChannelStatus_CHANNEL_STATUS_PENDING_REVIEW:
		return "PENDING_REVIEW"
	default:
		return "ACTIVE"
	}
}

// bizSubscriptionVideoToMedia converts a biz SubscriptionVideoItem to a proto Media.
func bizSubscriptionVideoToMedia(item *biz.SubscriptionVideoItem) *types.Media {
	if item == nil {
		return nil
	}
	m := &types.Media{
		Id:             item.ID,
		ShortToken:     item.ShortToken,
		Title:          item.Title,
		Description:    item.Description,
		Thumbnail:      item.Thumbnail,
		Poster:         item.Poster,
		Duration:       int32(item.Duration),
		ViewCount:      item.ViewCount,
		LikeCount:      item.LikeCount,
		CommentCount:   item.CommentCount,
		Type:           item.Type,
		State:          item.State,
		ChannelId:      item.ChannelID,
		UserId:         item.UserID,
		EncodingStatus: item.EncodingStatus,
	}
	// Add user information if available
	if item.Username != "" || item.Nickname != "" || item.UserAvatar != "" {
		m.User = &types.User{
			Id:       item.UserID,
			Username: item.Username,
			Nickname: item.Nickname,
			Avatar:   item.UserAvatar,
			Logo:     item.UserAvatar,
		}
	}
	if !item.CreateTime.IsZero() {
		m.CreateTime = timestamppb.New(item.CreateTime)
	}
	if !item.PublishedAt.IsZero() {
		m.PublishedAt = timestamppb.New(item.PublishedAt)
	}
	return m
}

// bizSubscriptionVideosToMedias converts a slice of biz SubscriptionVideoItem to proto Media.
func bizSubscriptionVideosToMedias(items []*biz.SubscriptionVideoItem) []*types.Media {
	result := make([]*types.Media, 0, len(items))
	for _, item := range items {
		if m := bizSubscriptionVideoToMedia(item); m != nil {
			result = append(result, m)
		}
	}
	return result
}

// bizChannelPlaylistToProto converts a biz ChannelPlaylistItem to a proto Playlist.
func bizChannelPlaylistToProto(item *biz.ChannelPlaylistItem) *types.Playlist {
	if item == nil {
		return nil
	}
	p := &types.Playlist{
		Id:          item.ID,
		ShortToken:  item.ShortToken,
		Title:       item.Title,
		Description: item.Description,
		UserId:      item.UserID,
	}
	if !item.CreateTime.IsZero() {
		p.CreateTime = timestamppb.New(item.CreateTime)
	}
	return p
}

// bizChannelPlaylistsToProto converts a slice of biz ChannelPlaylistItem to proto Playlist.
func bizChannelPlaylistsToProto(items []*biz.ChannelPlaylistItem) []*types.Playlist {
	result := make([]*types.Playlist, 0, len(items))
	for _, item := range items {
		if p := bizChannelPlaylistToProto(item); p != nil {
			result = append(result, p)
		}
	}
	return result
}

// normalizePagination normalizes page and pageSize values to valid defaults.
func normalizePagination(page, pageSize int32) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return int(page), int(pageSize)
}

// ListChannels returns channels with optional filtering.
// Supports 2 modes:
//  1. user_id set  -> Get all channels for a user
//  2. (no params)  -> List all public channels (paginated)
func (s *ChannelServiceServer) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	page, pageSize := normalizePagination(req.Page, req.PageSize)

	// Mode: by user_id
	if req.UserId != nil && *req.UserId != "" {
		items, total, err := s.uc.ListUserChannels(ctx, *req.UserId, page, pageSize)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list user channels: %v", err)
		}
		return &pb.ListChannelsResponse{
			Items:    bizChannelsToProto(items),
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// Default: list all public channels (paginated)
	items, total, err := s.uc.ListChannels(ctx, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list channels: %v", err)
	}
	return &pb.ListChannelsResponse{
		Items:    bizChannelsToProto(items),
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// GetMyChannel returns the current authenticated user's channel.
// If the user has no channel, a default public channel is automatically created.
func (s *ChannelServiceServer) GetMyChannel(ctx context.Context, req *pb.GetMyChannelRequest) (*pb.GetMyChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	userID := claims.GetUserID()
	channels, _, err := s.uc.ListUserChannels(ctx, userID, 1, 1)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user channel: %v", err)
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

		created, createErr := s.uc.CreateChannel(ctx, defaultChannel)
		if createErr != nil {
			s.log.Warnf("failed to auto-create default channel for user %s: %v", userID, createErr)
			return &pb.GetMyChannelResponse{
				Channel: nil,
			}, nil
		}

		s.log.Infof("auto-created default channel %s for user %s", created.ShortToken, userID)
		return &pb.GetMyChannelResponse{
			Channel: bizChannelToProto(created),
		}, nil
	}

	return &pb.GetMyChannelResponse{
		Channel: bizChannelToProto(channels[0]),
	}, nil
}

// UpdateMyChannelHandle updates the current user's channel handle/slug.
func (s *ChannelServiceServer) UpdateMyChannelHandle(ctx context.Context, req *pb.UpdateMyChannelHandleRequest) (*pb.UpdateMyChannelHandleResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Handle == "" {
		return nil, status.Error(codes.InvalidArgument, "handle is required")
	}

	channels, _, err := s.uc.ListUserChannels(ctx, claims.GetUserID(), 1, 1)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user channel: %v", err)
	}

	if len(channels) == 0 {
		return nil, status.Error(codes.NotFound, "You don't have a channel yet")
	}

	ch := channels[0]
	ch.Handle = req.Handle

	updated, err := s.uc.UpdateChannel(ctx, ch, claims.GetUserID(), claims.IsAdmin())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update channel handle: %v", err)
	}

	return &pb.UpdateMyChannelHandleResponse{
		Channel: bizChannelToProto(updated),
	}, nil
}

// ValidateChannelHandle checks if a handle is available.
func (s *ChannelServiceServer) ValidateChannelHandle(ctx context.Context, req *pb.ValidateChannelHandleRequest) (*pb.ValidateChannelHandleResponse, error) {
	if req.Handle == "" {
		return nil, status.Error(codes.InvalidArgument, "handle is required")
	}

	if !validation.IsValidHandle(req.Handle) {
		return &pb.ValidateChannelHandleResponse{
			Available: false,
			Message:   "invalid_format",
		}, nil
	}

	available, err := s.uc.ValidateHandle(ctx, req.Handle)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate handle: %v", err)
	}

	msg := ""
	if !available {
		msg = "handle_already_taken"
	}
	return &pb.ValidateChannelHandleResponse{
		Available: available,
		Message:   msg,
	}, nil
}

// GetChannelByToken returns a single channel by short_token.
func (s *ChannelServiceServer) GetChannelByToken(ctx context.Context, req *pb.GetChannelByTokenRequest) (*pb.GetChannelByTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// BUG-237/2026-08-20: resolve by short_token, UUID id, handle, or slug
	// (slug follows the authoritative tag slug rule — Base58 for non-ASCII,
	// hyphenated ASCII — so lookups must not be restricted to 6-12 tokens).
	chItem, err := s.uc.ResolveChannel(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.NotFound, "channel_not_found: no channel exists with the given token/slug")
	}

	return &pb.GetChannelByTokenResponse{
		Channel: bizChannelToProto(chItem),
	}, nil
}

// GetChannelVideos returns videos for a specific channel by short_token.
func (s *ChannelServiceServer) GetChannelVideos(ctx context.Context, req *pb.GetChannelVideosRequest) (*pb.GetChannelVideosResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if !validation.IsValidShortToken(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "invalid_token_format")
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)

	items, total, err := s.uc.GetChannelVideos(ctx, req.Token, "newest", page, pageSize)
	if err != nil {
		if err.Error() == "channel_not_found" {
			return nil, status.Error(codes.NotFound, "channel_not_found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get channel videos: %v", err)
	}

	return &pb.GetChannelVideosResponse{
		Items:    bizSubscriptionVideosToMedias(items),
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// GetChannelPlaylists returns playlists for a specific channel by short_token.
func (s *ChannelServiceServer) GetChannelPlaylists(ctx context.Context, req *pb.GetChannelPlaylistsRequest) (*pb.GetChannelPlaylistsResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if !validation.IsValidShortToken(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "invalid_token_format")
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)

	items, total, err := s.uc.GetChannelPlaylists(ctx, req.Token, page, pageSize)
	if err != nil {
		if err.Error() == "channel_not_found" {
			return nil, status.Error(codes.NotFound, "channel_not_found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get channel playlists: %v", err)
	}

	return &pb.GetChannelPlaylistsResponse{
		Items:    bizChannelPlaylistsToProto(items),
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// UpdateChannelNotification updates notification preferences for a channel subscription.
func (s *ChannelServiceServer) UpdateChannelNotification(ctx context.Context, req *pb.UpdateChannelNotificationRequest) (*pb.UpdateChannelNotificationResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	isSubscribed, err := s.uc.IsSubscribedToChannel(ctx, req.Token, claims.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check subscription: %v", err)
	}

	if !isSubscribed {
		return nil, status.Error(codes.InvalidArgument, "Not subscribed to this channel")
	}

	// BUG-198: persist the real notification preference instead of a NO-OP stub.
	if err := s.uc.UpdateSubscriptionNotificationPreference(ctx, req.Token, claims.GetUserID(), req.GetNotificationPreference()); err != nil {
		if err.Error() == "invalid_notification_preference" {
			return nil, status.Error(codes.InvalidArgument, "invalid_notification_preference")
		}
		return nil, status.Errorf(codes.Internal, "failed to update notification preference: %v", err)
	}

	return &pb.UpdateChannelNotificationResponse{
		Success: true,
	}, nil
}

// GetChannelSubscribers returns subscribers for a channel with optional count parameter.
func (s *ChannelServiceServer) GetChannelSubscribers(ctx context.Context, req *pb.GetChannelSubscribersRequest) (*pb.GetChannelSubscribersResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	if req.CountOnly {
		count, err := s.uc.GetChannelSubscriberCount(ctx, req.Token)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get subscriber count: %v", err)
		}
		return &pb.GetChannelSubscribersResponse{
			Count: int32(count),
		}, nil
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)

	subscribers, total, err := s.uc.GetChannelSubscribers(ctx, req.Token, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get channel subscribers: %v", err)
	}

	// subscribers from biz layer are user IDs ([]string); cannot convert to proto User without lookup
	_ = subscribers

	return &pb.GetChannelSubscribersResponse{
		Subscribers: []*types.User{},
		Total:       int32(total),
		Page:        int32(page),
		PageSize:    int32(pageSize),
	}, nil
}

// GetChannelSubscription returns the subscription status for the current user.
func (s *ChannelServiceServer) GetChannelSubscription(ctx context.Context, req *pb.GetChannelSubscriptionRequest) (*pb.GetChannelSubscriptionResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	isSubscribed, err := s.uc.IsSubscribedToChannel(ctx, req.Token, claims.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check subscription: %v", err)
	}

	// BUG-198: reflect the subscriber's notification preference on the button (gateway bridges GET via gRPC-gateway).
	pref := "all"
	if isSubscribed {
		pref, err = s.uc.GetSubscriptionNotificationPreference(ctx, req.Token, claims.GetUserID())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get notification preference: %v", err)
		}
	}

	return &pb.GetChannelSubscriptionResponse{
		IsSubscribed:          isSubscribed,
		NotificationPreference: pref,
	}, nil
}

// SubscribeChannel subscribes the current user to a channel.
func (s *ChannelServiceServer) SubscribeChannel(ctx context.Context, req *pb.SubscribeChannelRequest) (*pb.SubscribeChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	err := s.uc.SubscribeToChannel(ctx, req.Token, claims.GetUserID())
	if err != nil {
		errMsg := err.Error()
		if errMsg == "cannot_subscribe_own_channel" {
			return nil, status.Error(codes.InvalidArgument, errMsg)
		}
		return nil, status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	return &pb.SubscribeChannelResponse{
		Success:      true,
		IsSubscribed: true,
	}, nil
}

// UnsubscribeChannel unsubscribes the current user from a channel.
func (s *ChannelServiceServer) UnsubscribeChannel(ctx context.Context, req *pb.UnsubscribeChannelRequest) (*pb.UnsubscribeChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	err := s.uc.UnsubscribeFromChannel(ctx, req.Token, claims.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unsubscribe: %v", err)
	}

	return &pb.UnsubscribeChannelResponse{
		Success:      true,
		IsSubscribed: false,
	}, nil
}

// CreateChannel creates a new channel for the authenticated user.
func (s *ChannelServiceServer) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.CreateChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Channel == nil {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	if req.Channel.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Channel.ShortToken != "" && !validation.IsValidShortToken(req.Channel.ShortToken) {
		return nil, status.Error(codes.InvalidArgument, "invalid_short_token_format: must be 6-12 chars, alphanumeric, underscores and hyphens only")
	}

	slug := hashtag.GenerateTagSlug(req.Channel.Name)

	chItem := &biz.Channel{
		Name:        req.Channel.Name,
		Title:       req.Channel.Name,
		Slug:        slug,
		Handle:      slug,
		Description: req.Channel.Description,
		Avatar:      req.Channel.Avatar,
		Banner:      req.Channel.Banner,
		BannerLogo:  req.Channel.BannerLogo,
		Privacy:     protoPrivacyToBiz(req.Channel.Privacy),
		Tags:        req.Channel.Tags,
		Status:      "ACTIVE",
		UserID:      claims.GetUserID(),
		ShortToken:  req.Channel.ShortToken,
	}

	if req.Channel.CategoryId != 0 {
		catID := req.Channel.CategoryId
		chItem.CategoryID = &catID
	}

	if chItem.Privacy == "" {
		chItem.Privacy = "PUBLIC"
	}

	created, err := s.uc.CreateChannel(ctx, chItem)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "channel_limit_reached") {
			return nil, status.Error(codes.ResourceExhausted, errMsg)
		}
		if strings.Contains(errMsg, "short_token_already_taken") || strings.Contains(errMsg, "handle_already_taken") {
			return nil, status.Error(codes.AlreadyExists, "short_token_already_taken")
		}
		if strings.Contains(errMsg, "too_many_tags") {
			return nil, status.Error(codes.InvalidArgument, errMsg)
		}
		return nil, status.Errorf(codes.Internal, "failed to create channel: %v", err)
	}

	return &pb.CreateChannelResponse{
		Channel: bizChannelToProto(created),
	}, nil
}

// AddChannelMedia adds a media item to a channel.
func (s *ChannelServiceServer) AddChannelMedia(ctx context.Context, req *pb.AddChannelMediaRequest) (*pb.AddChannelMediaResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	if req.MediaId == "" {
		return nil, status.Error(codes.InvalidArgument, "media_id is required")
	}

	err := s.uc.AddMediaToChannel(
		ctx,
		req.Token,
		req.MediaId,
		claims.GetUserID(),
		claims.IsAdmin(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add media to channel: %v", err)
	}

	return &pb.AddChannelMediaResponse{
		Success: true,
	}, nil
}

// RemoveChannelMedia removes a media item from a channel.
func (s *ChannelServiceServer) RemoveChannelMedia(ctx context.Context, req *pb.RemoveChannelMediaRequest) (*pb.RemoveChannelMediaResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	if req.MediaId == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid media ID")
	}

	err := s.uc.RemoveMediaFromChannel(
		ctx,
		req.Token,
		req.MediaId,
		claims.GetUserID(),
		claims.IsAdmin(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove media from channel: %v", err)
	}

	return &pb.RemoveChannelMediaResponse{
		Success: true,
	}, nil
}

// InviteUserToChannel invites a user to join a channel.
func (s *ChannelServiceServer) InviteUserToChannel(ctx context.Context, req *pb.InviteUserToChannelRequest) (*pb.InviteUserToChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid channel token")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.uc.InviteUserToChannel(ctx, req.Token, req.UserId, claims.GetUserID(), claims.IsAdmin())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to invite user to channel: %v", err)
	}

	return &pb.InviteUserToChannelResponse{
		Success: true,
	}, nil
}

// AcceptChannelInvitation accepts a channel invitation.
func (s *ChannelServiceServer) AcceptChannelInvitation(ctx context.Context, req *pb.AcceptChannelInvitationRequest) (*pb.AcceptChannelInvitationResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid ID")
	}

	err := s.uc.AcceptChannelInvitation(ctx, req.Id, claims.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to accept channel invitation: %v", err)
	}

	return &pb.AcceptChannelInvitationResponse{
		Success: true,
	}, nil
}

// RejectChannelInvitation rejects a channel invitation.
func (s *ChannelServiceServer) RejectChannelInvitation(ctx context.Context, req *pb.RejectChannelInvitationRequest) (*pb.RejectChannelInvitationResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid ID")
	}

	err := s.uc.RejectChannelInvitation(ctx, req.Id, claims.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reject channel invitation: %v", err)
	}

	return &pb.RejectChannelInvitationResponse{
		Success: true,
	}, nil
}

// GetChannelInvitations returns the current user's channel invitations.
func (s *ChannelServiceServer) GetChannelInvitations(ctx context.Context, req *pb.GetChannelInvitationsRequest) (*pb.GetChannelInvitationsResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	invitations, err := s.uc.GetChannelInvitations(ctx, claims.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get channel invitations: %v", err)
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)

	return &pb.GetChannelInvitationsResponse{
		Total:    int32(len(invitations)),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// UpdateChannel updates a channel by short_token. Only the owner can update.
func (s *ChannelServiceServer) UpdateChannel(ctx context.Context, req *pb.UpdateChannelRequest) (*pb.UpdateChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	existingChannel, err := s.uc.GetByShortToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.NotFound, "channel_not_found")
	}

	// Start with existing values
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

	// Apply updates from proto Channel if provided
	if req.Channel != nil {
		if req.Channel.Name != "" {
			chItem.Name = req.Channel.Name
			chItem.Slug = hashtag.GenerateTagSlug(req.Channel.Name) // Regenerate slug on name change
		}
		if req.Channel.Title != "" {
			chItem.Title = req.Channel.Title
		}
		if req.Channel.Description != "" {
			chItem.Description = req.Channel.Description
		}
		if req.Channel.Avatar != "" {
			chItem.Avatar = req.Channel.Avatar
		}
		if req.Channel.Banner != "" {
			chItem.Banner = req.Channel.Banner
		}
		if req.Channel.BannerLogo != "" {
			chItem.BannerLogo = req.Channel.BannerLogo
		}
		if req.Channel.Privacy != types.Privacy_PRIVACY_UNSPECIFIED {
			chItem.Privacy = protoPrivacyToBiz(req.Channel.Privacy)
		}
		if req.Channel.Status != types.ChannelStatus_CHANNEL_STATUS_UNSPECIFIED {
			chItem.Status = protoStatusToBiz(req.Channel.Status)
		}
		if req.Channel.Tags != nil {
			chItem.Tags = req.Channel.Tags
		}
		if req.Channel.CategoryId != 0 {
			catID := req.Channel.CategoryId
			chItem.CategoryID = &catID
		}
		if req.Channel.Links != nil {
			chItem.Links = make([]biz.ChannelLink, len(req.Channel.Links))
			for i, l := range req.Channel.Links {
				chItem.Links[i] = biz.ChannelLink{
					Type:     l.Type,
					Platform: l.Platform,
					URL:      l.Url,
					Title:    l.Title,
				}
			}
		}
	}

	updated, err := s.uc.UpdateChannel(
		ctx,
		chItem,
		claims.GetUserID(),
		claims.IsAdmin(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update channel: %v", err)
	}

	return &pb.UpdateChannelResponse{
		Channel: bizChannelToProto(updated),
	}, nil
}

// DeleteChannel deletes a channel by short_token. Only the owner or admin can delete.
func (s *ChannelServiceServer) DeleteChannel(ctx context.Context, req *pb.DeleteChannelRequest) (*pb.DeleteChannelResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	existingChannel, err := s.uc.GetByShortToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.NotFound, "channel_not_found")
	}

	err = s.uc.DeleteChannel(ctx, existingChannel.ID, claims.GetUserID(), claims.IsAdmin())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete channel: %v", err)
	}

	return &pb.DeleteChannelResponse{
		Empty: &emptypb.Empty{},
	}, nil
}

// ResolveHandle resolves a @handle to a channel or user.
func (s *ChannelServiceServer) ResolveHandle(ctx context.Context, req *pb.ResolveHandleRequest) (*pb.ResolveHandleResponse, error) {
	if req.Handle == "" {
		return nil, status.Error(codes.InvalidArgument, "handle is required")
	}

	// Strip leading @ if present
	handle := strings.TrimPrefix(req.Handle, "@")

	result, err := s.uc.ResolveHandle(ctx, handle)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to resolve handle: %v", err)
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

	return &pb.ResolveHandleResponse{
		Resolution: pbResult,
	}, nil
}

// GetChannelLimits returns channel creation limits for the current user.
func (s *ChannelServiceServer) GetChannelLimits(ctx context.Context, req *pb.GetChannelLimitsRequest) (*pb.GetChannelLimitsResponse, error) {
	claims := s.extractClaims(ctx)
	isAdmin := false
	var userID string
	if claims != nil {
		userID = claims.GetUserID()
		isAdmin = claims.IsAdmin()
	}

	maxChannels, currentCount, canCreate, err := s.uc.GetChannelLimits(ctx, userID, isAdmin)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get channel limits: %v", err)
	}

	return &pb.GetChannelLimitsResponse{
		Limits: &types.ChannelLimits{
			MaxChannels:  int32(maxChannels),
			CurrentCount: int32(currentCount),
			CanCreate:    canCreate,
		},
	}, nil
}

// GetSubscriptionVideos returns the latest videos from all channels the current user is subscribed to.
func (s *ChannelServiceServer) GetSubscriptionVideos(ctx context.Context, req *pb.GetSubscriptionVideosRequest) (*pb.GetSubscriptionVideosResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	userID := claims.GetUserID()
	page, pageSize := normalizePagination(req.Page, req.PageSize)

	// Find all channel IDs the user is subscribed to
	channelIDs, err := s.uc.GetSubscribedChannelIDs(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get subscribed channel IDs: %v", err)
	}

	if len(channelIDs) == 0 {
		return &pb.GetSubscriptionVideosResponse{
			Items:    []*types.Media{},
			Total:    0,
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// Query videos from subscribed channels via biz layer
	items, total, err := s.uc.GetSubscriptionVideos(ctx, userID, channelIDs, "newest", page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get subscription videos: %v", err)
	}

	return &pb.GetSubscriptionVideosResponse{
		Items:    bizSubscriptionVideosToMedias(items),
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
