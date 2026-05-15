/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer implementations.
package dal

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/data/convpb"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/subscription"
	"origadmin/application/origstudio/internal/data/entity/user"
	"origadmin/application/origstudio/internal/helpers/idutil"
	"origadmin/application/origstudio/internal/features/user/dto"
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

	// In the unified schema, we use status enum instead of is_active
	if opt.Status != nil {
		status := userStatusFromInt32(*opt.Status)
		query = query.Where(user.StatusEQ(status))
	}

	// Filter by role if specified
	if opt.Role != "" {
		query = query.Where(user.RoleEQ(user.Role(opt.Role)))
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

	// Default sort: create_time DESC (newest first) to ensure stable pagination
	query = query.Order(entity.Desc(user.FieldCreateTime))

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

// ListEntities returns raw entity.User list (includes role field not in proto types).
func (r *userRepo) ListEntities(
	ctx context.Context,
	opts ...*dto.UserQueryOption,
) ([]*entity.User, int32, error) {
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

	if opt.Status != nil {
		status := userStatusFromInt32(*opt.Status)
		query = query.Where(user.StatusEQ(status))
	}

	// Filter by role if specified
	if opt.Role != "" {
		query = query.Where(user.RoleEQ(user.Role(opt.Role)))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	page := opt.Page
	pageSize := opt.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	query = query.Offset(int(offset)).Limit(int(pageSize))

	// Default sort: create_time DESC (newest first) to ensure stable pagination
	query = query.Order(entity.Desc(user.FieldCreateTime))

	users, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, int32(total), nil
}

// Create creates a new user.
// Per A009: registration does NOT auto-create a default channel.
// Channels are created on-demand when the user wants to publish content.
func (r *userRepo) Create(
	ctx context.Context,
	in *types.User,
	hashedPassword string,
	opts ...*dto.UserCreateOption,
) (*types.User, error) {
	// Default value handling
	status := user.StatusACTIVE
	if in.Status == types.UserStatus_USER_STATUS_INACTIVE {
		status = user.StatusINACTIVE
	}

	// Auto-generate slug from name/username
	slug, err := r.GenerateSlug(ctx, in.Nickname, in.Username)
	if err != nil {
		// Fallback: prefix username with 'u-'
		slug = "u-" + strings.ToLower(in.Username)
	}

	u, err := r.db.User.Create().
		SetID(idutil.GenUUID()).
		SetUsername(in.Username).
		SetName(in.Nickname).
		SetEmail(in.Email).
		SetPassword(hashedPassword).
		SetSlug(slug).
		SetStatus(status).
		SetIsStaff(in.IsStaff).
		SetIsSuperuser(in.IsSuperuser).
		SetLogo(in.Avatar).
		SetRole("user").
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
	u, err := r.db.User.UpdateOneID(in.Id).
		SetName(in.Nickname).
		SetEmail(in.Email).
		SetStatus(userStatusFromProto(in.Status)).
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
	return r.db.User.UpdateOneID(id).SetStatus(user.StatusACTIVE).Exec(ctx)
}

// GetByUsername retrieves a user by username.
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*types.User, error) {
	u, err := r.db.User.Query().Where(user.Username(username)).Only(ctx)
	if err != nil {
		return nil, err
	}
	// DEBUG: print user info from database
	slog.Info("user from db raw",
		"id", u.ID,
		"username", u.Username,
		"is_staff", u.IsStaff,
		"is_superuser", u.IsSuperuser,
		"status", u.Status,
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

// GetBySlug retrieves a user by their public slug.
func (r *userRepo) GetBySlug(ctx context.Context, slug string) (*types.User, error) {
	u, err := r.db.User.Query().Where(user.Slug(slug)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return convertUserToProto(u), nil
}

// GenerateSlug generates a unique slug from name/username.
// Ensures slug != username (case-insensitive) and excludes reserved words.
func (r *userRepo) GenerateSlug(ctx context.Context, name, username string) (string, error) {
	base := generateSlugFromName(name, username)
	slug := base
	suffix := 1
	for {
		exists, err := r.db.User.Query().Where(user.Slug(slug)).Exist(ctx)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		suffix++
		slug = fmt.Sprintf("%s-%d", base, suffix)
	}
}

// UpdateSlug updates a user's slug.
func (r *userRepo) UpdateSlug(ctx context.Context, userID string, slug string) error {
	return r.db.User.UpdateOneID(userID).SetSlug(slug).Exec(ctx)
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
// Maps proto UserStatus enum values to Ent entity Status values:
// 0=UNSPECIFIED→PENDING, 1=PENDING, 2=ACTIVE, 3=INACTIVE, 4=SUSPENDED, 5=REJECTED
func (r *userRepo) UpdateUserStatus(ctx context.Context, userID string, status int8) error {
	var entityStatus user.Status
	switch status {
	case 0, 1:
		entityStatus = user.StatusPENDING
	case 2:
		entityStatus = user.StatusACTIVE
	case 3:
		entityStatus = user.StatusINACTIVE
	case 4:
		entityStatus = user.StatusSUSPENDED
	case 5:
		entityStatus = user.StatusREJECTED
	default:
		// Unknown status, default to PENDING for safety
		entityStatus = user.StatusPENDING
	}
	return r.db.User.UpdateOneID(userID).SetStatus(entityStatus).Exec(ctx)
}

// Helper functions

func convertUserToProto(u *entity.User) *types.User {
	// Use the auto-generated convpb converter which includes all field mappings
	// (create_author, update_author, create_time, update_time, nickname, phone, avatar, etc.)
	// The previous manual implementation was missing many fields.
	return convpb.ConvertUserToUserPBFull(u)
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

// userStatusFromInt32 converts int32 (dto layer) to user.Status (entity enum string).
func userStatusFromInt32(status int32) user.Status {
	switch types.UserStatus(status) {
	case types.UserStatus_USER_STATUS_ACTIVE:
		return user.StatusACTIVE
	case types.UserStatus_USER_STATUS_INACTIVE:
		return user.StatusINACTIVE
	case types.UserStatus_USER_STATUS_PENDING:
		return user.StatusPENDING
	case types.UserStatus_USER_STATUS_SUSPENDED:
		return user.StatusSUSPENDED
	default:
		return user.StatusACTIVE
	}
}

func userStatusFromProto(status types.UserStatus) user.Status {
	switch status {
	case types.UserStatus_USER_STATUS_ACTIVE:
		return user.StatusACTIVE
	case types.UserStatus_USER_STATUS_INACTIVE:
		return user.StatusINACTIVE
	case types.UserStatus_USER_STATUS_PENDING:
		return user.StatusPENDING
	case types.UserStatus_USER_STATUS_SUSPENDED:
		return user.StatusSUSPENDED
	default:
		return user.StatusACTIVE
	}
}

func userStatusToProto(status user.Status) types.UserStatus {
	switch status {
	case user.StatusACTIVE:
		return types.UserStatus_USER_STATUS_ACTIVE
	case user.StatusINACTIVE:
		return types.UserStatus_USER_STATUS_INACTIVE
	case user.StatusPENDING:
		return types.UserStatus_USER_STATUS_PENDING
	case user.StatusSUSPENDED:
		return types.UserStatus_USER_STATUS_SUSPENDED
	case user.StatusREJECTED:
		return types.UserStatus_USER_STATUS_REJECTED
	default:
		return types.UserStatus_USER_STATUS_ACTIVE
	}
}

// generateSlugFromName generates a base slug from name/username.
// Ensures slug != username (case-insensitive) and excludes reserved words.
func generateSlugFromName(name, username string) string {
	// Try name-based slug first
	if name != "" {
		slug := sanitizeForSlug(name)
		if slug != "" && len(slug) >= 3 && !strings.EqualFold(slug, username) && !isReservedSlug(slug) {
			return slug
		}
	}
	// Fallback: prefix username with 'u-'
	return "u-" + strings.ToLower(username)
}

// sanitizeForSlug converts a string to a URL-friendly slug.
func sanitizeForSlug(s string) string {
	slug := strings.ToLower(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-"))
	slug = strings.Trim(slug, "-")
	return slug
}

// isReservedSlug checks if a slug is a reserved word.
func isReservedSlug(slug string) bool {
	reserved := map[string]bool{
		"admin": true, "support": true, "help": true, "api": true,
		"www": true, "mail": true, "ftp": true, "root": true,
		"system": true, "moderator": true, "mod": true, "staff": true,
		"user": true, "users": true, "channel": true, "channels": true,
		"media": true, "video": true, "playlist": true, "category": true,
		"tag": true, "tags": true, "article": true, "search": true,
		"explore": true, "settings": true, "auth": true, "login": true,
		"register": true, "signup": true, "signin": true, "signout": true,
	}
	return reserved[strings.ToLower(slug)]
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
