package identity

import (
	"context"
	"strings"

	"github.com/origadmin/runtime/log"
)

type UserIdentityRepo interface {
	UsernameExists(ctx context.Context, username string) (bool, error)
	SlugExists(ctx context.Context, slug string, excludeUserID string) (bool, error)
}

type ChannelIdentityRepo interface {
	HandleExists(ctx context.Context, handle string, excludeChannelID string) (bool, error)
}

type Checker struct {
	userRepo    UserIdentityRepo
	channelRepo ChannelIdentityRepo
}

func NewChecker(userRepo UserIdentityRepo, channelRepo ChannelIdentityRepo) *Checker {
	return &Checker{
		userRepo:    userRepo,
		channelRepo: channelRepo,
	}
}

type AvailabilityResult struct {
	Available bool
	Conflict  string
}

func (c *Checker) IsIdentifierAvailable(ctx context.Context, identifier string, excludeUserID string) (*AvailabilityResult, error) {
	identifier = strings.ToLower(identifier)

	usernameTaken, err := c.userRepo.UsernameExists(ctx, identifier)
	if err != nil {
		return nil, err
	}
	if usernameTaken {
		return &AvailabilityResult{Available: false, Conflict: "username"}, nil
	}

	slugTaken, err := c.userRepo.SlugExists(ctx, identifier, excludeUserID)
	if err != nil {
		return nil, err
	}
	if slugTaken {
		return &AvailabilityResult{Available: false, Conflict: "slug"}, nil
	}

	handleTaken, err := c.channelRepo.HandleExists(ctx, identifier, "")
	if err != nil {
		return nil, err
	}
	if handleTaken {
		return &AvailabilityResult{Available: false, Conflict: "handle"}, nil
	}

	return &AvailabilityResult{Available: true}, nil
}

func (c *Checker) IsHandleAvailable(ctx context.Context, handle string, excludeChannelID string) (*AvailabilityResult, error) {
	handle = strings.ToLower(handle)

	handleTaken, err := c.channelRepo.HandleExists(ctx, handle, excludeChannelID)
	if err != nil {
		return nil, err
	}
	if handleTaken {
		return &AvailabilityResult{Available: false, Conflict: "handle"}, nil
	}

	usernameTaken, err := c.userRepo.UsernameExists(ctx, handle)
	if err != nil {
		return nil, err
	}
	if usernameTaken {
		return &AvailabilityResult{Available: false, Conflict: "username"}, nil
	}

	slugTaken, err := c.userRepo.SlugExists(ctx, handle, "")
	if err != nil {
		return nil, err
	}
	if slugTaken {
		return &AvailabilityResult{Available: false, Conflict: "slug"}, nil
	}

	return &AvailabilityResult{Available: true}, nil
}

func (c *Checker) IsUsernameAvailable(ctx context.Context, username string, excludeUserID string) (*AvailabilityResult, error) {
	return c.IsIdentifierAvailable(ctx, username, excludeUserID)
}

func (c *Checker) IsSlugAvailable(ctx context.Context, slug string, excludeUserID string) (*AvailabilityResult, error) {
	return c.IsIdentifierAvailable(ctx, slug, excludeUserID)
}

var _ = log.Debug
