/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/channel"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/mediaplaylist"
	"origadmin/application/origcms/internal/data/entity/playlist"
	schema "origadmin/application/origcms/internal/data/entity/schema"
	"origadmin/application/origcms/internal/data/entity/setting"
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

// systemConfigRepo implements biz.SystemConfigRepo using the existing Setting entity.
type systemConfigRepo struct {
	data *Data
	log  *log.Helper
}

// NewSystemConfigRepo creates a new SystemConfigRepo.
func NewSystemConfigRepo(data *Data, logger log.Logger) biz.SystemConfigRepo {
	return &systemConfigRepo{data: data, log: log.NewHelper(log.With(logger, "module", "system_config.data"))}
}

// channelUserRepo implements biz.UserRepo for handle resolution.
type channelUserRepo struct {
	data *Data
	log  *log.Helper
}

// NewChannelUserRepo creates a new UserRepo for handle resolution.
func NewChannelUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &channelUserRepo{data: data, log: log.NewHelper(log.With(logger, "module", "channel_user.data"))}
}

func (r *playlistRepo) Create(ctx context.Context, p *biz.Playlist) (*biz.Playlist, error) {
	privacy := playlist.PrivacyPUBLIC
	if !p.IsPublic {
		privacy = playlist.PrivacyPRIVATE
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
	// Get media details for the playlist
	mediaDetails, err := r.GetPlaylistMediaDetails(ctx, id)
	if err == nil {
		playlist.MediaDetails = mediaDetails
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
	// Get media details for the playlist
	mediaDetails, err := r.GetPlaylistMediaDetails(ctx, ent.ID)
	if err == nil {
		playlist.MediaDetails = mediaDetails
	}
	return playlist, nil
}

func (r *playlistRepo) Update(ctx context.Context, p *biz.Playlist) (*biz.Playlist, error) {
	privacy := playlist.PrivacyPUBLIC
	if !p.IsPublic {
		privacy = playlist.PrivacyPRIVATE
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
		Order(entity.Desc(playlist.FieldCreateTime)).
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
		Order(entity.Desc(playlist.FieldCreateTime)).
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

func (r *playlistRepo) GetPlaylistMediaDetails(ctx context.Context, playlistID string) ([]biz.PlaylistMediaItem, error) {
	// First get ordered media IDs from the join table
	mediaPlaylistItems, err := r.data.db.MediaPlaylist.Query().
		Where(mediaplaylist.PlaylistIDEQ(playlistID)).
		Order(entity.Asc(mediaplaylist.FieldOrdering)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(mediaPlaylistItems) == 0 {
		return []biz.PlaylistMediaItem{}, nil
	}

	// Collect media IDs in order
	mediaIDs := make([]string, len(mediaPlaylistItems))
	for i, mp := range mediaPlaylistItems {
		mediaIDs[i] = mp.MediaID
	}

	// Fetch all media entities by IDs
	mediaEnts, err := r.data.db.Media.Query().
		Where(media.IDIn(mediaIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Build a map for quick lookup
	mediaMap := make(map[string]*entity.Media, len(mediaEnts))
	for _, m := range mediaEnts {
		mediaMap[m.ID] = m
	}

	// Build result in the original order from the join table
	result := make([]biz.PlaylistMediaItem, 0, len(mediaIDs))
	for _, id := range mediaIDs {
		m, ok := mediaMap[id]
		if !ok {
			continue // Skip deleted media
		}
		item := biz.PlaylistMediaItem{
			ID:             m.ID,
			ShortToken:     m.ShortToken,
			Title:          m.Title,
			Thumbnail:      m.Thumbnail,
			Duration:       m.Duration,
			Type:           m.Type,
			ViewCount:      m.ViewCount,
			EncodingStatus: m.EncodingStatus,
		}
		if !m.CreateTime.IsZero() {
			item.CreateTime = m.CreateTime
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *channelRepo) Create(ctx context.Context, ch *biz.Channel) (*biz.Channel, error) {
	privacy := channel.PrivacyPUBLIC
	if ch.Privacy == "PRIVATE" {
		privacy = channel.PrivacyPRIVATE
	} else if ch.Privacy == "UNLISTED" {
		privacy = channel.PrivacyUNLISTED
	} else if ch.Privacy == "PAID" {
		privacy = channel.PrivacyPAID
	} else if ch.Privacy == "SUBSCRIBERS_ONLY" {
		privacy = channel.PrivacySUBSCRIBERS_ONLY
	}

	status := channel.StatusACTIVE
	if ch.Status == "INACTIVE" {
		status = channel.StatusINACTIVE
	} else if ch.Status == "SUSPENDED" {
		status = channel.StatusSUSPENDED
	} else if ch.Status == "PENDING_REVIEW" {
		status = channel.StatusPENDING_REVIEW
	}

	builder := r.data.db.Channel.Create().
		SetName(ch.Name).
		SetTitle(ch.Title).
		SetHandle(ch.Handle).
		SetDescription(ch.Description).
		SetUserID(ch.UserID).
		SetPrivacy(privacy).
		SetStatus(status)

	if ch.Slug != "" {
		builder.SetSlug(ch.Slug)
	}
	// Only set ShortToken when explicitly provided (non-empty);
	// otherwise let ent schema's DefaultFunc (idutil.GenShortID) auto-generate one.
	if ch.ShortToken != "" {
		builder.SetShortToken(ch.ShortToken)
	}
	if ch.Avatar != "" {
		builder.SetAvatar(ch.Avatar)
	}
	if ch.Banner != "" {
		builder.SetBanner(ch.Banner)
	}
	if ch.BannerLogo != "" {
		builder.SetBannerLogo(ch.BannerLogo)
	}
	if ch.Tags != nil {
		builder.SetTags(ch.Tags)
	}
	if ch.CategoryID != nil {
		builder.SetCategoryID(*ch.CategoryID)
	}
	if ch.Links != nil {
		builder.SetLinks(convertBizLinksToSchema(ch.Links))
	}

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
	// Per A009: no default channel concept. Return first channel for backward compat.
	ent, err := r.data.db.Channel.Query().
		Where(channel.UserIDEQ(userID)).
		Order(entity.Asc(channel.FieldID)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return mapChannel(ent), nil
}

func (r *channelRepo) GetByHandle(ctx context.Context, handle string) (*biz.Channel, error) {
	ent, err := r.data.db.Channel.Query().Where(channel.HandleEQ(handle)).Only(ctx)
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

func (r *channelRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	return r.data.db.Channel.Query().Where(channel.UserIDEQ(userID)).Count(ctx)
}

func (r *channelRepo) Update(ctx context.Context, ch *biz.Channel) (*biz.Channel, error) {
	privacy := channel.PrivacyPUBLIC
	if ch.Privacy == "PRIVATE" {
		privacy = channel.PrivacyPRIVATE
	} else if ch.Privacy == "UNLISTED" {
		privacy = channel.PrivacyUNLISTED
	} else if ch.Privacy == "PAID" {
		privacy = channel.PrivacyPAID
	} else if ch.Privacy == "SUBSCRIBERS_ONLY" {
		privacy = channel.PrivacySUBSCRIBERS_ONLY
	}

	status := channel.StatusACTIVE
	if ch.Status == "INACTIVE" {
		status = channel.StatusINACTIVE
	} else if ch.Status == "SUSPENDED" {
		status = channel.StatusSUSPENDED
	} else if ch.Status == "PENDING_REVIEW" {
		status = channel.StatusPENDING_REVIEW
	}

	builder := r.data.db.Channel.UpdateOneID(ch.ID).
		SetName(ch.Name).
		SetTitle(ch.Title).
		SetHandle(ch.Handle).
		SetDescription(ch.Description).
		SetPrivacy(privacy).
		SetStatus(status)

	if ch.Slug != "" {
		builder.SetSlug(ch.Slug)
	}
	if ch.Avatar != "" {
		builder.SetAvatar(ch.Avatar)
	}
	if ch.Banner != "" {
		builder.SetBanner(ch.Banner)
	}
	if ch.BannerLogo != "" {
		builder.SetBannerLogo(ch.BannerLogo)
	}
	if ch.Tags != nil {
		builder.SetTags(ch.Tags)
	}
	if ch.CategoryID != nil {
		builder.SetCategoryID(*ch.CategoryID)
	}
	if ch.Links != nil {
		builder.SetLinks(convertBizLinksToSchema(ch.Links))
	}

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
		Order(entity.Desc(channel.FieldCreateTime)).
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
	query := r.data.db.Channel.Query().Where(channel.PrivacyEQ(channel.PrivacyPUBLIC))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	ents, err := query.
		Order(entity.Desc(channel.FieldCreateTime)).
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
			media.PrivacyEQ(media.PrivacyPUBLIC), // public only
		)

	// Apply sorting
	switch sortBy {
	case "newest":
		query.Order(entity.Desc(media.FieldCreateTime))
	case "most_viewed":
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreateTime))
	case "trending":
		// Trending: prioritize recent videos with high engagement
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		query.Where(media.CreateTimeGTE(sevenDaysAgo))
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreateTime))
	default:
		query.Order(entity.Desc(media.FieldCreateTime))
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
		if !m.CreateTime.IsZero() {
			item.CreateTime = m.CreateTime
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
		query.Order(entity.Desc(media.FieldCreateTime))
	case "oldest":
		query.Order(entity.Asc(media.FieldCreateTime))
	case "popular":
		query.Order(entity.Desc(media.FieldViewCount), entity.Desc(media.FieldCreateTime))
	default:
		query.Order(entity.Desc(media.FieldCreateTime))
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
		if !m.CreateTime.IsZero() {
			item.CreateTime = m.CreateTime
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
			playlist.PrivacyEQ(playlist.PrivacyPUBLIC),
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
			Privacy:     string(p.Privacy),
		}
		if !p.AddDate.IsZero() {
			item.CreateTime = p.AddDate
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
		IsPublic:    ent.Privacy == playlist.PrivacyPUBLIC,
		CreateTime:  ent.AddDate,
		UpdateTime:  ent.AddDate,
		MediaItems:  []string{},
	}
}

func mapChannel(ent *entity.Channel) *biz.Channel {
	if ent == nil {
		return nil
	}

	var tags []string
	if ent.Tags != nil {
		tags = ent.Tags
	}

	var links []biz.ChannelLink
	if ent.Links != nil {
		links = convertSchemaLinksToBiz(ent.Links)
	}

	var categoryID *int64
	if ent.CategoryID != 0 {
		v := ent.CategoryID
		categoryID = &v
	}

	return &biz.Channel{
		ID:              ent.ID,
		Name:            ent.Name,
		Title:           ent.Title,
		Slug:            ent.Slug,
		Handle:          ent.Handle,
		Description:     ent.Description,
		Avatar:          ent.Avatar,
		Banner:          ent.Banner,
		BannerLogo:      ent.BannerLogo,
		ShortToken:      ent.ShortToken,
		Status:          string(ent.Status),
		Privacy:         string(ent.Privacy),
		IsVerified:      ent.IsVerified,
		Tags:            tags,
		CategoryID:      categoryID,
		SubscriberCount: ent.SubscriberCount,
		MediaCount:      ent.MediaCount,
		ArticleCount:    ent.ArticleCount,
		TotalViews:      ent.TotalViews,
		Links:           links,
		UserID:          ent.UserID,
		CreateTime:      ent.CreateTime,
		UpdateTime:      ent.UpdateTime,
	}
}

func convertBizLinksToSchema(links []biz.ChannelLink) []schema.ChannelLink {
	result := make([]schema.ChannelLink, len(links))
	for i, l := range links {
		result[i] = schema.ChannelLink{
			Type:     l.Type,
			Platform: l.Platform,
			URL:      l.URL,
			Title:    l.Title,
		}
	}
	return result
}

func convertSchemaLinksToBiz(links []schema.ChannelLink) []biz.ChannelLink {
	result := make([]biz.ChannelLink, len(links))
	for i, l := range links {
		result[i] = biz.ChannelLink{
			Type:     l.Type,
			Platform: l.Platform,
			URL:      l.URL,
			Title:    l.Title,
		}
	}
	return result
}

// systemConfigRepo implementation

// configCache provides an in-memory TTL cache for system settings.
var configCache = struct {
	sync.RWMutex
	items map[string]cacheItem
}{items: make(map[string]cacheItem)}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

const configCacheTTL = 5 * time.Minute

func (r *systemConfigRepo) Get(ctx context.Context, key string) (string, error) {
	// Check cache first
	configCache.RLock()
	if item, ok := configCache.items[key]; ok && time.Now().Before(item.expiresAt) {
		configCache.RUnlock()
		return item.value, nil
	}
	configCache.RUnlock()

	// Query DB
	ent, err := r.data.db.Setting.Query().Where(setting.KeyEQ(key)).Only(ctx)
	if err != nil {
		return "", fmt.Errorf("setting not found: %s: %w", key, err)
	}

	// Update cache
	configCache.Lock()
	configCache.items[key] = cacheItem{value: ent.Value, expiresAt: time.Now().Add(configCacheTTL)}
	configCache.Unlock()

	return ent.Value, nil
}

func (r *systemConfigRepo) Set(ctx context.Context, key, value string) error {
	// Upsert: try update first, then create
	_, err := r.data.db.Setting.Update().Where(setting.KeyEQ(key)).SetValue(value).Save(ctx)
	if err != nil {
		// Try create if update found nothing
		_, err = r.data.db.Setting.Create().
			SetKey(key).
			SetValue(value).
			SetCategory(setting.CategoryGeneral).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to set setting %s: %w", key, err)
		}
	}

	// Invalidate cache
	configCache.Lock()
	delete(configCache.items, key)
	configCache.Unlock()

	return nil
}

func (r *systemConfigRepo) ListByCategory(ctx context.Context, category string) (map[string]string, error) {
	items, err := r.data.db.Setting.Query().Where(setting.CategoryEQ(setting.Category(category))).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings for category %s: %w", category, err)
	}

	result := make(map[string]string, len(items))
	for _, item := range items {
		result[item.Key] = item.Value
	}
	return result, nil
}

func (r *systemConfigRepo) Delete(ctx context.Context, key string) error {
	err := r.data.db.Setting.DeleteOneID(key).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete setting %s: %w", key, err)
	}

	// Invalidate cache
	configCache.Lock()
	delete(configCache.items, key)
	configCache.Unlock()

	return nil
}

// channelUserRepo implementation

func (r *channelUserRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	ent, err := r.data.db.User.Query().Where(user.UsernameEQ(username)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found: %s: %w", username, err)
	}
	return &biz.User{
		ID:          ent.ID,
		Username:    ent.Username,
		Name:        ent.Name,
		Logo:        ent.Logo,
		Description: ent.Description,
		CreateTime:  ent.CreateTime,
	}, nil
}
