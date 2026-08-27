/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service is the service layer for the user service.
package service

import (
	"context"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/origadmin/contrib/security"
	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"origadmin/application/origstudio/api/gen/v1/types"
	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	userv1 "origadmin/application/origstudio/api/gen/v1/user"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/infra/auth"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/user/biz"
	"origadmin/application/origstudio/internal/features/user/dto"
)

// UserService implements the user gRPC service.
type UserService struct {
	userv1.UnimplementedUserServiceServer
	uc               *biz.UserUseCase
	publisher        message.Publisher
	hasher           hash.Crypto
	jwtMgr           *auth.Manager
	likeFavoriteUC   *contentbiz.LikeFavoriteUseCase
	playlistUC       *contentbiz.PlaylistChannelUseCase
	historyUC        *contentbiz.HistoryUseCase
	mediaClient      mediav1.MediaServiceClient
	log              *log.Helper
}

// resolveRefreshToken returns the effective refresh token.
//
// Priority:
//  1. explicit bodyToken (proto contract `refresh_token` field)
//  2. `Authorization: Bearer <token>` header carried via gRPC incoming
//     metadata (gateway authForwardInterceptor) or directly via an HTTP
//     adapter context. `grpcgateway-authorization` is accepted as an alias.
//
// The `Bearer ` prefix is stripped when present; any other scheme (e.g.
// Basic) is ignored and an empty string is returned so the caller can
// produce a clean 401.
func resolveRefreshToken(ctx context.Context, bodyToken string) string {
	if bodyToken != "" {
		return bodyToken
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
		const prefix = "Bearer "
		if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
			return strings.TrimSpace(header[len(prefix):])
		}
	}
	return ""
}

// NewUserService creates a new UserService.
func NewUserService(
	uc *biz.UserUseCase,
	publisher message.Publisher,
	hasher hash.Crypto,
	jwtMgr *auth.Manager,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistUC *contentbiz.PlaylistChannelUseCase,
	historyUC *contentbiz.HistoryUseCase,
	mediaClient mediav1.MediaServiceClient,
	logger log.Logger,
) *UserService {
	return &UserService{
		uc:               uc,
		publisher:        publisher,
		hasher:           hasher,
		jwtMgr:           jwtMgr,
		likeFavoriteUC:   likeFavoriteUC,
		playlistUC:       playlistUC,
		historyUC:        historyUC,
		mediaClient:      mediaClient,
		log:              log.NewHelper(log.With(logger, "module", "user.service")),
	}
}

// getUserIDFromContext extracts user ID from context using security principal.
func (s *UserService) getUserIDFromContext(ctx context.Context) (string, error) {
	p, ok := security.FromContext(ctx)
	if ok {
		return p.GetID(), nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			authHeaders = md.Get("grpcgateway-authorization")
		}
		for _, header := range authHeaders {
			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				continue
			}
			claims, err := s.jwtMgr.Parse(token)
			if err != nil {
				continue
			}
			userID := claims.GetUserID()
			if userID != "" {
				return userID, nil
			}
		}
	}
	return "", errors.Unauthorized("UNAUTHENTICATED", "user not authenticated")
}

// ========== Auth Module Methods ==========

// Login authenticates a user and returns JWT tokens.
func (s *UserService) Login(
	ctx context.Context,
	req *userv1.LoginRequest,
) (*userv1.LoginResponse, error) {
	// Find user by username
	userInfo, err := s.uc.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.Unauthorized("INVALID_CREDENTIALS", "Invalid username or password")
		}
		return nil, err
	}

	// Get user entity to access role field
	userEnt, err := s.uc.GetUserEntity(ctx, userInfo.Id)
	if err != nil {
		return nil, err
	}

	if userInfo.Status != types.UserStatus_USER_STATUS_ACTIVE {
		return nil, errors.Forbidden("ACCOUNT_NOT_ACTIVE", "Account is not active")
	}

	// Verify password
	err = s.uc.VerifyPassword(ctx, userInfo.Id, req.GetPassword())
	if err != nil {
		return nil, errors.Unauthorized("INVALID_CREDENTIALS", "Invalid username or password")
	}

	// Generate JWT tokens
	accessToken, err := s.jwtMgr.Generate(userInfo.Id, userInfo.Username, string(userEnt.Role))
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	// Generate new refresh token
	refreshToken, err := s.jwtMgr.GenerateRefreshToken(userInfo.Id, userInfo.Username, string(userEnt.Role))
	if err != nil {
		s.log.Errorf("Failed to generate refresh token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate refresh token")
	}

	// Return login response
	return &userv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtMgr.TTL()).Unix(),
		User:         userInfo,
	}, nil
}

// Logout logs out the user (client-side token removal, server just acknowledges).
func (s *UserService) Logout(
	ctx context.Context,
	req *userv1.LogoutRequest,
) (*userv1.LogoutResponse, error) {
	// For JWT-based auth, logout is handled client-side by removing the token.
	// Server-side token blacklist can be implemented later if needed.
	return &userv1.LogoutResponse{Success: true}, nil
}

// RefreshToken refreshes the access token using a refresh token.
//
// The refresh token is accepted from two sources (in order of priority) to
// accommodate common frontend authentication libraries:
//
//  1. Request body: {"refresh_token": "..."} — the canonical proto contract.
//  2. Authorization header: "Bearer <refresh_token>" — widely used by client
//     libraries that put the current refresh token into the Bearer slot when
//     the access token has expired and the body is left empty (e.g. axios
//     interceptors, NextAuth, OIDC clients, custom fetch wrappers).
//
// If neither source contains a valid refresh token we return 401.
func (s *UserService) RefreshToken(
	ctx context.Context,
	req *userv1.RefreshTokenRequest,
) (*userv1.RefreshTokenResponse, error) {
	refreshToken := resolveRefreshToken(ctx, req.GetRefreshToken())

	// Parse and validate refresh token
	claims, err := s.jwtMgr.Parse(refreshToken)
	if err != nil {
		return nil, errors.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token")
	}

	// Get user info
	userInfo, err := s.uc.GetUser(ctx, claims.GetUserID())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.Unauthorized("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	// Get user entity to access role field
	userEnt, err := s.uc.GetUserEntity(ctx, userInfo.Id)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := s.jwtMgr.Generate(userInfo.Id, userInfo.Username, string(userEnt.Role))
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	// Reuse the outer `refreshToken` variable (already declared at the top of
	// RefreshToken method to hold the resolved incoming token) by using `=`
	// assignment instead of `:=` to avoid "no new variables on left side of :="
	// since `err` was also already declared above.
	refreshToken, err = s.jwtMgr.GenerateRefreshToken(userInfo.Id, userInfo.Username, string(userEnt.Role))
	if err != nil {
		s.log.Errorf("Failed to generate refresh token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate refresh token")
	}

	return &userv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtMgr.TTL()).Unix(),
	}, nil
}

// Register creates a new user account.
func (s *UserService) Register(
	ctx context.Context,
	req *userv1.RegisterRequest,
) (*userv1.RegisterResponse, error) {
	if req.GetUsername() == "" {
		return nil, errors.BadRequest("MISSING_USERNAME", "Username is required")
	}
	if req.GetPassword() == "" {
		return nil, errors.BadRequest("MISSING_PASSWORD", "Password is required")
	}

	existing, _ := s.uc.GetUserByUsername(ctx, req.GetUsername())
	if existing != nil {
		return nil, errors.Conflict("USERNAME_ALREADY_EXISTS", "Username is already taken")
	}

	if req.GetEmail() != "" {
		existingEmail, _ := s.uc.GetUserByEmail(ctx, req.GetEmail())
		if existingEmail != nil {
			return nil, errors.Conflict("EMAIL_ALREADY_EXISTS", "Email is already registered")
		}
	}

	userInfo := &types.User{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Nickname: req.GetNickname(),
		Status:   types.UserStatus_USER_STATUS_ACTIVE,
	}

	count, _ := s.uc.CountUsers(ctx)
	isFirstUser := count == 0

	hashedPassword, err := s.hasher.Hash(req.GetPassword())
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, errors.InternalServer("PASSWORD_HASH_FAILED", "Failed to process password")
	}

	createdUser, err := s.uc.CreateUser(ctx, userInfo, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.Conflict("USER_ALREADY_EXISTS", "Username or email already exists")
		}
		s.log.Errorf("Failed to create user: %v", err)
		return nil, errors.InternalServer("USER_CREATE_FAILED", "Failed to create user")
	}

	userRole := "user"
	if isFirstUser {
		userRole = "admin"
		_ = s.uc.SetUserRole(ctx, createdUser.Id, "admin")
		_ = s.uc.SetUserSuperuser(ctx, createdUser.Id, true)
	}

	// Auto login after registration
	token, err := s.jwtMgr.Generate(createdUser.Id, createdUser.Username, userRole)
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(createdUser.Id, createdUser.Username, userRole)
	if err != nil {
		s.log.Errorf("Failed to generate refresh token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate refresh token")
	}

	return &userv1.RegisterResponse{
		User:         createdUser,
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtMgr.TTL()).Unix(),
	}, nil
}

// ForgotPassword initiates password reset process.
func (s *UserService) ForgotPassword(
	ctx context.Context,
	req *userv1.ForgotPasswordRequest,
) (*userv1.ForgotPasswordResponse, error) {
	// TODO: Implement password reset email sending
	return &userv1.ForgotPasswordResponse{
		Success: true,
		Message: "Password reset initiated. Check your email for instructions.",
	}, nil
}

// ResetPassword resets user password with token.
func (s *UserService) ResetPassword(
	ctx context.Context,
	req *userv1.ResetPasswordRequest,
) (*userv1.ResetPasswordResponse, error) {
	// TODO: Implement password reset with token validation
	return nil, errors.New(501, "RESET_PASSWORD_NOT_IMPLEMENTED", "Reset password not implemented")
}

// GetCurrentUser returns the current authenticated user (deprecated, use GetMe).
func (s *UserService) GetCurrentUser(
	ctx context.Context,
	req *userv1.GetCurrentUserRequest,
) (*userv1.GetCurrentUserResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.uc.GetUser(ctx, userID)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	return &userv1.GetCurrentUserResponse{User: userInfo}, nil
}

// ========== Me Module Methods (Current User) ==========

// GetMe returns the current user's information.
func (s *UserService) GetMe(
	ctx context.Context,
	req *userv1.GetMeRequest,
) (*userv1.GetMeResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.uc.GetUser(ctx, userID)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	return &userv1.GetMeResponse{User: userInfo}, nil
}

// UpdateMe updates the current user's information.
func (s *UserService) UpdateMe(
	ctx context.Context,
	req *userv1.UpdateMeRequest,
) (*userv1.UpdateMeResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get current user
	userInfo, err := s.uc.GetUser(ctx, userID)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	// Update fields if provided
	if req.GetNickname() != "" {
		userInfo.Nickname = req.GetNickname()
	}
	if req.GetEmail() != "" {
		userInfo.Email = req.GetEmail()
	}
	if req.GetAvatar() != "" {
		userInfo.Avatar = req.GetAvatar()
	}
	if req.GetBio() != "" {
		if userInfo.Profile == nil {
			userInfo.Profile = &types.UserProfile{}
		}
		userInfo.Profile.Bio = req.GetBio()
	}

	updatedUser, err := s.uc.UpdateUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	return &userv1.UpdateMeResponse{User: updatedUser}, nil
}

// UpdateMyPassword updates the current user's password.
func (s *UserService) UpdateMyPassword(
	ctx context.Context,
	req *userv1.UpdateMyPasswordRequest,
) (*userv1.UpdateMyPasswordResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Verify old password
	err = s.uc.VerifyPassword(ctx, userID, req.GetCurrentPassword())
	if err != nil {
		return nil, errors.Unauthorized("INVALID_OLD_PASSWORD", "Invalid old password")
	}

	// Hash new password
	hashedPassword, err := s.hasher.Hash(req.GetNewPassword())
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, errors.InternalServer("PASSWORD_HASH_FAILED", "Failed to process password")
	}

	// Update password
	err = s.uc.UpdateUserPassword(ctx, userID, hashedPassword)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	return &userv1.UpdateMyPasswordResponse{Success: true}, nil
}

func (s *UserService) ChangePassword(
	ctx context.Context,
	req *userv1.ChangePasswordRequest,
) (*userv1.ChangePasswordResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.uc.VerifyPassword(ctx, userID, req.GetCurrentPassword())
	if err != nil {
		return nil, errors.Unauthorized("INVALID_OLD_PASSWORD", "Invalid current password")
	}

	hashedPassword, err := s.hasher.Hash(req.GetNewPassword())
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, errors.InternalServer("PASSWORD_HASH_FAILED", "Failed to process password")
	}

	err = s.uc.UpdateUserPassword(ctx, userID, hashedPassword)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	return &userv1.ChangePasswordResponse{Success: true}, nil
}

// playlistStatusToProto maps the persisted playlist status string onto the
// proto enum, defaulting to ACTIVE for legacy rows with an empty status.
func playlistStatusToProto(s string) types.PlaylistStatus {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "INACTIVE":
		return types.PlaylistStatus_PLAYLIST_STATUS_INACTIVE
	case "DRAFT":
		return types.PlaylistStatus_PLAYLIST_STATUS_DRAFT
	case "ARCHIVED":
		return types.PlaylistStatus_PLAYLIST_STATUS_ARCHIVED
	case "ACTIVE", "":
		return types.PlaylistStatus_PLAYLIST_STATUS_ACTIVE
	default:
		return types.PlaylistStatus_PLAYLIST_STATUS_UNSPECIFIED
	}
}

// userPlaylistsToProto converts biz playlists into proto playlists.
//
// BUG-128 root cause B: GetMyPlaylists and GetUserPlaylists each carried a
// duplicated conversion block that hardcoded `MediaCount: 0` and dropped
// Thumbnail/Status, so "My playlists" always rendered "0 videos" even when the
// playlist held items. The denormalized `content_playlist.media_count` column is
// not maintained by AddMedia/RemoveMedia either, so the only trustworthy count at
// read time is the number of join rows the repository already loaded into
// MediaItems - the same source content-side bizPlaylistToProto uses.
func userPlaylistsToProto(playlists []*contentbiz.Playlist) []*types.Playlist {
	out := make([]*types.Playlist, 0, len(playlists))
	for _, p := range playlists {
		if p == nil {
			continue
		}
		privacy := types.Privacy_PRIVACY_PRIVATE
		if p.IsPublic {
			privacy = types.Privacy_PRIVACY_PUBLIC
		}
		out = append(out, &types.Playlist{
			Id:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			ShortToken:  p.ShortToken,
			UserId:      p.UserID,
			Thumbnail:   p.Thumbnail,
			Privacy:     privacy,
			Status:      playlistStatusToProto(p.Status),
			MediaCount:  int64(len(p.MediaItems)),
			CreateTime:  timestamppb.New(p.CreateTime),
			UpdateTime:  timestamppb.New(p.UpdateTime),
		})
	}
	return out
}

// GetMyPlaylists returns the current user's playlists.
func (s *UserService) GetMyPlaylists(
	ctx context.Context,
	req *userv1.GetMyPlaylistsRequest,
) (*userv1.GetMyPlaylistsResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	playlists, total, err := s.playlistUC.ListUserPlaylists(ctx, userID, page, pageSize)
	if err != nil {
		s.log.Errorf("Failed to list user playlists: %v", err)
		return nil, errors.InternalServer("LIST_PLAYLISTS_FAILED", "Failed to list playlists")
	}

	return &userv1.GetMyPlaylistsResponse{
		Items:    userPlaylistsToProto(playlists),
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// GetMyFavorites returns the current user's favorites.
func (s *UserService) GetMyFavorites(
	ctx context.Context,
	req *userv1.GetMyFavoritesRequest,
) (*userv1.GetMyFavoritesResponse, error) {
	_, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	// TODO: Need to get media details for favorites
	// For now, return empty response

	return &userv1.GetMyFavoritesResponse{
		Items:    []*types.Media{},
		Total:    0,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// GetMyLikes returns the current user's likes.
func (s *UserService) GetMyLikes(
	ctx context.Context,
	req *userv1.GetMyLikesRequest,
) (*userv1.GetMyLikesResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	likes, err := s.likeFavoriteUC.ListUserLikes(ctx, userID)
	if err != nil {
		s.log.Errorf("Failed to list user likes: %v", err)
		return nil, errors.InternalServer("LIST_LIKES_FAILED", "Failed to list likes")
	}

	// Convert to proto types
	protoLikes := make([]*types.Like, 0, len(likes))
	for _, l := range likes {
		protoLikes = append(protoLikes, &types.Like{
			Id:         l.ID,
			MediaId:    l.MediaID,
			UserId:     l.UserID,
			LikeType:   l.LikeType,
			CreateTime: timestamppb.New(l.CreateTime),
		})
	}

	return &userv1.GetMyLikesResponse{
		Likes: protoLikes,
		Total: int32(len(protoLikes)),
		Page:  req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// GetMySubscriptions returns the current user's channel subscriptions.
func (s *UserService) GetMySubscriptions(
	ctx context.Context,
	req *userv1.GetMySubscriptionsRequest,
) (*userv1.GetMySubscriptionsResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	subscriptions, total, err := s.uc.GetSubscriptions(ctx, userID, page, pageSize)
	if err != nil {
		s.log.Errorf("Failed to list user subscriptions: %v", err)
		return nil, errors.InternalServer("LIST_SUBSCRIPTIONS_FAILED", "Failed to list subscriptions")
	}

	// BUG-194: subscriptions now carry ShortToken from the channels table, so the
	// sidebar can prefer jumping to /c/{short_token} instead of degrading to /u/{uuid}.
	channels := make([]*types.Channel, 0, len(subscriptions))
	for _, ch := range subscriptions {
		if ch == nil {
			continue
		}
		channels = append(channels, &types.Channel{
			Id:         ch.Id,
			Title:      ch.Title,
			UserId:     ch.UserId,
			Name:       ch.Name,
			Slug:       ch.Slug,
			Avatar:     ch.Avatar,
			ShortToken: ch.ShortToken,
		})
	}

	return &userv1.GetMySubscriptionsResponse{
		Items:    channels,
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// GetMyHistory returns the current user's watch history.
func (s *UserService) GetMyHistory(
	ctx context.Context,
	req *userv1.GetMyHistoryRequest,
) (*userv1.GetMyHistoryResponse, error) {
	_, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	// TODO: Need to map ContentType enum properly
	// For now, return empty response
	return &userv1.GetMyHistoryResponse{
		Items:    []*types.HistoryItem{},
		Total:    0,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// UpsertHistory creates or updates a history record (progress reporting).
func (s *UserService) UpsertHistory(
	ctx context.Context,
	req *userv1.UpsertHistoryRequest,
) (*userv1.UpsertHistoryResponse, error) {
	_, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Need to implement proper mapping
	// For now, return empty response
	return &userv1.UpsertHistoryResponse{
		Item: &types.HistoryItem{},
	}, nil
}

// SyncHistory batch-syncs history records (login merge).
func (s *UserService) SyncHistory(
	ctx context.Context,
	req *userv1.SyncHistoryRequest,
) (*userv1.SyncHistoryResponse, error) {
	_, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Need to implement proper mapping
	// For now, return empty response
	return &userv1.SyncHistoryResponse{
		Items:       []*types.HistoryItem{},
		MergedCount: 0,
	}, nil
}

// ClearHistory clears all watch history for the current user.
func (s *UserService) ClearHistory(
	ctx context.Context,
	req *userv1.ClearHistoryRequest,
) (*userv1.ClearHistoryResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	deletedCount, err := s.historyUC.ClearAll(ctx, userID)
	if err != nil {
		s.log.Errorf("Failed to clear history: %v", err)
		return nil, errors.InternalServer("CLEAR_HISTORY_FAILED", "Failed to clear history")
	}

	return &userv1.ClearHistoryResponse{
		DeletedCount: int32(deletedCount),
	}, nil
}

// RemoveHistoryItem removes a single history record.
func (s *UserService) RemoveHistoryItem(
	ctx context.Context,
	req *userv1.RemoveHistoryItemRequest,
) (*userv1.RemoveHistoryItemResponse, error) {
	_, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.historyUC.Remove(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("Failed to remove history item: %v", err)
		return nil, errors.InternalServer("REMOVE_HISTORY_FAILED", "Failed to remove history item")
	}

	return &userv1.RemoveHistoryItemResponse{}, nil
}

// GetMyStats returns the current user's statistics.
func (s *UserService) GetMyStats(
	ctx context.Context,
	req *userv1.GetMyStatsRequest,
) (*userv1.GetMyStatsResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &userv1.GetMyStatsResponse{
		TotalMedias: s.countUserMedias(ctx, userID),
	}, nil
}

// countUserMedias returns the real number of (public) videos owned by userID,
// sourced from the media service. It is best-effort: any failure is logged and
// treated as 0 so the profile header degrades gracefully instead of erroring.
func (s *UserService) countUserMedias(ctx context.Context, userID string) int32 {
	if s.mediaClient == nil || userID == "" {
		return 0
	}
	resp, err := s.mediaClient.ListMedias(ctx, &mediav1.ListMediasRequest{
		UserId:   &userID,
		PageSize: 1,
	})
	if err != nil {
		s.log.Errorf("countUserMedias: failed to count medias for user %s: %v", userID, err)
		return 0
	}
	return resp.GetTotal()
}

// ========== User Resource Methods ==========

func (s *UserService) ListUsers(
	ctx context.Context,
	req *userv1.ListUsersRequest,
) (*userv1.ListUsersResponse, error) {
	queryOpt := dto.ListUsersRequestToQueryOption(req)
	users, total, err := s.uc.ListUsers(ctx, queryOpt)
	if err != nil {
		return nil, err
	}

	return &userv1.ListUsersResponse{
		Items:    users,
		Total:    total,
		PageSize: queryOpt.PageSize,
		Page:     queryOpt.Page,
	}, nil
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *userv1.GetUserRequest,
) (*userv1.GetUserResponse, error) {
	identifier := req.GetId()
	if identifier == "" {
		return nil, errors.BadRequest("INVALID_IDENTIFIER", "Identifier is required")
	}

	userInfo, err := s.uc.GetUserBySlug(ctx, identifier)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		s.log.Errorf("Failed to get user: %v", err)
		return nil, errors.InternalServer("GET_USER_FAILED", "Failed to get user")
	}
	return &userv1.GetUserResponse{User: userInfo}, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	req *userv1.CreateUserRequest,
) (*userv1.CreateUserResponse, error) {
	opts := dto.CreateUserOptionsFromRequest(req)

	hashedPassword, err := s.hasher.Hash(req.GetPassword())
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, errors.InternalServer("PASSWORD_HASH_FAILED", "Failed to process password")
	}

	userInfo, err := s.uc.CreateUser(ctx, req.GetUser(), hashedPassword, opts)
	if err != nil {
		return nil, err
	}
	return &userv1.CreateUserResponse{User: userInfo}, nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	req *userv1.UpdateUserRequest,
) (*userv1.UpdateUserResponse, error) {
	opts := dto.UpdateUserOptionsFromRequest(req)
	userInfo, err := s.uc.UpdateUser(ctx, req.GetUser(), opts)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	return &userv1.UpdateUserResponse{User: userInfo}, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	req *userv1.DeleteUserRequest,
) (*userv1.DeleteUserResponse, error) {
	err := s.uc.DeleteUser(ctx, req.GetId())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	return &userv1.DeleteUserResponse{}, nil
}

func (s *UserService) UpdateUserStatus(
	ctx context.Context,
	req *userv1.UpdateUserStatusRequest,
) (*userv1.UpdateUserStatusResponse, error) {
	err := s.uc.UpdateUserStatus(ctx, req.GetId(), int8(req.GetStatus()))
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	return &userv1.UpdateUserStatusResponse{}, nil
}

func (s *UserService) UpdateUserRoles(
	ctx context.Context,
	req *userv1.UpdateUserRolesRequest,
) (*userv1.UpdateUserRolesResponse, error) {
	userInfo, err := s.uc.GetUser(ctx, req.GetId())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	
	// TODO: Update roles (requires role management)
	// For now, return the user as is
	return &userv1.UpdateUserRolesResponse{User: userInfo}, nil
}

func (s *UserService) ChangeUserPassword(
	ctx context.Context,
	req *userv1.ChangeUserPasswordRequest,
) (*userv1.ChangeUserPasswordResponse, error) {
	hashedPassword, err := s.hasher.Hash(req.GetPassword())
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, errors.InternalServer("PASSWORD_HASH_FAILED", "Failed to process password")
	}

	err = s.uc.UpdateUserPassword(ctx, req.GetId(), hashedPassword)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	return &userv1.ChangeUserPasswordResponse{}, nil
}

func (s *UserService) VerifyPassword(
	ctx context.Context,
	req *userv1.VerifyPasswordRequest,
) (*userv1.VerifyPasswordResponse, error) {
	var userID string
	
	if req.GetUserId() != "" {
		userID = req.GetUserId()
	} else if req.GetUsername() != "" {
		userInfo, err := s.uc.GetUserByUsername(ctx, req.GetUsername())
		if err != nil {
			if repotypes.IsNotFound(err) {
				return &userv1.VerifyPasswordResponse{Valid: false}, nil
			}
			return nil, err
		}
		userID = userInfo.Id
	} else {
		return nil, errors.BadRequest("MISSING_IDENTIFIER", "Either user_id or username is required")
	}
	
	hashedPassword, err := s.uc.GetUserPasswordHash(ctx, userID)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return &userv1.VerifyPasswordResponse{Valid: false}, nil
		}
		return nil, err
	}

	if validErr := s.hasher.Verify(hashedPassword, req.GetPassword()); validErr != nil {
		return &userv1.VerifyPasswordResponse{Valid: false}, nil
	}
	return &userv1.VerifyPasswordResponse{Valid: true}, nil
}

func (s *UserService) ListUserRoles(
	ctx context.Context,
	req *userv1.ListUserRolesRequest,
) (*userv1.ListUserRolesResponse, error) {
	// TODO: Implement role listing
	return nil, errors.New(501, "LIST_USER_ROLES_NOT_IMPLEMENTED", "List user roles not implemented")
}

func (s *UserService) GetUserStats(
	ctx context.Context,
	req *userv1.GetUserStatsRequest,
) (*userv1.GetUserStatsResponse, error) {
	// BUG-193 G3-C: /users/{id}/stats is reached with slug/username/uuid — resolve
	// to the canonical user id first (previously slug forms returned 0).
	userID, err := s.resolveUserID(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("Failed to resolve stats target: %v", err)
		return nil, errors.InternalServer("RESOLVE_STATS_TARGET_FAILED", "Failed to resolve stats target")
	}
	if userID == "" {
		return &userv1.GetUserStatsResponse{}, nil
	}
	return &userv1.GetUserStatsResponse{
		TotalMedias: s.countUserMedias(ctx, userID),
	}, nil
}

func (s *UserService) GetUserPlaylists(
	ctx context.Context,
	req *userv1.GetUserPlaylistsRequest,
) (*userv1.GetUserPlaylistsResponse, error) {
	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	playlists, total, err := s.playlistUC.ListUserPlaylists(ctx, req.GetId(), page, pageSize)
	if err != nil {
		s.log.Errorf("Failed to list user playlists: %v", err)
		return nil, errors.InternalServer("LIST_PLAYLISTS_FAILED", "Failed to list playlists")
	}

	return &userv1.GetUserPlaylistsResponse{
		Items:    userPlaylistsToProto(playlists),
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

func (s *UserService) GetUserFollowers(
	ctx context.Context,
	req *userv1.GetUserFollowersRequest,
) (*userv1.GetUserFollowersResponse, error) {
	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	followers, total, err := s.uc.GetSubscribers(ctx, req.GetId(), page, pageSize)
	if err != nil {
		s.log.Errorf("Failed to list user followers: %v", err)
		return nil, errors.InternalServer("LIST_FOLLOWERS_FAILED", "Failed to list followers")
	}

	return &userv1.GetUserFollowersResponse{
		Followers: followers,
		Total:     int32(total),
		Page:      req.GetPage(),
		PageSize:  req.GetPageSize(),
	}, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	req *userv1.GetUserByUsernameRequest,
) (*userv1.GetUserByUsernameResponse, error) {
	user, err := s.uc.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		s.log.Errorf("Failed to get user by username: %v", err)
		return nil, errors.InternalServer("GET_USER_FAILED", "Failed to get user")
	}

	return &userv1.GetUserByUsernameResponse{
		User: user,
	}, nil
}

func (s *UserService) SubscribeUser(
	ctx context.Context,
	req *userv1.SubscribeUserRequest,
) (*userv1.SubscribeUserResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	channels, err := s.resolveUserChannels(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("Failed to resolve subscription target: %v", err)
		return nil, errors.InternalServer("RESOLVE_SUBSCRIPTION_TARGET_FAILED", "Failed to resolve subscription target")
	}
	if len(channels) == 0 {
		return nil, errors.NotFound("TARGET_HAS_NO_CHANNEL", "target user has no channel to subscribe to")
	}

	for _, ch := range channels {
		if err := s.uc.Subscribe(ctx, userID, ch.ID); err != nil {
			s.log.Errorf("Failed to subscribe user: %v", err)
			return nil, errors.InternalServer("SUBSCRIBE_FAILED", "Failed to subscribe")
		}
	}

	isSubscribed := false
	for _, ch := range channels {
		subscribed, _ := s.uc.IsSubscribed(ctx, userID, ch.ID)
		if subscribed {
			isSubscribed = true
			break
		}
	}

	return &userv1.SubscribeUserResponse{
		Success:      true,
		IsSubscribed: isSubscribed,
	}, nil
}

func (s *UserService) UnsubscribeUser(
	ctx context.Context,
	req *userv1.UnsubscribeUserRequest,
) (*userv1.UnsubscribeUserResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	channels, err := s.resolveUserChannels(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("Failed to resolve subscription target: %v", err)
		return nil, errors.InternalServer("RESOLVE_SUBSCRIPTION_TARGET_FAILED", "Failed to resolve subscription target")
	}

	for _, ch := range channels {
		if err := s.uc.Unsubscribe(ctx, userID, ch.ID); err != nil {
			s.log.Errorf("Failed to unsubscribe user: %v", err)
			return nil, errors.InternalServer("UNSUBSCRIBE_FAILED", "Failed to unsubscribe")
		}
	}

	isSubscribed := false
	for _, ch := range channels {
		subscribed, _ := s.uc.IsSubscribed(ctx, userID, ch.ID)
		if subscribed {
			isSubscribed = true
			break
		}
	}

	return &userv1.UnsubscribeUserResponse{
		Success:      true,
		IsSubscribed: isSubscribed,
	}, nil
}

// resolveUserChannels resolves a user identifier (slug → username → uuid) to the
// target user's channels. Returns (nil, nil) when the target user does not exist.
// BUG-193 G3-B2 / BUG-194: the subscriptions table stores channel ids, not user
// ids, so every user-level subscription endpoint must resolve the target user's
// channel(s) before reading/writing subscription rows.
func (s *UserService) resolveUserChannels(ctx context.Context, identifier string) ([]*contentbiz.Channel, error) {
	target, err := s.uc.GetUserBySlug(ctx, identifier)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	channels, _, err := s.playlistUC.ListUserChannels(ctx, target.GetId(), 1, 100)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// resolveUserID resolves a portal user identifier (slug → username → uuid) to the
// canonical internal user id. Returns ("", nil) when the user does not exist.
func (s *UserService) resolveUserID(ctx context.Context, identifier string) (string, error) {
	target, err := s.uc.GetUserBySlug(ctx, identifier)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return target.GetId(), nil
}

func (s *UserService) GetUserSubscription(
	ctx context.Context,
	req *userv1.GetUserSubscriptionRequest,
) (*userv1.GetUserSubscriptionResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	channels, err := s.resolveUserChannels(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("Failed to resolve subscription target: %v", err)
		return nil, errors.InternalServer("RESOLVE_SUBSCRIPTION_TARGET_FAILED", "Failed to resolve subscription target")
	}

	// Subscribed to the user iff subscribed to ANY of their channels.
	isSubscribed := false
	for _, ch := range channels {
		subscribed, err := s.uc.IsSubscribed(ctx, userID, ch.ID)
		if err != nil {
			s.log.Errorf("Failed to check subscription: %v", err)
			return nil, errors.InternalServer("CHECK_SUBSCRIPTION_FAILED", "Failed to check subscription")
		}
		if subscribed {
			isSubscribed = true
			break
		}
	}

	return &userv1.GetUserSubscriptionResponse{
		IsSubscribed: isSubscribed,
	}, nil
}

func (s *UserService) GetMyFollowers(
	ctx context.Context,
	req *userv1.GetMyFollowersRequest,
) (*userv1.GetMyFollowersResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	followers, total, err := s.uc.GetSubscribers(ctx, userID, page, pageSize)
	if err != nil {
		s.log.Errorf("Failed to list my followers: %v", err)
		return nil, errors.InternalServer("LIST_MY_FOLLOWERS_FAILED", "Failed to list followers")
	}

	return &userv1.GetMyFollowersResponse{
		Followers: followers,
		Total:     int32(total),
		Page:      req.GetPage(),
		PageSize:  req.GetPageSize(),
	}, nil
}

func (s *UserService) GetMyChannels(
	ctx context.Context,
	req *userv1.GetMyChannelsRequest,
) (*userv1.GetMyChannelsResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	channels, total, err := s.playlistUC.ListUserChannels(ctx, userID, page, pageSize)
	if err != nil {
		s.log.Errorf("Failed to list my channels: %v", err)
		return nil, errors.InternalServer("LIST_MY_CHANNELS_FAILED", "Failed to list channels")
	}

	protoChannels := make([]*types.Channel, 0, len(channels))
	for _, ch := range channels {
		privacy := types.Privacy_PRIVACY_PRIVATE
		if ch.Privacy == "public" {
			privacy = types.Privacy_PRIVACY_PUBLIC
		} else if ch.Privacy == "unlisted" {
			privacy = types.Privacy_PRIVACY_UNLISTED
		}

		status := types.ChannelStatus_CHANNEL_STATUS_ACTIVE
		switch ch.Status {
		case "inactive":
			status = types.ChannelStatus_CHANNEL_STATUS_INACTIVE
		case "suspended":
			status = types.ChannelStatus_CHANNEL_STATUS_SUSPENDED
		case "pending_review":
			status = types.ChannelStatus_CHANNEL_STATUS_PENDING_REVIEW
		}

		protoCh := &types.Channel{
			Id:              ch.ID,
			Title:           ch.Title,
			Name:            ch.Name,
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
			UserId:          ch.UserID,
			CreateTime:      timestamppb.New(ch.CreateTime),
			UpdateTime:      timestamppb.New(ch.UpdateTime),
		}
		if ch.CategoryID != nil {
			protoCh.CategoryId = *ch.CategoryID
		}
		protoChannels = append(protoChannels, protoCh)
	}

	return &userv1.GetMyChannelsResponse{
		Channels: protoChannels,
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

func (s *UserService) RemoveFavorite(
	ctx context.Context,
	req *userv1.RemoveFavoriteRequest,
) (*userv1.RemoveFavoriteResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.likeFavoriteUC.RemoveFavoriteByID(ctx, userID, req.GetId())
	if err != nil {
		s.log.Errorf("Failed to remove favorite: %v", err)
		return nil, errors.InternalServer("REMOVE_FAVORITE_FAILED", "Failed to remove favorite")
	}

	return &userv1.RemoveFavoriteResponse{
		Success: true,
	}, nil
}

// Ensure compile-time interface satisfaction.
var _ userv1.UserServiceServer = (*UserService)(nil)

// Ensure types are used.
var _ = (*types.User)(nil)
