/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer implementations.
package data

import (
	"context"
	"log/slog"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/subscription"
	"origadmin/application/origcms/internal/data/entity/user"
	"origadmin/application/origcms/internal/helpers/idutil"
	"origadmin/application/origcms/internal/svc-user/dto"

	"google.golang.org/protobuf/types/known/timestamppb"
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
	id string,
	opts ...*dto.UserQueryOption,
) (*types.User, error) {
	u, err := r.db.User.Get(ctx, id)
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

	// Set default pagination values if not provided
	page := opt.Page
	pageSize := opt.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize
	query = query.Offset(int(offset)).Limit(int(pageSize))

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

// Create creates a new user and automatically creates a default channel.
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
		SetID(idutil.GenUUID()).
		SetUsername(in.Username).
		SetHandle("@" + in.Username).
		SetName(in.Nickname).
		SetEmail(in.Email).
		SetPassword(hashedPassword).
		SetIsActive(isActive).
		SetIsStaff(in.IsStaff).
		SetIsSuperuser(in.IsSuperuser).
		SetLogo(in.Avatar).
		SetRole("user").
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 为新用户创建默认频道
	_, err = r.db.Channel.Create().
		SetID(idutil.GenUUID()).
		SetUserID(u.ID).
		SetTitle(u.Username + "'s Channel").
		SetSlug(u.Username).
		SetDescription("Default channel for " + u.Username).
		SetIsPublic(true).
		Save(ctx)
	if err != nil {
		// 记录错误但不影响用户创建
		slog.Error("Failed to create default channel for user", "user_id", u.ID, "error", err)
	}

	return convertUserToProto(u), nil
}

// Update updates an existing user.
func (r *userRepo) Update(
	ctx context.Context,
	in *types.User,
	opts ...*dto.UserUpdateOption,
) (*types.User, error) {
	u, err := r.db.User.UpdateOneID(in.Uuid).
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
func (r *userRepo) Delete(ctx context.Context, id string) error {
	// Hard delete
	return r.db.User.DeleteOneID(id).Exec(ctx)
}

// Restore reactivates a soft-deleted user.
func (r *userRepo) Restore(ctx context.Context, id string) error {
	return r.db.User.UpdateOneID(id).SetIsActive(true).Exec(ctx)
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

// GetByHandle retrieves a user by handle.
func (r *userRepo) GetByHandle(ctx context.Context, handle string) (*types.User, error) {
	u, err := r.db.User.Query().Where(user.Handle(handle)).Only(ctx)
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
func (r *userRepo) GetUserAndPassword(ctx context.Context, id string) (*types.User, string, error) {
	u, err := r.db.User.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}
	return convertUserToProto(u), u.Password, nil
}

// ChangeUserPassword changes a user's password.
func (r *userRepo) ChangeUserPassword(ctx context.Context, userID string, hashedPassword string) error {
	return r.db.User.UpdateOneID(userID).SetPassword(hashedPassword).Exec(ctx)
}

// UpdateUserProfile updates a user's profile fields directly on the User entity.
func (r *userRepo) UpdateUserProfile(ctx context.Context, userID string, profile *types.UserProfile) error {
	return r.db.User.UpdateOneID(userID).
		SetName(profile.Name).
		SetDescription(profile.Bio).
		SetLocation(profile.Location).
		SetLogo(profile.Avatar).
		Exec(ctx)
}

// GetUserProfile retrieves a user's profile from the User entity.
func (r *userRepo) GetUserProfile(ctx context.Context, userID string) (*types.UserProfile, error) {
	u, err := r.db.User.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return convertUserToProfileProto(u), nil
}

// UpdateUserSetting is a stub (settings could be added to schema later if needed).
func (r *userRepo) UpdateUserSetting(ctx context.Context, userID string, setting *types.UserSetting) error {
	return nil
}

// GetUserSetting is a stub.
func (r *userRepo) GetUserSetting(ctx context.Context, userID string) (*types.UserSetting, error) {
	return &types.UserSetting{}, nil
}

// UpdateUserStatus updates a user's status.
func (r *userRepo) UpdateUserStatus(ctx context.Context, userID string, status int8) error {
	return r.db.User.UpdateOneID(userID).SetIsActive(status == 1).Exec(ctx)
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
		Uuid:        u.ID,
		Username:    u.Username,
		Handle:      u.Handle,
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

// GetEntity returns the raw ent entity.User (for fields not in proto types, e.g. role).
func (r *userRepo) GetEntity(ctx context.Context, id string) (*entity.User, error) {
	return r.db.User.Get(ctx, id)
}

// SetUserRole updates a user's role field directly.
func (r *userRepo) SetUserRole(ctx context.Context, id string, role string) error {
	return r.db.User.UpdateOneID(id).SetRole(user.Role(role)).Exec(ctx)
}

// ==================== Subscription Methods ====================

// IsSubscribed checks if a user is subscribed to a channel
func (r *userRepo) IsSubscribed(ctx context.Context, subscriberID, channelID string) (bool, error) {
	count, err := r.db.Subscription.Query().
		Where(
			subscription.SubscriberID(subscriberID),
			subscription.ChannelID(channelID),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetSubscriberCount gets the number of subscribers for a channel
func (r *userRepo) GetSubscriberCount(ctx context.Context, channelID string) (int, error) {
	count, err := r.db.Subscription.Query().
		Where(subscription.ChannelID(channelID)).
		Count(ctx)
	return count, err
}

// Subscribe adds a subscription
func (r *userRepo) Subscribe(ctx context.Context, subscriberID, channelID string) error {
	// Check if already subscribed
	exists, err := r.IsSubscribed(ctx, subscriberID, channelID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already subscribed, no error
	}

	_, err = r.db.Subscription.Create().
		SetSubscriberID(subscriberID).
		SetChannelID(channelID).
		Save(ctx)
	return err
}

// Unsubscribe removes a subscription
func (r *userRepo) Unsubscribe(ctx context.Context, subscriberID, channelID string) error {
	// Check if subscription exists
	exists, err := r.IsSubscribed(ctx, subscriberID, channelID)
	if err != nil {
		return err
	}
	if !exists {
		return nil // Not subscribed, nothing to do
	}

	_, err = r.db.Subscription.Delete().
		Where(
			subscription.SubscriberID(subscriberID),
			subscription.ChannelID(channelID),
		).
		Exec(ctx)
	return err
}

// GetSubscriptions gets all channels a user is subscribed to
func (r *userRepo) GetSubscriptions(ctx context.Context, subscriberID string, page, pageSize int) ([]*types.User, int, error) {
	subs, err := r.db.Subscription.Query().
		Where(subscription.SubscriberID(subscriberID)).
		WithChannel().
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.db.Subscription.Query().
		Where(subscription.SubscriberID(subscriberID)).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.User, len(subs))
	for i, sub := range subs {
		if channel := sub.Edges.Channel; channel != nil {
			result[i] = convertUserToProto(channel)
		}
	}

	return result, total, nil
}

// GetSubscribers gets all subscribers for a channel
func (r *userRepo) GetSubscribers(ctx context.Context, channelID string, page, pageSize int) ([]*types.User, int, error) {
	subs, err := r.db.Subscription.Query().
		Where(subscription.ChannelID(channelID)).
		WithSubscriber().
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.db.Subscription.Query().
		Where(subscription.ChannelID(channelID)).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.User, len(subs))
	for i, sub := range subs {
		if subscriber := sub.Edges.Subscriber; subscriber != nil {
			result[i] = convertUserToProto(subscriber)
		}
	}

	return result, total, nil
}
