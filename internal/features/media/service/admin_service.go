package service

import (
	"context"
	"math"
	"strings"

	"github.com/origadmin/runtime/log"

	media "origadmin/application/origstudio/api/gen/v1/media"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemdto "origadmin/application/origstudio/internal/features/system/dto"
	"origadmin/application/origstudio/internal/features/system/dal"
	types "origadmin/application/origstudio/api/gen/v1/types"
)

type AdminService struct {
	media.UnimplementedAdminServiceServer
	statsRepo  *dal.StatsRepo
	settingUC  *systembiz.SettingUseCase
	log        *log.Helper
}

func NewAdminService(statsRepo *dal.StatsRepo, settingUC *systembiz.SettingUseCase, logger log.Logger) *AdminService {
	return &AdminService{
		statsRepo: statsRepo,
		settingUC: settingUC,
		log:       log.NewHelper(log.With(logger, "module", "media.service.admin")),
	}
}

func (s *AdminService) GetDashboardStats(ctx context.Context, req *media.GetDashboardStatsRequest) (*media.GetDashboardStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetDashboardStatsResponse{}, nil
	}

	stats, err := s.statsRepo.GetExtendedDashboardStats(ctx)
	if err != nil {
		s.log.Errorf("GetDashboardStats failed: %v", err)
		return &media.GetDashboardStatsResponse{}, nil
	}

	return &media.GetDashboardStatsResponse{
		TotalUsers:          int32(stats.TotalUsers),
		TotalMedias:         int32(stats.TotalMedia),
		TotalViews:          int32(stats.TotalViews),
		TotalChannels:       int32(stats.TotalChannels),
		PendingReviews:      int32(stats.PendingReviews),
		TotalComments:       int32(stats.TotalComments),
		TotalSubscribers:    int32(stats.TotalSubscribers),
		NewUsersToday:       int32(stats.NewUsersToday),
		NewMediaToday:       int32(stats.NewMediaToday),
		NewCommentsToday:    int32(stats.NewCommentsToday),
		NewSubscribersToday: int32(stats.NewSubsToday),
		MediaByType:         toInt32Map(stats.MediaByType),
		UsersByRole:         toInt32Map(stats.UsersByRole),
	}, nil
}

// toInt32Map converts a map[string]int to map[string]int32 for proto responses.
func toInt32Map(in map[string]int) map[string]int32 {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int32, len(in))
	for k, v := range in {
		out[k] = int32(v)
	}
	return out
}

func (s *AdminService) GetMediaStats(ctx context.Context, req *media.GetMediaStatsRequest) (*media.GetMediaStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetMediaStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetMediaStats(ctx)
	if err != nil {
		s.log.Errorf("GetMediaStats failed: %v", err)
		return &media.GetMediaStatsResponse{}, nil
	}
	days := periodDays(req.GetPeriod())
	daily := s.statsRepo.GetMediaByDate(ctx, days)
	return &media.GetMediaStatsResponse{
		TotalUploads: int32(stats.Total),
		TotalViews:   int32(stats.TotalViews),
		DailyStats:   toStatPoints(daily),
		StorageUsed:  stats.StorageUsed,
	}, nil
}

func (s *AdminService) GetUserStats(ctx context.Context, req *media.GetUserStatsRequest) (*media.GetUserStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetUserStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetUserStats(ctx)
	if err != nil {
		s.log.Errorf("GetUserStats failed: %v", err)
		return &media.GetUserStatsResponse{}, nil
	}
	days := periodDays(req.GetPeriod())
	daily := s.statsRepo.GetUserByDate(ctx, days)
	return &media.GetUserStatsResponse{
		TotalUsers:   int32(stats.Total),
		NewUsers:     int32(stats.NewToday),
		ActiveUsers:  int32(stats.ActiveToday),
		DailyStats:   toStatPoints(daily),
		ActiveTotal:  int32(stats.ActiveTotal),
		AdminCount:   int32(stats.AdminCount),
		EditorCount:  int32(stats.EditorCount),
	}, nil
}

func (s *AdminService) GetTrafficStats(ctx context.Context, req *media.GetTrafficStatsRequest) (*media.GetTrafficStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetTrafficStatsResponse{}, nil
	}
	days := periodDays(req.GetPeriod())
	// No bandwidth/request log exists in the schema; totals are 0 (documented gap).
	// Daily traffic is approximated by daily view counts (best available signal).
	daily := s.statsRepo.GetViewsByDate(ctx, days)
	return &media.GetTrafficStatsResponse{
		TotalBandwidth: 0,
		TotalRequests:  0,
		DailyStats:     toStatPoints(daily),
	}, nil
}

func (s *AdminService) GetRevenueStats(ctx context.Context, req *media.GetRevenueStatsRequest) (*media.GetRevenueStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetRevenueStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetRevenueStats(ctx)
	if err != nil {
		s.log.Errorf("GetRevenueStats failed: %v", err)
		return &media.GetRevenueStatsResponse{}, nil
	}
	return &media.GetRevenueStatsResponse{
		TotalRevenue:        int64(math.Round(stats.TotalRevenue * 100)),
		SubscriptionRevenue: int64(math.Round(stats.SubscriptionRevenue * 100)),
		AdRevenue:           0,
		DailyStats:          toStatPoints(stats.DailyRevenue),
	}, nil
}

// GetPromotionStats aggregates promotion counters (BUG-219).
func (s *AdminService) GetPromotionStats(ctx context.Context, req *media.GetPromotionStatsRequest) (*media.GetPromotionStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetPromotionStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetPromotionStats(ctx)
	if err != nil {
		s.log.Errorf("GetPromotionStats failed: %v", err)
		return &media.GetPromotionStatsResponse{}, nil
	}
	return &media.GetPromotionStatsResponse{
		TotalChannels:       int32(stats.TotalChannels),
		ActiveChannels:      int32(stats.ActiveChannels),
		TotalSubscriptions:  int32(stats.TotalSubscriptions),
		ActiveSubscriptions: int32(stats.ActiveSubscriptions),
		SentToday:           int32(stats.SentToday),
		TotalLogs:           int32(stats.TotalLogs),
		TotalTasks:          int32(stats.TotalTasks),
		TotalTemplates:      int32(stats.TotalTemplates),
	}, nil
}

// GetChannelStats returns channel page statistics (BUG-211).
// Replaces the channels page's reuse of the list endpoint with page_size=1000.
func (s *AdminService) GetChannelStats(ctx context.Context, req *media.GetChannelStatsRequest) (*media.GetChannelStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetChannelStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetChannelStats(ctx)
	if err != nil {
		s.log.Errorf("GetChannelStats failed: %v", err)
		return &media.GetChannelStatsResponse{}, nil
	}
	return &media.GetChannelStatsResponse{
		Total:            int32(stats.Total),
		TotalSubscribers: int32(stats.TotalSubscribers),
		VerifiedCount:    int32(stats.VerifiedCount),
		PendingCount:     int32(stats.PendingCount),
	}, nil
}

// GetTagStats returns tag page statistics (BUG-211).
// Replaces the tags page's reuse of the list endpoint with page_size=1000.
func (s *AdminService) GetTagStats(ctx context.Context, req *media.GetTagStatsRequest) (*media.GetTagStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetTagStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetTagStats(ctx)
	if err != nil {
		s.log.Errorf("GetTagStats failed: %v", err)
		return &media.GetTagStatsResponse{}, nil
	}
	return &media.GetTagStatsResponse{
		Total:        int32(stats.Total),
		ActiveCount:  int32(stats.ActiveCount),
		UnusedCount:  int32(stats.UnusedCount),
		ColorAlerts:  int32(stats.ColorAlerts),
	}, nil
}

// GetPlaylistStats returns playlist page statistics (BUG-211).
// Replaces the playlists page's reuse of the list endpoint with page_size=1000.
func (s *AdminService) GetPlaylistStats(ctx context.Context, req *media.GetPlaylistStatsRequest) (*media.GetPlaylistStatsResponse, error) {
	if s.statsRepo == nil {
		return &media.GetPlaylistStatsResponse{}, nil
	}
	stats, err := s.statsRepo.GetPlaylistStats(ctx)
	if err != nil {
		s.log.Errorf("GetPlaylistStats failed: %v", err)
		return &media.GetPlaylistStatsResponse{}, nil
	}
	return &media.GetPlaylistStatsResponse{
		Total:       int32(stats.Total),
		PublicCount: int32(stats.PublicCount),
		TotalItems:  int32(stats.TotalItems),
		TotalViews:  stats.TotalViews,
	}, nil
}

// periodDays maps a stats period string to the number of days for the daily series.
func periodDays(period string) int {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "day":
		return 1
	case "week":
		return 7
	case "month":
		return 30
	case "year":
		return 365
	default:
		return 30
	}
}

// toStatPoints maps repo daily-items (Date + int Count) to proto StatPoint values.
func toStatPoints(items []dal.MediaByDateItem) []*types.StatPoint {
	pts := make([]*types.StatPoint, 0, len(items))
	for _, it := range items {
		pts = append(pts, &types.StatPoint{
			Date:  it.Date,
			Value: int64(it.Count),
		})
	}
	return pts
}

func (s *AdminService) GetPendingReviews(ctx context.Context, req *media.GetPendingReviewsRequest) (*media.GetPendingReviewsResponse, error) {
	return &media.GetPendingReviewsResponse{}, nil
}

func (s *AdminService) GetReviewHistory(ctx context.Context, req *media.GetReviewHistoryRequest) (*media.GetReviewHistoryResponse, error) {
	return &media.GetReviewHistoryResponse{}, nil
}

func (s *AdminService) GetReview(ctx context.Context, req *media.GetReviewRequest) (*media.GetReviewResponse, error) {
	return &media.GetReviewResponse{}, nil
}

func (s *AdminService) UpdateReview(ctx context.Context, req *media.UpdateReviewRequest) (*media.UpdateReviewResponse, error) {
	return &media.UpdateReviewResponse{}, nil
}

func (s *AdminService) BatchUpdateReviews(ctx context.Context, req *media.BatchUpdateReviewsRequest) (*media.BatchUpdateReviewsResponse, error) {
	return &media.BatchUpdateReviewsResponse{}, nil
}

func (s *AdminService) GetSettings(ctx context.Context, req *media.GetSettingsRequest) (*media.GetSettingsResponse, error) {
	if s.settingUC == nil {
		return &media.GetSettingsResponse{}, nil
	}
	return &media.GetSettingsResponse{
		Settings: s.settingUC.GetAll(ctx),
	}, nil
}

func (s *AdminService) UpdateSettings(ctx context.Context, req *media.UpdateSettingsRequest) (*media.UpdateSettingsResponse, error) {
	if s.settingUC == nil {
		return &media.UpdateSettingsResponse{}, nil
	}
	settings := req.GetSettings()
	for k, v := range settings {
		existing, err := s.settingUC.GetByKey(ctx, k)
		if err != nil || existing == nil {
			// BUG-139: new settings must carry legal Category/Type enum values —
			// ent Setting.CategoryValidator rejects the empty string and the
			// previous `_, _ =` swallow made PUT return 200 while nothing
			// persisted (BUG-138 #41 GET round-trip failure). Default to
			// feature/general + string so any key round-trips.
			_, _ = s.settingUC.Upsert(ctx, &systemdto.SettingDTO{
				Key:      k,
				Value:    v,
				Type:     systemdto.SettingTypeString,
				Category: categoryForSettingKey(k),
			})
			continue
		}
		// Existing setting: only update Value, preserve other fields
		existing.Value = v
		if _, err := s.settingUC.Upsert(ctx, existing); err != nil {
			s.log.Errorf("UpdateSettings: failed to upsert %s: %v", k, err)
		}
	}
	return &media.UpdateSettingsResponse{
		Settings: settings,
	}, nil
}

// categoryForSettingKey returns the settings category for a key. Feature
// toggles / *_mode enums (review_mode, comments_mode, downloads_mode) belong
// to the `feature` category; everything else defaults to `general`.
func categoryForSettingKey(key string) systemdto.SettingCategory {
	if strings.HasPrefix(key, "feature_") || strings.HasSuffix(key, "_mode") {
		return systemdto.SettingCategoryFeature
	}
	return systemdto.SettingCategoryGeneral
}
