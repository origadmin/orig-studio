/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service is the service layer for the user service.
package service

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	"origadmin/application/origcms/api/gen/v1/types"
	userv1 "origadmin/application/origcms/api/gen/v1/user"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/dto"
)

// UserService implements the user gRPC service.
type UserService struct {
	userv1.UnimplementedUserServiceServer
	uc        *biz.UserUseCase
	publisher message.Publisher
	hasher    hash.Crypto
	jwtMgr    *auth.Manager
	log       *log.Helper
}

// NewUserService creates a new UserService.
func NewUserService(
	uc *biz.UserUseCase,
	publisher message.Publisher,
	hasher hash.Crypto,
	jwtMgr *auth.Manager,
	logger log.Logger,
) *UserService {
	return &UserService{
		uc:        uc,
		publisher: publisher,
		hasher:    hasher,
		jwtMgr:    jwtMgr,
		log:       log.NewHelper(log.With(logger, "module", "user.service")),
	}
}

func (s *UserService) UpdateUserStatus(
	ctx context.Context,
	req *userv1.UpdateUserStatusRequest,
) (*userv1.UpdateUserStatusResponse, error) {
	err := s.uc.UpdateUserStatus(ctx, req.GetId(), int8(req.GetStatus()))
	if err != nil {
		if entity.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	return &userv1.UpdateUserStatusResponse{}, nil
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
		if entity.IsNotFound(err) {
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
	hashedPassword, err := s.uc.GetUserPasswordHash(ctx, req.GetUserId())
	if err != nil {
		if entity.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}

	if validErr := s.hasher.Verify(hashedPassword, req.GetPassword()); validErr != nil {
		return nil, errors.BadRequest("INVALID_PASSWORD", "Invalid password")
	}
	return &userv1.VerifyPasswordResponse{Valid: true}, nil
}

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
		Users:    users,
		Total:    total,
		PageSize: req.PageSize,
		Page:     req.Page,
	}, nil
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *userv1.GetUserRequest,
) (*userv1.GetUserResponse, error) {
	queryOpt := dto.GetUserRequestToQueryOption(req)
	userInfo, err := s.uc.GetUser(ctx, req.GetId(), queryOpt)
	if err != nil {
		if entity.IsNotFound(err) {
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
		if entity.IsNotFound(err) {
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
		if entity.IsNotFound(err) {
			return nil, errors.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, err
	}
	return &userv1.DeleteUserResponse{}, nil
}

// Login authenticates a user and returns JWT tokens.
func (s *UserService) Login(
	ctx context.Context,
	req *userv1.LoginRequest,
) (*userv1.LoginResponse, error) {
	// Find user by username
	userInfo, err := s.uc.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		if entity.IsNotFound(err) {
			return nil, errors.Unauthorized("INVALID_CREDENTIALS", "Invalid username or password")
		}
		return nil, err
	}

	// Verify password
	err = s.uc.VerifyPassword(ctx, userInfo.Id, req.GetPassword())
	if err != nil {
		return nil, errors.Unauthorized("INVALID_CREDENTIALS", "Invalid username or password")
	}

	// Generate JWT token
	token, err := s.jwtMgr.Generate(userInfo.Id, userInfo.Username, userInfo.IsStaff)
	if err != nil {
		s.log.Errorf("Failed to generate JWT token: %v", err)
		return nil, errors.InternalServer("TOKEN_GENERATION_FAILED", "Failed to generate token")
	}

	// Return login response
	return &userv1.LoginResponse{
		AccessToken:  token,
		RefreshToken: "", // TODO: implement refresh token
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

// Ensure compile-time interface satisfaction.
var _ userv1.UserServiceServer = (*UserService)(nil)

// Ensure types are used.
var _ = (*types.User)(nil)
