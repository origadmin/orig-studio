/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service is the service layer for the user service.
package service

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/origadmin/contrib/security"
	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	"google.golang.org/protobuf/types/known/timestamppb"
	"origadmin/application/origstudio/api/gen/v1/types"
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
	log              *log.Helper
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
		log:              log.NewHelper(log.With(logger, "module", "user.service")),
	}
}

// getUserIDFromContext extracts user ID from context using security principal.
func (s *UserService) getUserIDFromContext(ctx context.Context) (string, error) {
	p, ok := security.FromContext(ctx)
	if !ok {
		return "", errors.Unauthorized("UNAUTHENTICATED", "user not authenticated")
	}
	return p.GetID(), nil
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

	// Verify password
	err = s.uc.VerifyPassword(ctx, userInfo.Id, req.GetPassword())
	if err != nil {
		return nil, errors.Unauthorized("INVALID_CREDENTIALS", "Invalid username or password")
	}

	// Generate JWT tokens
	accessToken, err := s.jwtMgr.Generate(userInfo.Id, userInfo.Username, userInfo.IsStaff, string(userEnt.Role))
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(userInfo.Id, userInfo.Username, userInfo.IsStaff, string(userEnt.Role))
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
func (s *UserService) RefreshToken(
	ctx context.Context,
	req *userv1.RefreshTokenRequest,
) (*userv1.RefreshTokenResponse, error) {
	// Parse and validate refresh token
	claims, err := s.jwtMgr.Parse(req.GetRefreshToken())
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
	accessToken, err := s.jwtMgr.Generate(userInfo.Id, userInfo.Username, userInfo.IsStaff, string(userEnt.Role))
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	// Generate new refresh token
	refreshToken, err := s.jwtMgr.GenerateRefreshToken(userInfo.Id, userInfo.Username, userInfo.IsStaff, string(userEnt.Role))
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
	// Create user object
	userInfo := &types.User{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Nickname: req.GetNickname(),
		Status:   types.UserStatus_USER_STATUS_ACTIVE,
	}

	hashedPassword, err := s.hasher.Hash(req.GetPassword())
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, errors.InternalServer("PASSWORD_HASH_FAILED", "Failed to process password")
	}

	createdUser, err := s.uc.CreateUser(ctx, userInfo, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Auto login after registration
	token, err := s.jwtMgr.Generate(createdUser.Id, createdUser.Username, createdUser.IsStaff, "user")
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(createdUser.Id, createdUser.Username, createdUser.IsStaff, "user")
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

	// Convert to proto types
	protoPlaylists := make([]*types.Playlist, 0, len(playlists))
	for _, p := range playlists {
		privacy := types.Privacy_PRIVACY_PRIVATE
		if p.IsPublic {
			privacy = types.Privacy_PRIVACY_PUBLIC
		}
		protoPlaylists = append(protoPlaylists, &types.Playlist{
			Id:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			ShortToken:  p.ShortToken,
			UserId:      p.UserID,
			Privacy:     privacy,
			MediaCount:  0,
			CreateTime:  timestamppb.New(p.CreateTime),
			UpdateTime:  timestamppb.New(p.UpdateTime),
		})
	}

	return &userv1.GetMyPlaylistsResponse{
		Items:    protoPlaylists,
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
			Type:       l.LikeType,
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

	channels := make([]*types.Channel, 0, len(subscriptions))
	for _, u := range subscriptions {
		channels = append(channels, &types.Channel{
			Id:      u.Id,
			Title:   u.Nickname,
			UserId:  u.Id,
			Name:    u.Username,
			Slug:    u.Slug,
			Avatar:  u.Avatar,
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
	_, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Implement proper stats calculation
	// For now, return placeholder values
	return &userv1.GetMyStatsResponse{
		TotalViews:     0,
		TotalLikes:     0,
		TotalMedias:    0,
		TotalFollowers: 0,
		TotalFollowing: 0,
	}, nil
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
	queryOpt := dto.GetUserRequestToQueryOption(req)
	userInfo, err := s.uc.GetUser(ctx, req.GetId(), queryOpt)
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
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
	// TODO: Implement user stats calculation
	return &userv1.GetUserStatsResponse{
		TotalViews:     0,
		TotalLikes:     0,
		TotalMedias:    0,
		TotalFollowers: 0,
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

	// Convert to proto types
	protoPlaylists := make([]*types.Playlist, 0, len(playlists))
	for _, p := range playlists {
		privacy := types.Privacy_PRIVACY_PRIVATE
		if p.IsPublic {
			privacy = types.Privacy_PRIVACY_PUBLIC
		}
		protoPlaylists = append(protoPlaylists, &types.Playlist{
			Id:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			ShortToken:  p.ShortToken,
			UserId:      p.UserID,
			Privacy:     privacy,
			MediaCount:  0,
			CreateTime:  timestamppb.New(p.CreateTime),
			UpdateTime:  timestamppb.New(p.UpdateTime),
		})
	}

	return &userv1.GetUserPlaylistsResponse{
		Items:    protoPlaylists,
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

// Ensure compile-time interface satisfaction.
var _ userv1.UserServiceServer = (*UserService)(nil)

// Ensure types are used.
var _ = (*types.User)(nil)
