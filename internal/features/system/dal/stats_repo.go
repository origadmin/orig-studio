/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Stats repository - provides dashboard and system statistics
 */

package dal

import (
	"context"
	"math"
	"strconv"
	"time"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/category"
	"origadmin/application/origstudio/internal/data/entity/channel"
	"origadmin/application/origstudio/internal/data/entity/comment"
	"origadmin/application/origstudio/internal/data/entity/encodingtask"
	"origadmin/application/origstudio/internal/data/entity/media"
	"origadmin/application/origstudio/internal/data/entity/order"
	"origadmin/application/origstudio/internal/data/entity/playlist"
	"origadmin/application/origstudio/internal/data/entity/subscription"
	"origadmin/application/origstudio/internal/data/entity/tag"
	"origadmin/application/origstudio/internal/data/entity/user"
)

// StatsRepo provides statistics for dashboard and system
type StatsRepo struct {
	db *entity.Client
}

// NewStatsRepo creates a new Stats repository
func NewStatsRepo(db *entity.Client) *StatsRepo {
	return &StatsRepo{db: db}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalUsers      int `json:"total_users"`
	TotalMedia      int `json:"total_media"`
	TotalViews      int `json:"total_views"`
	NewUsersToday   int `json:"new_users_today"`
	NewMediaToday   int `json:"new_media_today"`
	ViewsToday      int `json:"views_today"`
	EncodingPending int `json:"encoding_pending"`
	EncodingFailed  int `json:"encoding_failed"`
}

// GetDashboardStats gets all dashboard statistics
func (r *StatsRepo) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Total users
	totalUsers, err := r.db.User.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers

	// Total media
	totalMedia, err := r.db.Media.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalMedia = totalMedia

	// Total views (sum of all media view counts)
	mediaList, err := r.db.Media.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	totalViews := int64(0)
	for _, m := range mediaList {
		totalViews += m.ViewCount
	}
	stats.TotalViews = int(totalViews)

	// New users today
	today := time.Now().Truncate(24 * time.Hour)
	newUsersToday, err := r.db.User.Query().
		Where(user.DateAddedGTE(today)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.NewUsersToday = newUsersToday

	// New media today
	newMediaToday, err := r.db.Media.Query().
		Where(media.CreateTimeGTE(today)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.NewMediaToday = newMediaToday

	// Encoding pending
	encodingPending, err := r.db.EncodingTask.Query().
		Where(encodingtask.StatusEQ("pending")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EncodingPending = encodingPending

	// Encoding failed
	encodingFailed, err := r.db.EncodingTask.Query().
		Where(encodingtask.StatusEQ("failed")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EncodingFailed = encodingFailed

	return stats, nil
}

// ExtendedDashboardStats provides the full admin dashboard stats
// including media-by-type, users-by-role, and today's deltas.
type ExtendedDashboardStats struct {
	TotalUsers       int                    `json:"total_users"`
	TotalMedia       int                    `json:"total_media"`
	TotalViews       int64                  `json:"total_views"`
	TotalChannels    int                    `json:"total_channels"`
	PendingReviews   int                    `json:"pending_reviews"`
	TotalComments    int                    `json:"total_comments"`
	TotalSubscribers int                    `json:"total_subscribers"`
	NewUsersToday    int                    `json:"new_users_today"`
	NewMediaToday    int                    `json:"new_media_today"`
	NewCommentsToday int                    `json:"new_comments_today"`
	NewSubsToday     int                    `json:"new_subscribers_today"`
	MediaByType      map[string]int         `json:"media_by_type"`
	UsersByRole      map[string]int         `json:"users_by_role"`
}

// GetExtendedDashboardStats gets the full admin dashboard statistics.
// This method consolidates all the entity.Client queries that were previously
// done directly in the admin service handler.
func (r *StatsRepo) GetExtendedDashboardStats(ctx context.Context) (*ExtendedDashboardStats, error) {
	stats := &ExtendedDashboardStats{}
	today := time.Now().Truncate(24 * time.Hour)

	// Total users
	stats.TotalUsers, _ = r.db.User.Query().Count(ctx)

	// Total media
	stats.TotalMedia, _ = r.db.Media.Query().Count(ctx)

	// Total views
	mediaList, err := r.db.Media.Query().All(ctx)
	if err == nil {
		for _, m := range mediaList {
			stats.TotalViews += m.ViewCount
		}
	}

	// Total channels
	stats.TotalChannels, _ = r.db.Channel.Query().Count(ctx)

	// Pending reviews (media with review_status pending)
	stats.PendingReviews, _ = r.db.Media.Query().Where(media.ReviewStatusEQ("pending")).Count(ctx)

	// Total comments
	stats.TotalComments, _ = r.db.Comment.Query().Count(ctx)

	// Total subscribers
	stats.TotalSubscribers, _ = r.db.Subscription.Query().Count(ctx)

	// New users today
	stats.NewUsersToday, _ = r.db.User.Query().Where(user.DateAddedGTE(today)).Count(ctx)

	// New media today
	stats.NewMediaToday, _ = r.db.Media.Query().Where(media.CreateTimeGTE(today)).Count(ctx)

	// New comments today
	stats.NewCommentsToday, _ = r.db.Comment.Query().Where(comment.CreateTimeGTE(today)).Count(ctx)

	// New subscribers today
	stats.NewSubsToday, _ = r.db.Subscription.Query().Where(subscription.CreateTimeGTE(today)).Count(ctx)

	// Media by type
	videoCount, _ := r.db.Media.Query().Where(media.TypeEQ("video")).Count(ctx)
	imageCount, _ := r.db.Media.Query().Where(media.TypeEQ("image")).Count(ctx)
	audioCount, _ := r.db.Media.Query().Where(media.TypeEQ("audio")).Count(ctx)
	otherMediaCount := stats.TotalMedia - videoCount - imageCount - audioCount
	stats.MediaByType = map[string]int{
		"video": videoCount,
		"image": imageCount,
		"audio": audioCount,
		"other": otherMediaCount,
	}

	// Users by role
	adminCount, _ := r.db.User.Query().Where(user.RoleEQ("admin")).Count(ctx)
	editorCount, _ := r.db.User.Query().Where(user.RoleEQ("editor")).Count(ctx)
	regularCount, _ := r.db.User.Query().Where(user.RoleEQ("user")).Count(ctx)
	stats.UsersByRole = map[string]int{
		"admin":  adminCount,
		"editor": editorCount,
		"user":   regularCount,
	}

	return stats, nil
}

// MediaStats represents media statistics
type MediaStats struct {
	Total           int `json:"total"`
	TotalViews      int `json:"total_views"`
	VideoCount      int `json:"video_count"`
	AudioCount      int `json:"audio_count"`
	ImageCount      int `json:"image_count"`
	PublicCount     int `json:"public_count"`
	PrivateCount    int `json:"private_count"`
	EncodingPending int   `json:"encoding_pending"`
	EncodingFailed  int   `json:"encoding_failed"`
	StorageUsed     int64 `json:"storage_used"` // sum of media file sizes in bytes
}

// GetMediaStats gets media statistics
func (r *StatsRepo) GetMediaStats(ctx context.Context) (*MediaStats, error) {
	stats := &MediaStats{}

	total, err := r.db.Media.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// Total views (sum of all media view counts)
	mediaList, err := r.db.Media.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	var totalViews int64
	var storageUsed int64
	for _, m := range mediaList {
		totalViews += m.ViewCount
		// Size is an int64 byte-count stored as a string in the Media schema.
		if bytes, err := strconv.ParseInt(m.Size, 10, 64); err == nil {
			storageUsed += bytes
		}
	}
	stats.TotalViews = int(totalViews)
	stats.StorageUsed = storageUsed

	videoCount, err := r.db.Media.Query().Where(media.TypeEQ("video")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.VideoCount = videoCount

	audioCount, err := r.db.Media.Query().Where(media.TypeEQ("audio")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.AudioCount = audioCount

	imageCount, err := r.db.Media.Query().Where(media.TypeEQ("image")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ImageCount = imageCount

	publicCount, err := r.db.Media.Query().Where(media.PrivacyEQ(media.PrivacyPUBLIC)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PublicCount = publicCount

	privateCount, err := r.db.Media.Query().Where(media.PrivacyEQ(media.PrivacyPRIVATE)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PrivateCount = privateCount

	encodingPending, err := r.db.EncodingTask.Query().Where(encodingtask.StatusEQ("pending")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EncodingPending = encodingPending

	encodingFailed, err := r.db.EncodingTask.Query().Where(encodingtask.StatusEQ("failed")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EncodingFailed = encodingFailed

	return stats, nil
}

// UserStats represents user statistics
type UserStats struct {
	Total        int `json:"total"`
	ActiveToday  int `json:"active_today"`
	NewToday     int `json:"new_today"`
	AdminCount   int `json:"admin_count"`
	EditorCount  int `json:"editor_count"`
	RegularCount int `json:"regular_count"`
	// BUG-211: ActiveTotal = count of users with status=ACTIVE (not "logged in today").
	// Page-level stats card uses this for the "active" card.
	ActiveTotal int `json:"active_total"`
}

// MediaByDateItem represents media count/views for a single date
type MediaByDateItem struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// TopCategoryItem represents a top category with its media count
type TopCategoryItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// TopCreatorItem represents a top creator with their media count
type TopCreatorItem struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Count    int    `json:"count"`
}

// TopMediaItem represents a top media with its view count
type TopMediaItem struct {
	Title string `json:"title"`
	Views int64  `json:"views"`
}

// GetActiveUsersToday returns the count of users who logged in today
func (r *StatsRepo) GetActiveUsersToday(ctx context.Context) int {
	today := time.Now().Truncate(24 * time.Hour)
	count, err := r.db.User.Query().Where(user.LastLoginGTE(today)).Count(ctx)
	if err != nil {
		return 0
	}
	return count
}

// GetMediaByDate returns media counts grouped by creation date for the last N days
func (r *StatsRepo) GetMediaByDate(ctx context.Context, days int) []MediaByDateItem {
	now := time.Now().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -days+1)

	var result []MediaByDateItem
	// Walk each day and count media created on that day
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		next := d.AddDate(0, 0, 1)
		count, err := r.db.Media.Query().
			Where(media.CreateTimeGTE(d), media.CreateTimeLTE(next)).
			Count(ctx)
		if err != nil {
			count = 0
		}
		result = append(result, MediaByDateItem{
			Date:  d.Format("2006-01-02"),
			Count: count,
		})
	}
	return result
}

// GetViewsByDate returns view counts (sum of media view_count) grouped by media creation date for the last N days
func (r *StatsRepo) GetViewsByDate(ctx context.Context, days int) []MediaByDateItem {
	now := time.Now().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -days+1)

	var result []MediaByDateItem
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		next := d.AddDate(0, 0, 1)
		mediaList, err := r.db.Media.Query().
			Where(media.CreateTimeGTE(d), media.CreateTimeLTE(next)).
			All(ctx)
		if err != nil {
			result = append(result, MediaByDateItem{Date: d.Format("2006-01-02"), Count: 0})
			continue
		}
		var totalViews int64
		for _, m := range mediaList {
			totalViews += m.ViewCount
		}
		result = append(result, MediaByDateItem{
			Date:  d.Format("2006-01-02"),
			Count: int(totalViews),
		})
	}
	return result
}

// GetUserByDate returns user counts grouped by creation date for the last N days
func (r *StatsRepo) GetUserByDate(ctx context.Context, days int) []MediaByDateItem {
	now := time.Now().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -days+1)

	var result []MediaByDateItem
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		next := d.AddDate(0, 0, 1)
		count, err := r.db.User.Query().
			Where(user.DateAddedGTE(d), user.DateAddedLT(next)).
			Count(ctx)
		if err != nil {
			count = 0
		}
		result = append(result, MediaByDateItem{
			Date:  d.Format("2006-01-02"),
			Count: count,
		})
	}
	return result
}

// GetTopCategories returns the top N categories by media count
func (r *StatsRepo) GetTopCategories(ctx context.Context, limit int) []TopCategoryItem {
	ents, err := r.db.Category.Query().
		Where(category.MediaCountGT(0)).
		Order(entity.Desc(category.FieldMediaCount)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil
	}
	result := make([]TopCategoryItem, len(ents))
	for i, ent := range ents {
		result[i] = TopCategoryItem{
			Name:  ent.Name,
			Count: ent.MediaCount,
		}
	}
	return result
}

// GetTopCreators returns the top N creators (users) by media count
func (r *StatsRepo) GetTopCreators(ctx context.Context, limit int) []TopCreatorItem {
	ents, err := r.db.User.Query().
		Where(user.MediaCountGT(0)).
		Order(entity.Desc(user.FieldMediaCount)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil
	}
	result := make([]TopCreatorItem, len(ents))
	for i, ent := range ents {
		result[i] = TopCreatorItem{
			Username: ent.Username,
			Name:     ent.Name,
			Count:    ent.MediaCount,
		}
	}
	return result
}

// GetTopMedia returns the top N media by view count
func (r *StatsRepo) GetTopMedia(ctx context.Context, limit int) []TopMediaItem {
	ents, err := r.db.Media.Query().
		Where(media.ViewCountGT(0)).
		Order(entity.Desc(media.FieldViewCount)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil
	}
	result := make([]TopMediaItem, len(ents))
	for i, ent := range ents {
		result[i] = TopMediaItem{
			Title: ent.Title,
			Views: ent.ViewCount,
		}
	}
	return result
}

// GetUserStats gets user statistics
func (r *StatsRepo) GetUserStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{}

	total, err := r.db.User.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	today := time.Now().Truncate(24 * time.Hour)
	activeToday, err := r.db.User.Query().Where(user.LastLoginGTE(today)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveToday = activeToday

	newToday, err := r.db.User.Query().Where(user.DateAddedGTE(today)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.NewToday = newToday

	adminCount, err := r.db.User.Query().Where(user.RoleEQ("admin")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.AdminCount = adminCount

	editorCount, err := r.db.User.Query().Where(user.RoleEQ("editor")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EditorCount = editorCount

	regularCount, err := r.db.User.Query().Where(user.RoleEQ("user")).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.RegularCount = regularCount

	// BUG-211: ActiveTotal = users with status=ACTIVE (enum value, uppercase).
	activeTotal, err := r.db.User.Query().Where(user.StatusEQ(user.StatusACTIVE)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveTotal = activeTotal

	return stats, nil
}

// RevenueStats represents revenue statistics.
// Amounts are in major currency units (e.g. dollars); the service converts them
// to minor units (cents) when mapping to the proto response.
type RevenueStats struct {
	TotalRevenue        float64
	SubscriptionRevenue float64
	AdRevenue           float64 // always 0: schema has no ad-revenue data source (documented gap)
	DailyRevenue        []MediaByDateItem
}

// GetRevenueStats computes revenue from paid/completed orders.
//   - total_revenue: sum of all paid/completed order amounts
//   - subscription_revenue: sum of paid/completed order amounts linked to a plan (PlanID set)
//   - ad_revenue: 0 (no ad data source in schema)
//   - daily_revenue: paid/completed order amounts bucketed by creation date (last 30 days)
func (r *StatsRepo) GetRevenueStats(ctx context.Context) (*RevenueStats, error) {
	stats := &RevenueStats{}
	now := time.Now().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -29)

	daily := make(map[string]float64)
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		daily[d.Format("2006-01-02")] = 0
	}

	orders, err := r.db.Order.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		if o.Status != order.StatusPaid && o.Status != order.StatusCompleted {
			continue
		}
		stats.TotalRevenue += o.Amount
		if o.PlanID != "" {
			stats.SubscriptionRevenue += o.Amount
		}
		if o.CreateTime.IsZero() {
			continue
		}
		day := o.CreateTime.Truncate(24 * time.Hour).Format("2006-01-02")
		if _, ok := daily[day]; ok {
			daily[day] += o.Amount
		}
	}

	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		day := d.Format("2006-01-02")
		stats.DailyRevenue = append(stats.DailyRevenue, MediaByDateItem{
			Date:  day,
			Count: int(math.Round(daily[day] * 100)),
		})
	}

	return stats, nil
}

// PromotionStats aggregates promotion feature counters for the admin
// /admin/stats/promotion endpoint. All tables may be empty — the endpoint
// returns honest zeros instead of fabricated marketing figures.
type PromotionStats struct {
	TotalChannels       int `json:"total_channels"`
	ActiveChannels      int `json:"active_channels"`
	TotalSubscriptions  int `json:"total_subscriptions"`
	ActiveSubscriptions int `json:"active_subscriptions"`
	SentToday           int `json:"sent_today"`
	TotalLogs           int `json:"total_logs"`
	TotalTasks          int `json:"total_tasks"`
	TotalTemplates      int `json:"total_templates"`
}

// GetPromotionStats aggregates promotion counters.
func (r *StatsRepo) GetPromotionStats(ctx context.Context) (*PromotionStats, error) {
	stats := &PromotionStats{}

	channels, err := r.db.PromotionChannel.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalChannels = len(channels)
	for _, ch := range channels {
		if ch.IsActive {
			stats.ActiveChannels++
		}
		stats.SentToday += ch.SentToday
	}

	subs, err := r.db.PromotionSubscription.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalSubscriptions = len(subs)
	for _, s := range subs {
		if s.IsActive {
			stats.ActiveSubscriptions++
		}
	}

	if logs, err := r.db.PromotionLog.Query().Count(ctx); err == nil {
		stats.TotalLogs = logs
	}
	if tasks, err := r.db.PromotionTask.Query().Count(ctx); err == nil {
		stats.TotalTasks = tasks
	}
	if templates, err := r.db.PromotionTemplate.Query().Count(ctx); err == nil {
		stats.TotalTemplates = templates
	}

	return stats, nil
}

// ChannelStats represents channel statistics (BUG-211).
// Backs the independent /admin/stats/channels endpoint so the channels page no
// longer reuses the list endpoint with page_size=1000 and a frontend reduce.
type ChannelStats struct {
	Total            int   `json:"total"`
	TotalSubscribers int64 `json:"total_subscribers"`
	VerifiedCount    int   `json:"verified_count"`
	PendingCount     int   `json:"pending_count"`
}

// GetChannelStats gets channel statistics.
//   - Total: count of all channels
//   - TotalSubscribers: SUM(subscriber_count) across all channels
//   - VerifiedCount: channels with is_verified=true (schema/channel.go has no
//     status='verified'; verification is the is_verified bool flag)
//   - PendingCount: channels with status=PENDING_REVIEW (the schema's pending state)
func (r *StatsRepo) GetChannelStats(ctx context.Context) (*ChannelStats, error) {
	stats := &ChannelStats{}

	total, err := r.db.Channel.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// SUM(subscriber_count): load all channels and sum the denormalized counter.
	channels, err := r.db.Channel.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	var totalSubs int64
	for _, ch := range channels {
		totalSubs += ch.SubscriberCount
	}
	stats.TotalSubscribers = totalSubs

	verified, err := r.db.Channel.Query().Where(channel.IsVerifiedEQ(true)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.VerifiedCount = verified

	pending, err := r.db.Channel.Query().Where(channel.StatusEQ(channel.StatusPENDING_REVIEW)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PendingCount = pending

	return stats, nil
}

// TagStats represents tag statistics (BUG-211).
type TagStats struct {
	Total       int `json:"total"`
	ActiveCount int `json:"active_count"`
	UnusedCount int `json:"unused_count"`
	ColorAlerts int `json:"color_alerts"`
}

// GetTagStats gets tag statistics.
//   - Total: count of all tags
//   - ActiveCount: tags with status=ACTIVE (enum, uppercase)
//   - UnusedCount: tags with media_count=0
//   - ColorAlerts: tags with NULL color (Optional field; unset == NULL)
func (r *StatsRepo) GetTagStats(ctx context.Context) (*TagStats, error) {
	stats := &TagStats{}

	total, err := r.db.Tag.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	active, err := r.db.Tag.Query().Where(tag.StatusEQ(tag.StatusACTIVE)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveCount = active

	unused, err := r.db.Tag.Query().Where(tag.MediaCountEQ(0)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.UnusedCount = unused

	colorAlerts, err := r.db.Tag.Query().Where(tag.ColorIsNil()).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ColorAlerts = colorAlerts

	return stats, nil
}

// PlaylistStats represents playlist statistics (BUG-211).
type PlaylistStats struct {
	Total       int   `json:"total"`
	PublicCount int   `json:"public_count"`
	TotalItems  int   `json:"total_items"`
	TotalViews  int64 `json:"total_views"`
}

// GetPlaylistStats gets playlist statistics.
//   - Total: count of all playlists
//   - PublicCount: playlists with privacy=PUBLIC (schema has no is_public; the
//     privacy enum PUBLIC/PRIVATE/UNLISTED/PAID is the authoritative visibility)
//   - TotalItems: SUM(media_count) across all playlists (denormalized per-playlist count)
//   - TotalViews: SUM over MediaPlaylist associations of the linked media's view_count.
//     A media in N playlists is counted N times, consistent with TotalItems counting
//     per association.
func (r *StatsRepo) GetPlaylistStats(ctx context.Context) (*PlaylistStats, error) {
	stats := &PlaylistStats{}

	total, err := r.db.Playlist.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	public, err := r.db.Playlist.Query().Where(playlist.PrivacyEQ(playlist.PrivacyPUBLIC)).Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PublicCount = public

	// TotalItems = SUM(media_count) across all playlists.
	playlists, err := r.db.Playlist.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	var totalItems int
	for _, pl := range playlists {
		totalItems += pl.MediaCount
	}
	stats.TotalItems = totalItems

	// TotalViews = SUM over MediaPlaylist rows of the linked media's view_count.
	// Build media_id -> view_count once to avoid N+1 (load only media that appear in
	// any playlist).
	mediaPlaylists, err := r.db.MediaPlaylist.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	mediaView := make(map[string]int64, len(mediaPlaylists))
	mediaIDs := make([]string, 0, len(mediaPlaylists))
	for _, mp := range mediaPlaylists {
		if _, ok := mediaView[mp.MediaID]; !ok {
			mediaView[mp.MediaID] = 0
			mediaIDs = append(mediaIDs, mp.MediaID)
		}
	}
	if len(mediaIDs) > 0 {
		medias, err := r.db.Media.Query().Where(media.IDIn(mediaIDs...)).All(ctx)
		if err != nil {
			return nil, err
		}
		for _, m := range medias {
			mediaView[m.ID] = m.ViewCount
		}
	}
	var totalViews int64
	for _, mp := range mediaPlaylists {
		totalViews += mediaView[mp.MediaID]
	}
	stats.TotalViews = totalViews

	return stats, nil
}
