/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package dto is the data transfer object package for the user service.
package dto

import (
	"context"

	"origadmin/application/origcms/api/v1/services/types"
	"origadmin/application/origcms/api/v1/services/user"
	"origadmin/application/origcms/internal/helpers/repo"
)

// UserRepo is a User repository interface.
type UserRepo interface {
	Get(context.Context, int64, ...*UserQueryOption) (*types.User, error)
	List(context.Context, ...*UserQueryOption) ([]*types.User, int32, error)
	Create(context.Context, *types.User, string, ...*UserCreateOption) (*types.User, error)
	Update(context.Context, *types.User, ...*UserUpdateOption) (*types.User, error)
	Delete(context.Context, int64) error
	Restore(context.Context, int64) error

	GetByUsername(context.Context, string) (*types.User, error)
	GetByEmail(context.Context, string) (*types.User, error)
	GetByPhone(context.Context, string) (*types.User, error)
	GetUserAndPassword(context.Context, int64) (*types.User, string, error)

	ChangeUserPassword(context.Context, int64, string) error

	UpdateUserProfile(context.Context, int64, *types.UserProfile) error
	GetUserProfile(context.Context, int64) (*types.UserProfile, error)
	UpdateUserSetting(context.Context, int64, *types.UserSetting) error
	GetUserSetting(context.Context, int64) (*types.UserSetting, error)

	UpdateUserStatus(context.Context, int64, int8) error
}

// UserQueryOption specifies options for querying users.
type UserQueryOption struct {
	repo.QueryOption
	WithProfile bool
	WithSetting bool
}

// UserCreateOption specifies options for creating a user.
type UserCreateOption struct{}

// UserUpdateOption specifies options for updating a user.
type UserUpdateOption struct {
	repo.UpdateOption
}

// GetUserRequestToQueryOption converts a GetUserRequest to a query option object.
func GetUserRequestToQueryOption(req *user.GetUserRequest) *UserQueryOption {
	if req == nil {
		return &UserQueryOption{}
	}
	return &UserQueryOption{
		QueryOption: repo.QueryOption{Page: 1, PageSize: 10},
		WithProfile: req.GetWithProfile(),
		WithSetting: req.GetWithSetting(),
	}
}

// ListUsersRequestToQueryOption converts an API request to a query option object.
func ListUsersRequestToQueryOption(req *user.ListUsersRequest) *UserQueryOption {
	if req == nil {
		return &UserQueryOption{}
	}
	opts := &UserQueryOption{
		QueryOption: repo.QueryOption{
			Page:     req.Page,
			PageSize: req.PageSize,
			Keyword:  req.Keyword,
		},
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	return opts
}

// CreateUserOptionsFromRequest converts a CreateUserRequest to a create option object.
func CreateUserOptionsFromRequest(req *user.CreateUserRequest) *UserCreateOption {
	return &UserCreateOption{}
}

// UpdateUserOptionsFromRequest converts an UpdateUserRequest to an update option object.
func UpdateUserOptionsFromRequest(req *user.UpdateUserRequest) *UserUpdateOption {
	return &UserUpdateOption{}
}
