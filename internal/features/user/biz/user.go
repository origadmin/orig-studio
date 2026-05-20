/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package biz is a biz layer for the user service.
package biz

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	kratosLog "github.com/go-kratos/kratos/v2/log"

	repotypes "origadmin/application/origstudio/internal/domain/types"
	"github.com/origadmin/contrib/security"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	"origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/features/user/dto"
)

// UserUseCase is a User use case.
type UserUseCase struct {
	repo   dto.UserRepo
	hasher hash.Crypto
	log    *kratosLog.Helper
}

// NewUserUseCase new a User use case.
func NewUserUseCase(repo dto.UserRepo, hasher hash.Crypto, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		repo:   repo,
		hasher: hasher,
		log:    log.NewHelper(log.With(logger, "module", "svc-user.biz.user")),
	}
}

func (uc *UserUseCase) UpdateUserStatus(ctx context.Context, id string, status int8) error {
	return uc.repo.UpdateUserStatus(ctx, id, status)
}

// UpdateUserPassword updates a user's password hash directly.
// Note: The password verification and hashing should be handled by the identity service.
func (uc *UserUseCase) UpdateUserPassword(ctx context.Context, userID string, hashedPassword string) error {
	return uc.repo.ChangeUserPassword(ctx, userID, hashedPassword)
}

// GetUserPasswordHash retrieves the encrypted password hash for a user.
func (uc *UserUseCase) GetUserPasswordHash(ctx context.Context, id string) (string, error) {
	_, hash, err := uc.repo.GetUserAndPassword(ctx, id)
	return hash, err
}

func (uc *UserUseCase) ListUsers(ctx context.Context, opts ...*dto.UserQueryOption) ([]*types.User, int32, error) {
	return uc.repo.List(ctx, opts...)
}

// ListUserEntities returns UserEntityDTO list (includes role field not in proto types).
func (uc *UserUseCase) ListUserEntities(ctx context.Context, opts ...*dto.UserQueryOption) ([]*dto.UserEntityDTO, int32, error) {
	return uc.repo.ListEntities(ctx, opts...)
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string, opts ...*dto.UserQueryOption) (*types.User, error) {
	return uc.repo.Get(ctx, id, opts...)
}

// CreateUser creates a new user, ensuring essential fields have valid default values.
// The password argument must be already hashed.
func (uc *UserUseCase) CreateUser(ctx context.Context, in *types.User, hashedPassword string, opts ...*dto.UserCreateOption) (*types.User, error) {
	// Default status to active if not set
	if in.Status == 0 {
		in.Status = 1 // Active
	}

	// Automatically set audit fields from the context
	p, ok := security.FromContext(ctx)
	if ok {
		in.CreateAuthor = p.GetID()
		in.UpdateAuthor = p.GetID()
	}

	return uc.repo.Create(ctx, in, hashedPassword, opts...)
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, in *types.User, opts ...*dto.UserUpdateOption) (*types.User, error) {
	return uc.repo.Update(ctx, in, opts...)
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// GetUserByUsername retrieves a user by username.
func (uc *UserUseCase) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	return uc.repo.GetByUsername(ctx, username)
}

// GetUserBySlug retrieves a user by their public slug.
func (uc *UserUseCase) GetUserBySlug(ctx context.Context, slug string) (*types.User, error) {
	return uc.repo.GetBySlug(ctx, slug)
}

// UpdateUserSlug updates a user's slug with validation.
func (uc *UserUseCase) UpdateUserSlug(ctx context.Context, userID string, slug string) error {
	// Validate slug is not empty
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	// Validate charset: only [a-z0-9-_], length 3-64
	if !isValidSlug(slug) {
		return fmt.Errorf("slug must match [a-z0-9-_], length 3-64")
	}
	// Validate not reserved
	if isReservedSlug(slug) {
		return fmt.Errorf("slug is reserved")
	}
	// Get user to check username
	u, err := uc.repo.Get(ctx, userID)
	if err != nil {
		return err
	}
	// Validate slug != username (case-insensitive)
	if strings.EqualFold(slug, u.Username) {
		return fmt.Errorf("slug cannot be the same as username")
	}
	return uc.repo.UpdateSlug(ctx, userID, slug)
}

func isValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 64 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$`, slug)
	return matched
}

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

// VerifyPassword checks whether the plain-text password matches the stored hash.
func (uc *UserUseCase) VerifyPassword(ctx context.Context, userID string, plainPassword string) error {
	_, hashedPassword, err := uc.repo.GetUserAndPassword(ctx, userID)
	if err != nil {
		return err
	}
	return uc.hasher.Verify(hashedPassword, plainPassword)
}

// HashPassword hashes a plain-text password.
func (uc *UserUseCase) HashPassword(plainPassword string) (string, error) {
	return uc.hasher.Hash(plainPassword)
}

// GetUserEntity returns the UserEntityDTO (for fields not in proto types, e.g. role).
func (uc *UserUseCase) GetUserEntity(ctx context.Context, id string) (*dto.UserEntityDTO, error) {
	return uc.repo.GetEntity(ctx, id)
}

// SetUserRole updates a user's role.
func (uc *UserUseCase) SetUserRole(ctx context.Context, id string, role string) error {
	return uc.repo.SetUserRole(ctx, id, role)
}

func (uc *UserUseCase) CountUsers(ctx context.Context) (int, error) {
	// For now, list with 0 limit to get count if List returns total
	_, total, err := uc.repo.List(ctx, &dto.UserQueryOption{QueryOption: repotypes.QueryOption{PageSize: 1}})
	return int(total), err
}

// ==================== Subscription Methods ====================

// IsSubscribed checks if a user is subscribed to a channel
func (uc *UserUseCase) IsSubscribed(ctx context.Context, subscriberID, channelID string) (bool, error) {
	return uc.repo.IsSubscribed(ctx, subscriberID, channelID)
}

// GetSubscriberCount gets the number of subscribers for a channel
func (uc *UserUseCase) GetSubscriberCount(ctx context.Context, channelID string) (int, error) {
	return uc.repo.GetSubscriberCount(ctx, channelID)
}

// Subscribe adds a subscription
func (uc *UserUseCase) Subscribe(ctx context.Context, subscriberID, channelID string) error {
	if subscriberID == channelID {
		return nil // Can't subscribe to yourself
	}
	return uc.repo.Subscribe(ctx, subscriberID, channelID)
}

// Unsubscribe removes a subscription
func (uc *UserUseCase) Unsubscribe(ctx context.Context, subscriberID, channelID string) error {
	return uc.repo.Unsubscribe(ctx, subscriberID, channelID)
}

// GetSubscriptions gets all channels a user is subscribed to
func (uc *UserUseCase) GetSubscriptions(ctx context.Context, subscriberID string, page, pageSize int) ([]*types.User, int, error) {
	return uc.repo.GetSubscriptions(ctx, subscriberID, page, pageSize)
}

// GetSubscribers gets all subscribers for a channel
func (uc *UserUseCase) GetSubscribers(ctx context.Context, channelID string, page, pageSize int) ([]*types.User, int, error) {
	return uc.repo.GetSubscribers(ctx, channelID, page, pageSize)
}
