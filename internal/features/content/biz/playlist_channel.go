/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// PlaylistMediaItem represents a media item within a playlist (simplified for display).
type PlaylistMediaItem struct {
	ID             string    `json:"id"`
	ShortToken     string    `json:"short_token"`
	Title          string    `json:"title"`
	Thumbnail      string    `json:"thumbnail"`
	Duration       int       `json:"duration"`
	Type           string    `json:"type"`
	ViewCount      int64     `json:"view_count"`
	EncodingStatus string    `json:"encoding_status"`
	CreateTime     time.Time `json:"create_time"`
}

// Playlist represents a user's media playlist.
type Playlist struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	ShortToken   string              `json:"short_token"`
	UserID       string              `json:"user_id"`
	IsPublic     bool                `json:"is_public"`
	CreateTime   time.Time           `json:"create_time"`
	UpdateTime   time.Time           `json:"update_time"`
	MediaItems   []string            `json:"media_items,omitempty"`
	MediaDetails []PlaylistMediaItem `json:"media_details,omitempty"`
}

// ChannelLink represents an external link associated with a channel.
type ChannelLink struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Title    string `json:"title"`
}

// Channel represents a content channel.
type Channel struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Title           string         `json:"title"`
	Slug            string         `json:"slug"`
	Handle          string         `json:"handle"`
	Description     string         `json:"description"`
	Avatar          string         `json:"avatar"`
	Banner          string         `json:"banner"`
	BannerLogo      string         `json:"banner_logo"` // DEPRECATED
	ShortToken      string         `json:"short_token"`
	Status          string         `json:"status"`
	Privacy         string         `json:"privacy"`
	IsVerified      bool           `json:"is_verified"`
	Tags            []string       `json:"tags"`
	CategoryID      *int64         `json:"category_id,omitempty"`
	SubscriberCount int64          `json:"subscriber_count"`
	MediaCount      int            `json:"media_count"`
	ArticleCount    int            `json:"article_count"`
	TotalViews      int64          `json:"total_views"`
	Links           []ChannelLink  `json:"links"`
	UserID          string         `json:"user_id"`
	CreateTime      time.Time      `json:"create_time"`
	UpdateTime      time.Time      `json:"update_time"`

	// View context (not stored in DB)
	IsOwner      bool `json:"is_owner"`
	IsSubscribed bool `json:"is_subscribed"`
}

// HandleResolutionResult represents the result of a handle resolution.
type HandleResolutionResult struct {
	Type    string    // "channel", "user", "not_found"
	Channel *Channel  // Set if Type == "channel"
	User    *User     // Set if Type == "user"
}

// User represents a minimal user for handle resolution.
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Logo        string    `json:"logo"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"create_time"`
}

// PlaylistRepo defines storage operations for playlists.
type PlaylistRepo interface {
	Create(ctx context.Context, p *Playlist) (*Playlist, error)
	Get(ctx context.Context, id string) (*Playlist, error)
	GetByShortToken(ctx context.Context, token string) (*Playlist, error)
	Update(ctx context.Context, p *Playlist) (*Playlist, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Playlist, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*Playlist, int, error)
	AddMedia(ctx context.Context, playlistID, mediaID string) error
	RemoveMedia(ctx context.Context, playlistID, mediaID string) error
	ReorderMedia(ctx context.Context, playlistID string, mediaOrders map[string]int) error
	GetPlaylistMedia(ctx context.Context, playlistID string) ([]string, error)
	GetPlaylistMediaDetails(ctx context.Context, playlistID string) ([]PlaylistMediaItem, error)
}

// ChannelRepo defines storage operations for channels.
type ChannelRepo interface {
	Create(ctx context.Context, ch *Channel) (*Channel, error)
	Get(ctx context.Context, id string) (*Channel, error)
	GetByUsername(ctx context.Context, username string) (*Channel, error)
	GetByShortToken(ctx context.Context, token string) (*Channel, error)
	GetByHandle(ctx context.Context, handle string) (*Channel, error)
	GetBySlug(ctx context.Context, slug string) (*Channel, error)
	CountByUser(ctx context.Context, userID string) (int, error)
	Update(ctx context.Context, ch *Channel) (*Channel, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Channel, int, error)
	ListPublic(ctx context.Context, page, pageSize int) ([]*Channel, int, error)
	AddMedia(ctx context.Context, channelID, mediaID string) error
	RemoveMedia(ctx context.Context, channelID, mediaID string) error
	// Subscription methods
	Subscribe(ctx context.Context, channelID, userID string) error
	Unsubscribe(ctx context.Context, channelID, userID string) error
	IsSubscribed(ctx context.Context, channelID, userID string) (bool, error)
	GetSubscribers(ctx context.Context, channelID string, page, pageSize int) ([]string, int, error)
	GetSubscriberCount(ctx context.Context, channelID string) (int, error)
	// Invitation methods
	InviteUserToChannel(ctx context.Context, channelID, userID string) error
	AcceptChannelInvitation(ctx context.Context, channelID, userID string) error
	RejectChannelInvitation(ctx context.Context, channelID, userID string) error
	GetChannelInvitations(ctx context.Context, userID string) ([]string, error)
	IsInvitedToChannel(ctx context.Context, channelID, userID string) (bool, error)
	// Cross-entity query methods (for handler migration)
	GetSubscribedChannelIDs(ctx context.Context, userID string) ([]string, error)
	GetSubscriptionVideos(ctx context.Context, userID string, channelIDs []string, sortBy string, page, limit int) ([]*SubscriptionVideoItem, int, error)
	GetChannelVideos(ctx context.Context, token string, sortBy string, page, limit int) ([]*SubscriptionVideoItem, int, error)
	GetChannelPlaylists(ctx context.Context, token string, page, limit int) ([]*ChannelPlaylistItem, int, error)
}

// SystemConfigRepo defines storage operations for system configuration.
type SystemConfigRepo interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	ListByCategory(ctx context.Context, category string) (map[string]string, error)
	Delete(ctx context.Context, key string) error
}

// UserRepo defines user lookup operations needed for handle resolution.
type UserRepo interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
}

// PlaylistChannelUseCase handles playlist and channel business logic.
type PlaylistChannelUseCase struct {
	playlistRepo PlaylistRepo
	channelRepo  ChannelRepo
	configRepo   SystemConfigRepo
	userRepo     UserRepo
	log          *log.Helper
}

func NewPlaylistChannelUseCase(pRepo PlaylistRepo, chRepo ChannelRepo, configRepo SystemConfigRepo, userRepo UserRepo, logger log.Logger) *PlaylistChannelUseCase {
	return &PlaylistChannelUseCase{
		playlistRepo: pRepo,
		channelRepo:  chRepo,
		configRepo:   configRepo,
		userRepo:     userRepo,
		log:          log.NewHelper(log.With(logger, "module", "playlist_channel.biz")),
	}
}

// Channel creation constants
const (
	MinHandleLength    = 3
	MaxHandleLength    = 39
	MinNameLength      = 3
	MaxNameLength      = 150
	MaxTagsPerChannel  = 10
	DefaultMaxChannels = 5
)

// Playlist methods

func (uc *PlaylistChannelUseCase) CreatePlaylist(ctx context.Context, p *Playlist) (*Playlist, error) {
	return uc.playlistRepo.Create(ctx, p)
}

func (uc *PlaylistChannelUseCase) ListPlaylists(ctx context.Context, page, pageSize int) ([]*Playlist, int, error) {
	return uc.playlistRepo.ListAll(ctx, page, pageSize)
}

func (uc *PlaylistChannelUseCase) GetPlaylist(ctx context.Context, id string) (*Playlist, error) {
	return uc.playlistRepo.Get(ctx, id)
}

func (uc *PlaylistChannelUseCase) GetPlaylistByShortToken(ctx context.Context, token string) (*Playlist, error) {
	return uc.playlistRepo.GetByShortToken(ctx, token)
}

func (uc *PlaylistChannelUseCase) ListUserPlaylists(ctx context.Context, userID string, page, pageSize int) ([]*Playlist, int, error) {
	return uc.playlistRepo.ListByUser(ctx, userID, page, pageSize)
}

func (uc *PlaylistChannelUseCase) UpdatePlaylist(ctx context.Context, p *Playlist, userID string, isAdmin bool) (*Playlist, error) {
	existing, err := uc.playlistRepo.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID && !isAdmin {
		return nil, fmt.Errorf("permission denied")
	}
	return uc.playlistRepo.Update(ctx, p)
}

func (uc *PlaylistChannelUseCase) DeletePlaylist(ctx context.Context, id string, userID string, isAdmin bool) error {
	existing, err := uc.playlistRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.playlistRepo.Delete(ctx, id)
}

func (uc *PlaylistChannelUseCase) AddMediaToPlaylist(ctx context.Context, playlistID, mediaID string, userID string, isAdmin bool) error {
	existing, err := uc.playlistRepo.Get(ctx, playlistID)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.playlistRepo.AddMedia(ctx, playlistID, mediaID)
}

func (uc *PlaylistChannelUseCase) RemoveMediaFromPlaylist(ctx context.Context, playlistID, mediaID string, userID string, isAdmin bool) error {
	existing, err := uc.playlistRepo.Get(ctx, playlistID)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.playlistRepo.RemoveMedia(ctx, playlistID, mediaID)
}

func (uc *PlaylistChannelUseCase) ReorderMediaInPlaylist(ctx context.Context, playlistID string, mediaOrders map[string]int, userID string, isAdmin bool) error {
	existing, err := uc.playlistRepo.Get(ctx, playlistID)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.playlistRepo.ReorderMedia(ctx, playlistID, mediaOrders)
}

// Channel methods

func (uc *PlaylistChannelUseCase) CreateChannel(ctx context.Context, ch *Channel) (*Channel, error) {
	// Check channel count limit (from system config)
	maxChannels, err := uc.GetMaxChannelsForUser(ctx, ch.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel limit: %w", err)
	}
	if maxChannels != -1 { // -1 means unlimited (admin)
		count, err := uc.channelRepo.CountByUser(ctx, ch.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to check channel count: %w", err)
		}
		if count >= maxChannels {
			return nil, fmt.Errorf("channel_limit_reached: maximum %d channels allowed (current: %d)", maxChannels, count)
		}
	}

	// Validate handle uniqueness
	existing, _ := uc.channelRepo.GetByHandle(ctx, ch.Handle)
	if existing != nil {
		return nil, fmt.Errorf("handle_already_taken: @%s", ch.Handle)
	}

	// Validate tags count
	if len(ch.Tags) > MaxTagsPerChannel {
		return nil, fmt.Errorf("too_many_tags: maximum %d tags allowed", MaxTagsPerChannel)
	}

	return uc.channelRepo.Create(ctx, ch)
}

func (uc *PlaylistChannelUseCase) GetChannel(ctx context.Context, id string) (*Channel, error) {
	return uc.channelRepo.Get(ctx, id)
}

// GetChannelByUsername looks up a channel by username (fallback for @username URLs when slug is not set).
func (uc *PlaylistChannelUseCase) GetChannelByUsername(ctx context.Context, username string) (*Channel, error) {
	return uc.channelRepo.GetByUsername(ctx, username)
}

func (uc *PlaylistChannelUseCase) GetByShortToken(ctx context.Context, token string) (*Channel, error) {
	return uc.channelRepo.GetByShortToken(ctx, token)
}

func (uc *PlaylistChannelUseCase) GetByHandle(ctx context.Context, handle string) (*Channel, error) {
	return uc.channelRepo.GetByHandle(ctx, handle)
}

// GetMaxChannelsForUser returns the maximum number of channels a user can create.
// Returns -1 for unlimited (admin users).
func (uc *PlaylistChannelUseCase) GetMaxChannelsForUser(ctx context.Context, userID string) (int, error) {
	// Admin users have unlimited channels
	// TODO: Check user role from context or user service
	// For now, read from config

	// Check per-role override
	// roleKey := fmt.Sprintf("max_channels_per_role:%s", role)
	// if val, err := uc.configRepo.Get(ctx, roleKey); err == nil {
	//     if limit, err := strconv.Atoi(val); err == nil {
	//         return limit, nil
	//     }
	// }

	// Fall back to global default
	val, err := uc.configRepo.Get(ctx, "max_channels_per_user")
	if err != nil {
		return DefaultMaxChannels, nil // Hard-coded fallback
	}
	limit := 0
	if _, err := fmt.Sscanf(val, "%d", &limit); err != nil || limit <= 0 {
		return DefaultMaxChannels, nil
	}
	return limit, nil
}

// GetChannelLimits returns the channel creation limits for a user.
func (uc *PlaylistChannelUseCase) GetChannelLimits(ctx context.Context, userID string, isAdmin bool) (maxChannels int, currentCount int, canCreate bool, err error) {
	if isAdmin {
		return -1, 0, true, nil // Unlimited
	}

	maxChannels, err = uc.GetMaxChannelsForUser(ctx, userID)
	if err != nil {
		return 0, 0, false, err
	}

	currentCount, err = uc.channelRepo.CountByUser(ctx, userID)
	if err != nil {
		return 0, 0, false, err
	}

	if maxChannels == -1 {
		canCreate = true // Unlimited
	} else {
		canCreate = currentCount < maxChannels
	}

	return maxChannels, currentCount, canCreate, nil
}

// ResolveHandle resolves a handle to a channel, user, or not_found.
func (uc *PlaylistChannelUseCase) ResolveHandle(ctx context.Context, handle string) (*HandleResolutionResult, error) {
	// Step 1: Try channel handle lookup (indexed)
	channel, err := uc.channelRepo.GetByHandle(ctx, handle)
	if err == nil && channel != nil {
		return &HandleResolutionResult{Type: "channel", Channel: channel}, nil
	}

	// Step 2: Try username lookup (indexed)
	user, err := uc.userRepo.GetByUsername(ctx, handle)
	if err == nil && user != nil {
		return &HandleResolutionResult{Type: "user", User: user}, nil
	}

	// Step 3: Not found
	return &HandleResolutionResult{Type: "not_found"}, nil
}

// ValidateHandle checks if a handle is available for use.
func (uc *PlaylistChannelUseCase) ValidateHandle(ctx context.Context, handle string) (bool, error) {
	existing, err := uc.channelRepo.GetByHandle(ctx, handle)
	if err != nil {
		return true, nil // No existing channel found = available
	}
	return existing == nil, nil
}

func (uc *PlaylistChannelUseCase) ListChannels(ctx context.Context, page, pageSize int) ([]*Channel, int, error) {
	return uc.channelRepo.ListPublic(ctx, page, pageSize)
}

func (uc *PlaylistChannelUseCase) ListUserChannels(ctx context.Context, userID string, page, pageSize int) ([]*Channel, int, error) {
	return uc.channelRepo.ListByUser(ctx, userID, page, pageSize)
}

func (uc *PlaylistChannelUseCase) UpdateChannel(ctx context.Context, ch *Channel, userID string, isAdmin bool) (*Channel, error) {
	existing, err := uc.channelRepo.Get(ctx, ch.ID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID && !isAdmin {
		return nil, fmt.Errorf("permission denied")
	}
	return uc.channelRepo.Update(ctx, ch)
}

func (uc *PlaylistChannelUseCase) DeleteChannel(ctx context.Context, id string, userID string, isAdmin bool) error {
	existing, err := uc.channelRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.channelRepo.Delete(ctx, id)
}

func (uc *PlaylistChannelUseCase) AddMediaToChannel(ctx context.Context, channelID, mediaID string, userID string, isAdmin bool) error {
	existing, err := uc.channelRepo.Get(ctx, channelID)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.channelRepo.AddMedia(ctx, channelID, mediaID)
}

func (uc *PlaylistChannelUseCase) RemoveMediaFromChannel(ctx context.Context, channelID, mediaID string, userID string, isAdmin bool) error {
	existing, err := uc.channelRepo.Get(ctx, channelID)
	if err != nil {
		return err
	}
	if existing.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.channelRepo.RemoveMedia(ctx, channelID, mediaID)
}

// Subscription methods

func (uc *PlaylistChannelUseCase) SubscribeToChannel(ctx context.Context, channelToken, userID string) error {
	// Resolve short_token to channel first
	ch, err := uc.channelRepo.GetByShortToken(ctx, channelToken)
	if err != nil {
		return fmt.Errorf("channel_not_found")
	}
	// Prevent self-subscription
	if ch.UserID == userID {
		return fmt.Errorf("cannot_subscribe_own_channel")
	}
	return uc.channelRepo.Subscribe(ctx, ch.ID, userID)
}

func (uc *PlaylistChannelUseCase) UnsubscribeFromChannel(ctx context.Context, channelToken, userID string) error {
	// Resolve short_token to channel first
	ch, err := uc.channelRepo.GetByShortToken(ctx, channelToken)
	if err != nil {
		return fmt.Errorf("channel_not_found")
	}
	return uc.channelRepo.Unsubscribe(ctx, ch.ID, userID)
}

func (uc *PlaylistChannelUseCase) IsSubscribedToChannel(ctx context.Context, channelToken, userID string) (bool, error) {
	// Resolve short_token to channel first
	ch, err := uc.channelRepo.GetByShortToken(ctx, channelToken)
	if err != nil {
		return false, fmt.Errorf("channel_not_found")
	}
	return uc.channelRepo.IsSubscribed(ctx, ch.ID, userID)
}

func (uc *PlaylistChannelUseCase) GetChannelSubscribers(ctx context.Context, channelToken string, page, pageSize int) ([]string, int, error) {
	// Resolve short_token to channel first
	ch, err := uc.channelRepo.GetByShortToken(ctx, channelToken)
	if err != nil {
		return nil, 0, fmt.Errorf("channel_not_found")
	}
	return uc.channelRepo.GetSubscribers(ctx, ch.ID, page, pageSize)
}

func (uc *PlaylistChannelUseCase) GetChannelSubscriberCount(ctx context.Context, channelToken string) (int, error) {
	// Resolve short_token to channel first
	ch, err := uc.channelRepo.GetByShortToken(ctx, channelToken)
	if err != nil {
		return 0, fmt.Errorf("channel_not_found")
	}
	return uc.channelRepo.GetSubscriberCount(ctx, ch.ID)
}

// Invitation methods

func (uc *PlaylistChannelUseCase) InviteUserToChannel(ctx context.Context, channelToken, userID, inviterID string, isAdmin bool) error {
	// Resolve short_token to channel first
	existing, err := uc.channelRepo.GetByShortToken(ctx, channelToken)
	if err != nil {
		return fmt.Errorf("channel_not_found")
	}
	// Check if inviter is channel owner or admin
	if existing.UserID != inviterID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.channelRepo.InviteUserToChannel(ctx, existing.ID, userID)
}

func (uc *PlaylistChannelUseCase) AcceptChannelInvitation(ctx context.Context, channelID, userID string) error {
	// Check if user is invited
	isInvited, err := uc.channelRepo.IsInvitedToChannel(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !isInvited {
		return fmt.Errorf("no invitation found")
	}
	// Accept invitation by subscribing
	err = uc.channelRepo.AcceptChannelInvitation(ctx, channelID, userID)
	if err != nil {
		return err
	}
	return uc.channelRepo.Subscribe(ctx, channelID, userID)
}

func (uc *PlaylistChannelUseCase) RejectChannelInvitation(ctx context.Context, channelID, userID string) error {
	// Check if user is invited
	isInvited, err := uc.channelRepo.IsInvitedToChannel(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !isInvited {
		return fmt.Errorf("no invitation found")
	}
	return uc.channelRepo.RejectChannelInvitation(ctx, channelID, userID)
}

func (uc *PlaylistChannelUseCase) GetChannelInvitations(ctx context.Context, userID string) ([]string, error) {
	return uc.channelRepo.GetChannelInvitations(ctx, userID)
}

func (uc *PlaylistChannelUseCase) IsInvitedToChannel(ctx context.Context, channelID, userID string) (bool, error) {
	return uc.channelRepo.IsInvitedToChannel(ctx, channelID, userID)
}

// --- Cross-entity query methods for handler migration ---

// SubscriptionVideoItem represents a video from a subscribed channel.
type SubscriptionVideoItem struct {
	ID             string    `json:"id"`
	ShortToken     string    `json:"short_token"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Thumbnail      string    `json:"thumbnail"`
	Duration       int       `json:"duration"`
	ViewCount      int64     `json:"view_count"`
	LikeCount      int64     `json:"like_count"`
	CommentCount   int64     `json:"comment_count"`
	Type           string    `json:"type"`
	ChannelID      string    `json:"channel_id"`
	UserID         string    `json:"user_id"`
	EncodingStatus string    `json:"encoding_status"`
	CreateTime     time.Time `json:"create_time"`
	PublishedAt    time.Time `json:"published_at"`
}

// ChannelPlaylistItem represents a playlist in a channel.
type ChannelPlaylistItem struct {
	ID          string    `json:"id"`
	ShortToken  string    `json:"short_token"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	Privacy     string    `json:"privacy"`
	CreateTime  time.Time `json:"create_time"`
}

// GetSubscriptionVideos returns paginated videos from channels the user is subscribed to.
func (uc *PlaylistChannelUseCase) GetSubscriptionVideos(ctx context.Context, userID string, channelIDs []string, sortBy string, page, limit int) ([]*SubscriptionVideoItem, int, error) {
	return uc.channelRepo.GetSubscriptionVideos(ctx, userID, channelIDs, sortBy, page, limit)
}

// GetChannelVideos returns paginated videos for a channel by short_token.
func (uc *PlaylistChannelUseCase) GetChannelVideos(ctx context.Context, token string, sortBy string, page, limit int) ([]*SubscriptionVideoItem, int, error) {
	return uc.channelRepo.GetChannelVideos(ctx, token, sortBy, page, limit)
}

// GetChannelPlaylists returns paginated playlists for a channel by short_token.
func (uc *PlaylistChannelUseCase) GetChannelPlaylists(ctx context.Context, token string, page, limit int) ([]*ChannelPlaylistItem, int, error) {
	return uc.channelRepo.GetChannelPlaylists(ctx, token, page, limit)
}

// GetSubscribedChannelIDs returns all channel IDs the user is subscribed to.
func (uc *PlaylistChannelUseCase) GetSubscribedChannelIDs(ctx context.Context, userID string) ([]string, error) {
	return uc.channelRepo.GetSubscribedChannelIDs(ctx, userID)
}
