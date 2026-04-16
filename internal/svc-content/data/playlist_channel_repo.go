/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/channel"
	"origadmin/application/origcms/internal/data/entity/mediaplaylist"
	"origadmin/application/origcms/internal/data/entity/playlist"
	"origadmin/application/origcms/internal/data/entity/subscription"
	"origadmin/application/origcms/internal/svc-content/biz"
)

type playlistRepo struct {
	data *Data
	log  *log.Helper
}

type channelRepo struct {
	data *Data
	log  *log.Helper
}

func NewPlaylistRepo(data *Data, logger log.Logger) biz.PlaylistRepo {
	return &playlistRepo{data: data, log: log.NewHelper(log.With(logger, "module", "playlist.data"))}
}

func NewChannelRepo(data *Data, logger log.Logger) biz.ChannelRepo {
	return &channelRepo{data: data, log: log.NewHelper(log.With(logger, "module", "channel.data"))}
}

func (r *playlistRepo) Create(ctx context.Context, p *biz.Playlist) (*biz.Playlist, error) {
	privacy := 1 // Default to public
	if !p.IsPublic {
		privacy = 0
	}
	ent, err := r.data.db.Playlist.Create().
		SetTitle(p.Name).
		SetDescription(p.Description).
		SetUserID(p.UserID).
		SetPrivacy(privacy).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapPlaylist(ent), nil
}

func (r *playlistRepo) Get(ctx context.Context, id string) (*biz.Playlist, error) {
	ent, err := r.data.db.Playlist.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	playlist := mapPlaylist(ent)
	// Get media items for the playlist
	mediaItems, err := r.GetPlaylistMedia(ctx, id)
	if err == nil {
		playlist.MediaItems = mediaItems
	}
	return playlist, nil
}

func (r *playlistRepo) Update(ctx context.Context, p *biz.Playlist) (*biz.Playlist, error) {
	privacy := 1 // Default to public
	if !p.IsPublic {
		privacy = 0
	}
	ent, err := r.data.db.Playlist.UpdateOneID(p.ID).
		SetTitle(p.Name).
		SetDescription(p.Description).
		SetPrivacy(privacy).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapPlaylist(ent), nil
}

func (r *playlistRepo) Delete(ctx context.Context, id string) error {
	// First delete associated media playlist entries
	_, err := r.data.db.MediaPlaylist.Delete().
		Where(mediaplaylist.PlaylistIDEQ(id)).
		Exec(ctx)
	if err != nil {
		return err
	}
	// Then delete the playlist itself
	return r.data.db.Playlist.DeleteOneID(id).Exec(ctx)
}

func (r *playlistRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*biz.Playlist, int, error) {
	query := r.data.db.Playlist.Query().Where(playlist.UserIDEQ(userID))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*biz.Playlist, len(ents))
	for i, ent := range ents {
		playlist := mapPlaylist(ent)
		// Get media items for each playlist
		mediaItems, err := r.GetPlaylistMedia(ctx, ent.ID)
		if err == nil {
			playlist.MediaItems = mediaItems
		}
		res[i] = playlist
	}
	return res, total, nil
}

func (r *playlistRepo) ListAll(ctx context.Context, page, pageSize int) ([]*biz.Playlist, int, error) {
	query := r.data.db.Playlist.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*biz.Playlist, len(ents))
	for i, ent := range ents {
		playlist := mapPlaylist(ent)
		// Get media items for each playlist
		mediaItems, err := r.GetPlaylistMedia(ctx, ent.ID)
		if err == nil {
			playlist.MediaItems = mediaItems
		}
		res[i] = playlist
	}
	return res, total, nil
}

func (r *playlistRepo) AddMedia(ctx context.Context, playlistID, mediaID string) error {
	// Check if already in playlist
	exists, _ := r.data.db.MediaPlaylist.Query().
		Where(
			mediaplaylist.PlaylistIDEQ(playlistID),
			mediaplaylist.MediaIDEQ(mediaID),
		).Exist(ctx)
	if exists {
		return nil
	}

	// Get current max ordering
	maxOrder := 0
	last, _ := r.data.db.MediaPlaylist.Query().
		Where(mediaplaylist.PlaylistIDEQ(playlistID)).
		Order(entity.Desc(mediaplaylist.FieldOrdering)).
		First(ctx)
	if last != nil {
		maxOrder = last.Ordering
	}

	return r.data.db.MediaPlaylist.Create().
		SetPlaylistID(playlistID).
		SetMediaID(mediaID).
		SetOrdering(maxOrder + 1).
		Exec(ctx)
}

func (r *playlistRepo) RemoveMedia(ctx context.Context, playlistID, mediaID string) error {
	_, err := r.data.db.MediaPlaylist.Delete().
		Where(
			mediaplaylist.PlaylistIDEQ(playlistID),
			mediaplaylist.MediaIDEQ(mediaID),
		).Exec(ctx)
	return err
}

func (r *playlistRepo) ReorderMedia(ctx context.Context, playlistID string, mediaOrders map[string]int) error {
	for mediaID, newOrder := range mediaOrders {
		err := r.data.db.MediaPlaylist.Update().
			Where(
				mediaplaylist.PlaylistIDEQ(playlistID),
				mediaplaylist.MediaIDEQ(mediaID),
			).
			SetOrdering(newOrder).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *playlistRepo) GetPlaylistMedia(ctx context.Context, playlistID string) ([]string, error) {
	medias, err := r.data.db.MediaPlaylist.Query().
		Where(mediaplaylist.PlaylistIDEQ(playlistID)).
		Order(entity.Asc(mediaplaylist.FieldOrdering)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	mediaIDs := make([]string, len(medias))
	for i, media := range medias {
		mediaIDs[i] = media.MediaID
	}
	return mediaIDs, nil
}

func (r *channelRepo) Create(ctx context.Context, ch *biz.Channel) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.Create().
		SetTitle(ch.Title).
		SetDescription(ch.Description).
		SetBannerLogo(ch.BannerLogo).
		SetUserID(ch.UserID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) Get(ctx context.Context, id string) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) GetBySlug(ctx context.Context, slug string) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.Query().Where(channel.SlugEQ(slug)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) Update(ctx context.Context, ch *biz.Channel) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.UpdateOneID(ch.ID).
		SetTitle(ch.Title).
		SetDescription(ch.Description).
		SetBannerLogo(ch.BannerLogo).
		SetIsPublic(ch.IsPublic).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.Channel.DeleteOneID(id).Exec(ctx)
}

func (r *channelRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*biz.Channel, int, error) {
	query := r.data.db.Channel.Query().Where(channel.UserIDEQ(userID))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*biz.Channel, len(ents))
	for i, ent := range ents {
		res[i] = mapChannel(ent)
	}
	return res, total, nil
}

func (r *channelRepo) ListAll(ctx context.Context, page, pageSize int) ([]*biz.Channel, int, error) {
	query := r.data.db.Channel.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*biz.Channel, len(ents))
	for i, ent := range ents {
		res[i] = mapChannel(ent)
	}
	return res, total, nil
}

func (r *channelRepo) AddMedia(ctx context.Context, channelID, mediaID string) error {
	return r.data.db.Media.UpdateOneID(mediaID).
		SetChannelID(channelID).
		Exec(ctx)
}

func (r *channelRepo) RemoveMedia(ctx context.Context, channelID, mediaID string) error {
	return r.data.db.Media.UpdateOneID(mediaID).
		ClearChannel().
		Exec(ctx)
}

// Subscription methods

func (r *channelRepo) Subscribe(ctx context.Context, channelID, userID string) error {
	// Check if subscription already exists
	exists, _ := r.data.db.Subscription.Query().
		Where(
			subscription.ChannelIDEQ(channelID),
			subscription.SubscriberIDEQ(userID),
		).Exist(ctx)
	if exists {
		return nil // Already subscribed
	}

	// Create new subscription
	_, err := r.data.db.Subscription.Create().
		SetChannelID(channelID).
		SetSubscriberID(userID).
		Save(ctx)
	return err
}

func (r *channelRepo) Unsubscribe(ctx context.Context, channelID, userID string) error {
	_, err := r.data.db.Subscription.Delete().
		Where(
			subscription.ChannelIDEQ(channelID),
			subscription.SubscriberIDEQ(userID),
		).Exec(ctx)
	return err
}

func (r *channelRepo) IsSubscribed(ctx context.Context, channelID, userID string) (bool, error) {
	return r.data.db.Subscription.Query().
		Where(
			subscription.ChannelIDEQ(channelID),
			subscription.SubscriberIDEQ(userID),
		).Exist(ctx)
}

func (r *channelRepo) GetSubscribers(ctx context.Context, channelID string, page, pageSize int) ([]string, int, error) {
	query := r.data.db.Subscription.Query().Where(subscription.ChannelIDEQ(channelID))
	
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	subscribers := make([]string, len(ents))
	for i, ent := range ents {
		subscribers[i] = ent.SubscriberID
	}
	
	return subscribers, total, nil
}

func (r *channelRepo) GetSubscriberCount(ctx context.Context, channelID string) (int, error) {
	return r.data.db.Subscription.Query().
		Where(subscription.ChannelIDEQ(channelID)).
		Count(ctx)
}

// Invitation methods - using in-memory storage as a workaround
// Note: This is a temporary solution. In a production environment, you should use a proper storage system.

var channelInvitations = make(map[string]map[string]bool) // channelID -> userID -> isInvited

func (r *channelRepo) InviteUserToChannel(ctx context.Context, channelID, userID string) error {
	if _, ok := channelInvitations[channelID]; !ok {
		channelInvitations[channelID] = make(map[string]bool)
	}
	channelInvitations[channelID][userID] = true
	return nil
}

func (r *channelRepo) AcceptChannelInvitation(ctx context.Context, channelID, userID string) error {
	if _, ok := channelInvitations[channelID]; ok {
		delete(channelInvitations[channelID], userID)
	}
	return nil
}

func (r *channelRepo) RejectChannelInvitation(ctx context.Context, channelID, userID string) error {
	if _, ok := channelInvitations[channelID]; ok {
		delete(channelInvitations[channelID], userID)
	}
	return nil
}

func (r *channelRepo) GetChannelInvitations(ctx context.Context, userID string) ([]string, error) {
	var invitations []string
	for channelID, users := range channelInvitations {
		if users[userID] {
			invitations = append(invitations, channelID)
		}
	}
	return invitations, nil
}

func (r *channelRepo) IsInvitedToChannel(ctx context.Context, channelID, userID string) (bool, error) {
	if users, ok := channelInvitations[channelID]; ok {
		return users[userID], nil
	}
	return false, nil
}

func mapPlaylist(ent *entity.Playlist) *biz.Playlist {
	return &biz.Playlist{
		ID:          ent.ID,
		Name:        ent.Title,
		Description: ent.Description,
		UserID:      ent.UserID,
		IsPublic:    ent.Privacy == 1,
		CreatedAt:   ent.AddDate,
		UpdatedAt:   ent.AddDate,
		MediaItems:  []string{},
	}
}

func mapChannel(ent *entity.Channel) *biz.Channel {
	return &biz.Channel{
		ID:            ent.ID,
		Title:         ent.Title,
		Description:   ent.Description,
		BannerLogo:    ent.BannerLogo,
		FriendlyToken: ent.Slug,
		IsPublic:      ent.IsPublic,
		UserID:        ent.UserID,
		CreatedAt:     ent.AddDate,
	}
}
