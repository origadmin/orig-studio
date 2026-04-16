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

// Playlist represents a user's media playlist.
type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MediaItems  []string  `json:"media_items,omitempty"`
}

// Channel represents a content channel.
type Channel struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	BannerLogo    string    `json:"banner_logo"`
	FriendlyToken string    `json:"friendly_token"`
	IsPublic      bool      `json:"is_public"`
	UserID        string    `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// PlaylistRepo defines storage operations for playlists.
type PlaylistRepo interface {
	Create(ctx context.Context, p *Playlist) (*Playlist, error)
	Get(ctx context.Context, id string) (*Playlist, error)
	Update(ctx context.Context, p *Playlist) (*Playlist, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Playlist, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*Playlist, int, error)
	AddMedia(ctx context.Context, playlistID, mediaID string) error
	RemoveMedia(ctx context.Context, playlistID, mediaID string) error
	ReorderMedia(ctx context.Context, playlistID string, mediaOrders map[string]int) error
	GetPlaylistMedia(ctx context.Context, playlistID string) ([]string, error)
}

// ChannelRepo defines storage operations for channels.
type ChannelRepo interface {
	Create(ctx context.Context, ch *Channel) (*Channel, error)
	Get(ctx context.Context, id string) (*Channel, error)
	GetBySlug(ctx context.Context, slug string) (*Channel, error)
	Update(ctx context.Context, ch *Channel) (*Channel, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Channel, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*Channel, int, error)
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
}

// PlaylistChannelUseCase handles playlist and channel business logic.
type PlaylistChannelUseCase struct {
	playlistRepo PlaylistRepo
	channelRepo  ChannelRepo
	log          *log.Helper
}

func NewPlaylistChannelUseCase(pRepo PlaylistRepo, chRepo ChannelRepo, logger log.Logger) *PlaylistChannelUseCase {
	return &PlaylistChannelUseCase{
		playlistRepo: pRepo,
		channelRepo:  chRepo,
		log:          log.NewHelper(log.With(logger, "module", "playlist_channel.biz")),
	}
}

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
	return uc.channelRepo.Create(ctx, ch)
}

func (uc *PlaylistChannelUseCase) GetChannel(ctx context.Context, id string) (*Channel, error) {
	return uc.channelRepo.Get(ctx, id)
}

func (uc *PlaylistChannelUseCase) GetChannelBySlug(ctx context.Context, slug string) (*Channel, error) {
	return uc.channelRepo.GetBySlug(ctx, slug)
}

func (uc *PlaylistChannelUseCase) ListChannels(ctx context.Context, page, pageSize int) ([]*Channel, int, error) {
	return uc.channelRepo.ListAll(ctx, page, pageSize)
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

func (uc *PlaylistChannelUseCase) SubscribeToChannel(ctx context.Context, channelID, userID string) error {
	// Check if channel exists
	_, err := uc.channelRepo.Get(ctx, channelID)
	if err != nil {
		return err
	}
	return uc.channelRepo.Subscribe(ctx, channelID, userID)
}

func (uc *PlaylistChannelUseCase) UnsubscribeFromChannel(ctx context.Context, channelID, userID string) error {
	return uc.channelRepo.Unsubscribe(ctx, channelID, userID)
}

func (uc *PlaylistChannelUseCase) IsSubscribedToChannel(ctx context.Context, channelID, userID string) (bool, error) {
	return uc.channelRepo.IsSubscribed(ctx, channelID, userID)
}

func (uc *PlaylistChannelUseCase) GetChannelSubscribers(ctx context.Context, channelID string, page, pageSize int) ([]string, int, error) {
	return uc.channelRepo.GetSubscribers(ctx, channelID, page, pageSize)
}

func (uc *PlaylistChannelUseCase) GetChannelSubscriberCount(ctx context.Context, channelID string) (int, error) {
	return uc.channelRepo.GetSubscriberCount(ctx, channelID)
}

// Invitation methods

func (uc *PlaylistChannelUseCase) InviteUserToChannel(ctx context.Context, channelID, userID, inviterID string, isAdmin bool) error {
	// Check if channel exists
	existing, err := uc.channelRepo.Get(ctx, channelID)
	if err != nil {
		return err
	}
	// Check if inviter is channel owner or admin
	if existing.UserID != inviterID && !isAdmin {
		return fmt.Errorf("permission denied")
	}
	return uc.channelRepo.InviteUserToChannel(ctx, channelID, userID)
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
