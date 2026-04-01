/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer implementations.
package data

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/types/known/timestamppb"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/user"
	"origadmin/application/origcms/internal/svc-user/dto"
)

// userRepo implements the dto.UserRepo interface using the shared entity package.
type userRepo struct {
	db *entity.Client
}

// NewUserRepo creates a new User repository.
func NewUserRepo(db *entity.Client) dto.UserRepo {
	return &userRepo{db: db}
}

// Get retrieves a user by ID.
func (r *userRepo) Get(
	ctx context.Context,
	id int64,
	opts ...*dto.UserQueryOption,
) (*types.User, error) {
	u, err := r.db.User.Get(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return convertUserToProto(u), nil
}

// List retrieves a list of users with pagination.
func (r *userRepo) List(
	ctx context.Context,
	opts ...*dto.UserQueryOption,
) ([]*types.User, int32, error) {
	opt := &dto.UserQueryOption{}
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	query := r.db.User.Query()

	if opt.Keyword != "" {
		query = query.Where(
			user.Or(
				user.UsernameContains(opt.Keyword),
				user.NameContains(opt.Keyword),
				user.EmailContains(opt.Keyword),
			),
		)
	}

	// In the unified schema, we use is_active instead of status
	if opt.Status != nil {
		isActive := *opt.Status == 1
		query = query.Where(user.IsActive(isActive))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (opt.Page - 1) * opt.PageSize
	query = query.Offset(int(offset)).Limit(int(opt.PageSize))

	users, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.User, len(users))
	for i, u := range users {
		result[i] = convertUserToProto(u)
	}

	return result, int32(total), nil
}

// Create creates a new user.
func (r *userRepo) Create(
	ctx context.Context,
	in *types.User,
	hashedPassword string,
	opts ...*dto.UserCreateOption,
) (*types.User, error) {
	// 默认值处理
	isActive := in.Status == 1
	if in.Status == 0 {
		isActive = true
	}

	u, err := r.db.User.Create().
		SetUsername(in.Username).
		SetName(in.Nickname).
		SetEmail(in.Email).
		SetPassword(hashedPassword).
		SetIsActive(isActive).
		SetIsStaff(in.IsStaff).
		SetIsSuperuser(in.IsSuperuser).
		SetLogo(in.Avatar).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return convertUserToProto(u), nil
}

// Update updates an existing user.
func (r *userRepo) Update(
	ctx context.Context,
	in *types.User,
	opts ...*dto.UserUpdateOption,
) (*types.User, error) {
	u, err := r.db.User.UpdateOneID(int(in.Id)).
		SetName(in.Nickname).
		SetEmail(in.Email).
		SetIsActive(in.Status == 1).
		SetLogo(in.Avatar).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return convertUserToProto(u), nil
}

// Delete deletes a user by ID.
func (r *userRepo) Delete(ctx context.Context, id int64) error {
	// Hard delete
	return r.db.User.DeleteOneID(int(id)).Exec(ctx)
}

// Restore reactivates a soft-deleted user.
func (r *userRepo) Restore(ctx context.Context, id int64) error {
	return r.db.User.UpdateOneID(int(id)).SetIsActive(true).Exec(ctx)
}

// GetByUsername retrieves a user by username.
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*types.User, error) {
	u, err := r.db.User.Query().Where(user.Username(username)).Only(ctx)
	if err != nil {
		return nil, err
	}
	// DEBUG: 打印从数据库读取的用户完整信息
	slog.Info("user from db raw",
		"id", u.ID,
		"username", u.Username,
		"is_staff", u.IsStaff,
		"is_superuser", u.IsSuperuser,
		"is_active", u.IsActive,
	)
	return convertUserToProto(u), nil
}

// GetByEmail retrieves a user by email.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	u, err := r.db.User.Query().Where(user.Email(email)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return convertUserToProto(u), nil
}

// GetByPhone is not supported in unified schema (email only).
func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*types.User, error) {
	return nil, nil
}

// GetUserAndPassword retrieves a user and their password hash.
func (r *userRepo) GetUserAndPassword(ctx context.Context, id int64) (*types.User, string, error) {
	u, err := r.db.User.Get(ctx, int(id))
	if err != nil {
		return nil, "", err
	}
	return convertUserToProto(u), u.Password, nil
}

// ChangeUserPassword changes a user's password.
func (r *userRepo) ChangeUserPassword(ctx context.Context, userID int64, hashedPassword string) error {
	return r.db.User.UpdateOneID(int(userID)).SetPassword(hashedPassword).Exec(ctx)
}

// UpdateUserProfile updates a user's profile fields directly on the User entity.
func (r *userRepo) UpdateUserProfile(ctx context.Context, userID int64, profile *types.UserProfile) error {
	return r.db.User.UpdateOneID(int(userID)).
		SetName(profile.Name).
		SetDescription(profile.Bio).
		SetLocation(profile.Location).
		SetLogo(profile.Avatar).
		Exec(ctx)
}

// GetUserProfile retrieves a user's profile from the User entity.
func (r *userRepo) GetUserProfile(ctx context.Context, userID int64) (*types.UserProfile, error) {
	u, err := r.db.User.Get(ctx, int(userID))
	if err != nil {
		return nil, err
	}
	return convertUserToProfileProto(u), nil
}

// UpdateUserSetting is a stub (settings could be added to schema later if needed).
func (r *userRepo) UpdateUserSetting(ctx context.Context, userID int64, setting *types.UserSetting) error {
	return nil
}

// GetUserSetting is a stub.
func (r *userRepo) GetUserSetting(ctx context.Context, userID int64) (*types.UserSetting, error) {
	return &types.UserSetting{}, nil
}

// UpdateUserStatus updates a user's status.
func (r *userRepo) UpdateUserStatus(ctx context.Context, userID int64, status int8) error {
	return r.db.User.UpdateOneID(int(userID)).SetIsActive(status == 1).Exec(ctx)
}

// Helper functions

func convertUserToProto(u *entity.User) *types.User {
	status := int32(0)
	if u.IsActive {
		status = 1
	}
	// DEBUG: 打印转换前的值
	slog.Info("convertUserToProto input",
		"is_staff", u.IsStaff,
		"is_superuser", u.IsSuperuser,
	)
	result := &types.User{
		Id:          int64(u.ID),
		Username:    u.Username,
		Nickname:    u.Name,
		Email:       u.Email,
		Status:      status,
		IsStaff:     u.IsStaff,
		IsSuperuser: u.IsSuperuser,
		Avatar:      u.Logo,
		CreateTime:  timestamppb.New(u.DateJoined),
		UpdateTime:  timestamppb.New(u.DateAdded),
	}
	// DEBUG: 打印转换后的值
	slog.Info("convertUserToProto output",
		"is_staff", result.IsStaff,
		"is_superuser", result.IsSuperuser,
	)
	return result
}

func convertUserToProfileProto(u *entity.User) *types.UserProfile {
	return &types.UserProfile{
		Avatar: u.Logo,
		Name:   u.Name,
	}
}
