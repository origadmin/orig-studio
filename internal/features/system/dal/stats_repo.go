/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Stats repository - provides dashboard and system statistics
 */

package dal

import (
	"context"
	"time"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/comment"
	"origadmin/application/origstudio/internal/dal/entity/encodingtask"
	"origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/dal/entity/subscription"
	"origadmin/application/origstudio/internal/dal/entity/user"
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
	VideoCount      int `json:"video_count"`
	AudioCount      int `json:"audio_count"`
	ImageCount      int `json:"image_count"`
	PublicCount     int `json:"public_count"`
	PrivateCount    int `json:"private_count"`
	EncodingPending int `json:"encoding_pending"`
	EncodingFailed  int `json:"encoding_failed"`
}

// GetMediaStats gets media statistics
func (r *StatsRepo) GetMediaStats(ctx context.Context) (*MediaStats, error) {
	stats := &MediaStats{}

	total, err := r.db.Media.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

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

	return stats, nil
}
