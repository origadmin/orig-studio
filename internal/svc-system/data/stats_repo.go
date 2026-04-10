/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Stats repository - provides dashboard and system statistics
 */

package data

import (
	"context"
	"time"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/encodingtask"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/user"
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
	// Note: This is a simplified version. In production, you might want to use a separate view counter table
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
		Where(media.CreatedAtGTE(today)).
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

	// Total media
	total, err := r.db.Media.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// Video count
	videoCount, err := r.db.Media.Query().
		Where(media.TypeEQ("video")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.VideoCount = videoCount

	// Audio count
	audioCount, err := r.db.Media.Query().
		Where(media.TypeEQ("audio")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.AudioCount = audioCount

	// Image count
	imageCount, err := r.db.Media.Query().
		Where(media.TypeEQ("image")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ImageCount = imageCount

	// Public count (privacy = 1)
	publicCount, err := r.db.Media.Query().
		Where(media.PrivacyEQ(1)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PublicCount = publicCount

	// Private count (privacy = 2)
	privateCount, err := r.db.Media.Query().
		Where(media.PrivacyEQ(2)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PrivateCount = privateCount

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

	// Total users
	total, err := r.db.User.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// Active today (users who logged in today)
	today := time.Now().Truncate(24 * time.Hour)
	activeToday, err := r.db.User.Query().
		Where(user.LastLoginGTE(today)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveToday = activeToday

	// New today
	newToday, err := r.db.User.Query().
		Where(user.DateAddedGTE(today)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.NewToday = newToday

	// Admin count
	adminCount, err := r.db.User.Query().
		Where(user.RoleEQ("admin")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.AdminCount = adminCount

	// Editor count
	editorCount, err := r.db.User.Query().
		Where(user.RoleEQ("editor")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.EditorCount = editorCount

	// Regular count (role = "user")
	regularCount, err := r.db.User.Query().
		Where(user.RoleEQ("user")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.RegularCount = regularCount

	return stats, nil
}
