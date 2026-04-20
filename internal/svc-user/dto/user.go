/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package dto is the data transfer object package for the user service.
package dto

import (
	"context"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/api/gen/v1/user"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/helpers/repo"
)

// UserRepo is a User repository interface.
type UserRepo interface {
	Get(context.Context, string, ...*UserQueryOption) (*types.User, error)
	List(context.Context, ...*UserQueryOption) ([]*types.User, int32, error)
	Create(context.Context, *types.User, string, ...*UserCreateOption) (*types.User, error)
	Update(context.Context, *types.User, ...*UserUpdateOption) (*types.User, error)
	Delete(context.Context, string) error
	Restore(context.Context, string) error

	GetByUsername(context.Context, string) (*types.User, error)
	GetByEmail(context.Context, string) (*types.User, error)
	GetByHandle(context.Context, string) (*types.User, error)
	GetByPhone(context.Context, string) (*types.User, error)
	GetUserAndPassword(context.Context, string) (*types.User, string, error)

	ChangeUserPassword(context.Context, string, string) error

	UpdateUserProfile(context.Context, string, *types.UserProfile) error
	GetUserProfile(context.Context, string) (*types.UserProfile, error)
	UpdateUserSetting(context.Context, string, *types.UserSetting) error
	GetUserSetting(context.Context, string) (*types.UserSetting, error)

	UpdateUserStatus(context.Context, string, int8) error

	// GetEntity returns the raw ent entity (for fields not in proto types, e.g. role).
	GetEntity(context.Context, string) (*entity.User, error)

	// SetUserRole updates a user's role directly via ent.
	SetUserRole(context.Context, string, string) error

	// Subscription methods
	IsSubscribed(ctx context.Context, subscriberID, channelID string) (bool, error)
	GetSubscriberCount(ctx context.Context, channelID string) (int, error)
	Subscribe(ctx context.Context, subscriberID, channelID string) error
	Unsubscribe(ctx context.Context, subscriberID, channelID string) error
	GetSubscriptions(ctx context.Context, subscriberID string, page, pageSize int) ([]*types.User, int, error)
	GetSubscribers(ctx context.Context, channelID string, page, pageSize int) ([]*types.User, int, error)
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
