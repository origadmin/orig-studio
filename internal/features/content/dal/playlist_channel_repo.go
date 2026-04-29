/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/channel"
	ent "origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/mediaplaylist"
	"origadmin/application/origcms/internal/data/entity/playlist"
	"origadmin/application/origcms/internal/data/entity/subscription"
	"origadmin/application/origcms/internal/data/entity/user"
	"origadmin/application/origcms/internal/features/content/biz"
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
		SetTitle(p.Title).
		SetDescription(p.Description).
		SetUserID(p.UserID).
		SetPrivacy(privacy).
		AddUserIDs(p.UserID).
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

func (r *playlistRepo) GetByShortToken(ctx context.Context, token string) (*biz.Playlist, error) {
	ent, err := r.data.db.Playlist.Query().Where(playlist.ShortTokenEQ(token)).Only(ctx)
	if err != nil {
		return nil, err
	}
	playlist := mapPlaylist(ent)
	// Get media items for the playlist
	mediaItems, err := r.GetPlaylistMedia(ctx, ent.ID)
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
		SetTitle(p.Title).
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
	builder := r.data.db.Channel.Create().
		SetTitle(ch.Title).
		SetDescription(ch.Description).
		SetBannerLogo(ch.BannerLogo).
		SetUserID(ch.UserID).
		SetIsPublic(ch.IsPublic)

	ent, err := builder.Save(ctx)
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

func (r *channelRepo) GetByUsername(ctx context.Context, username string) (*biz.Channel, error) {
	userEnt, err := r.data.db.User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	channels, _, err := r.ListByUser(ctx, userEnt.ID, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel for user: %w", err)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no channel found for user %s", username)
	}

	return channels[0], nil
}

func (r *channelRepo) GetByShortToken(ctx context.Context, token string) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.Query().Where(channel.ShortTokenEQ(token)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) GetDefaultChannel(ctx context.Context, userID string) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.Query().
		Where(channel.UserIDEQ(userID)).
		Order(ent.Asc(channel.FieldID)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) Update(ctx context.Context, ch *biz.Channel) (*biz.Channel, error) {
	builder := r.data.db.Channel.UpdateOneID(ch.ID).
		SetTitle(ch.Title).
		SetDescription(ch.Description).
		SetBannerLogo(ch.BannerLogo).
		SetIsPublic(ch.IsPublic)

	ent, err := builder.Save(ctx)
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

func (r *channelRepo) ListPublic(ctx context.Context, page, pageSize int) ([]*biz.Channel, int, error) {
	query := r.data.db.Channel.Query().Where(channel.IsPublicEQ(true))
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

// --- Cross-entity query methods ---

// GetSubscribedChannelIDs returns all channel IDs the user is subscribed to.
func (r *channelRepo) GetSubscribedChannelIDs(ctx context.Context, userID string) ([]string, error) {
	subscriptions, err := r.data.db.Subscription.Query().
		Where(subscription.SubscriberID(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	channelIDs := make([]string, 0, len(subscriptions))
	for _, sub := range subscriptions {
		channelIDs = append(channelIDs, sub.ChannelID)
	}
	return channelIDs, nil
}

// GetSubscriptionVideos returns paginated videos from channels the user is subscribed to.
func (r *channelRepo) GetSubscriptionVideos(ctx context.Context, userID string, channelIDs []string, sortBy string, page, limit int) ([]*biz.SubscriptionVideoItem, int, error) {
	if len(channelIDs) == 0 {
		return []*biz.SubscriptionVideoItem{}, 0, nil
	}

	query := r.data.db.Media.Query().
		Where(
			media.ChannelIDIn(channelIDs...),
			media.StateEQ("active"),
			media.PrivacyEQ(1), // public only
		)

	// Apply sorting
	switch sortBy {
	case "newest":
		query.Order(entity.Desc(media.FieldCreatedAt))
	case "most_viewed":
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreatedAt))
	case "trending":
		// Trending: prioritize recent videos with high engagement
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		query.Where(media.CreatedAtGTE(sevenDaysAgo))
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreatedAt))
	default:
		query.Order(entity.Desc(media.FieldCreatedAt))
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	medias, err := query.
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.SubscriptionVideoItem, 0, len(medias))
	for _, m := range medias {
		item := &biz.SubscriptionVideoItem{
			ID:             m.ID,
			ShortToken:     m.ShortToken,
			Title:          m.Title,
			Description:    m.Description,
			Thumbnail:      m.Thumbnail,
			Duration:       m.Duration,
			ViewCount:      m.ViewCount,
			LikeCount:      m.LikeCount,
			CommentCount:   m.CommentCount,
			Type:           m.Type,
			ChannelID:      m.ChannelID,
			UserID:         m.UserID,
			EncodingStatus: m.EncodingStatus,
		}
		if !m.CreatedAt.IsZero() {
			item.CreatedAt = m.CreatedAt
		}
		if !m.PublishedAt.IsZero() {
			item.PublishedAt = m.PublishedAt
		}
		items = append(items, item)
	}

	return items, total, nil
}

// GetChannelVideos returns paginated videos for a channel by short_token.
func (r *channelRepo) GetChannelVideos(ctx context.Context, token string, sortBy string, page, limit int) ([]*biz.SubscriptionVideoItem, int, error) {
	// Resolve short_token to channel ID
	ch, err := r.data.db.Channel.Query().
		Where(channel.ShortToken(token)).
		Only(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("channel_not_found")
	}

	// Build media query for this channel
	query := r.data.db.Media.Query().
		Where(
			media.ChannelID(ch.ID),
			media.StateEQ("active"),
		)

	// Apply sorting
	switch sortBy {
	case "newest":
		query.Order(entity.Desc(media.FieldCreatedAt))
	case "oldest":
		query.Order(entity.Asc(media.FieldCreatedAt))
	case "popular":
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreatedAt))
	default:
		query.Order(entity.Desc(media.FieldCreatedAt))
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	medias, err := query.
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.SubscriptionVideoItem, 0, len(medias))
	for _, m := range medias {
		item := &biz.SubscriptionVideoItem{
			ID:             m.ID,
			ShortToken:     m.ShortToken,
			Title:          m.Title,
			Description:    m.Description,
			Thumbnail:      m.Thumbnail,
			Duration:       m.Duration,
			ViewCount:      m.ViewCount,
			LikeCount:      m.LikeCount,
			CommentCount:   m.CommentCount,
			Type:           m.Type,
			ChannelID:      m.ChannelID,
			UserID:         m.UserID,
			EncodingStatus: m.EncodingStatus,
		}
		if !m.CreatedAt.IsZero() {
			item.CreatedAt = m.CreatedAt
		}
		if !m.PublishedAt.IsZero() {
			item.PublishedAt = m.PublishedAt
		}
		items = append(items, item)
	}

	return items, total, nil
}

// GetChannelPlaylists returns paginated playlists for a channel by short_token.
func (r *channelRepo) GetChannelPlaylists(ctx context.Context, token string, page, limit int) ([]*biz.ChannelPlaylistItem, int, error) {
	// Resolve short_token to channel, then get user_id for playlist lookup
	ch, err := r.data.db.Channel.Query().
		Where(channel.ShortToken(token)).
		Only(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("channel_not_found")
	}

	// Query playlists by user_id (channel owner)
	query := r.data.db.Playlist.Query().
		Where(
			playlist.UserID(ch.UserID),
			playlist.PrivacyEQ(1), // public playlists only
		).
		Order(entity.Desc(playlist.FieldAddDate))

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	playlists, err := query.
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.ChannelPlaylistItem, 0, len(playlists))
	for _, p := range playlists {
		item := &biz.ChannelPlaylistItem{
			ID:          p.ID,
			ShortToken:  p.ShortToken,
			Title:       p.Title,
			Description: p.Description,
			UserID:      p.UserID,
			Privacy:     p.Privacy,
		}
		if !p.AddDate.IsZero() {
			item.CreatedAt = p.AddDate
		}
		items = append(items, item)
	}

	return items, total, nil
}

func mapPlaylist(ent *entity.Playlist) *biz.Playlist {
	return &biz.Playlist{
		ID:          ent.ID,
		Title:       ent.Title,
		Description: ent.Description,
		ShortToken:  ent.ShortToken,
		UserID:      ent.UserID,
		IsPublic:    ent.Privacy == 1,
		CreatedAt:   ent.AddDate,
		UpdatedAt:   ent.AddDate,
		MediaItems:  []string{},
	}
}

func mapChannel(ent *entity.Channel) *biz.Channel {
	return &biz.Channel{
		ID:          ent.ID,
		Title:       ent.Title,
		Description: ent.Description,
		BannerLogo:  ent.BannerLogo,
		ShortToken:  ent.ShortToken,
		IsPublic:    ent.IsPublic,
		UserID:      ent.UserID,
		CreatedAt:   ent.AddDate,
	}
}
