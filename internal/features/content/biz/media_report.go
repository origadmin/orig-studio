package biz

import (
	"context"
	"fmt"
	"time"

	systembiz "origadmin/application/origstudio/internal/features/system/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type MediaReportItem struct {
	ID          string    `json:"id"`
	MediaID     string    `json:"media_id"`
	ReporterID  string    `json:"reporter_id"`
	Reason      string    `json:"reason"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	Username    string    `json:"username,omitempty"`
}

type MediaReportRepo interface {
	Create(ctx context.Context, mediaID string, reporterID string, reason string, description string) (*MediaReportItem, error)
	Exists(ctx context.Context, reporterID, mediaID string) (bool, error)
}

type MediaReportModerationRepo interface {
	IncrementReportCount(ctx context.Context, mediaID string) (int, error)
	GetMediaOwnerID(ctx context.Context, mediaID string) (string, error)
}

type MediaReportUseCase struct {
	reportRepo     MediaReportRepo
	moderationRepo MediaReportModerationRepo
	configProvider systembiz.ConfigProvider
	log            *log.Helper
}

func NewMediaReportUseCase(
	reportRepo MediaReportRepo,
	moderationRepo MediaReportModerationRepo,
	configProvider systembiz.ConfigProvider,
	logger log.Logger,
) *MediaReportUseCase {
	return &MediaReportUseCase{
		reportRepo:     reportRepo,
		moderationRepo: moderationRepo,
		configProvider: configProvider,
		log:            log.NewHelper(log.With(logger, "module", "media_report.biz")),
	}
}

func (uc *MediaReportUseCase) ReportMedia(ctx context.Context, mediaID string, reporterID string, reason string, description string) (int, string, error) {
	validReasons := map[string]bool{
		"SPAM":          true,
		"HARASSMENT":    true,
		"INAPPROPRIATE": true,
		"OTHER":         true,
	}
	if !validReasons[reason] {
		return 0, "", fmt.Errorf("invalid report reason")
	}

	ownerID, err := uc.moderationRepo.GetMediaOwnerID(ctx, mediaID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get media")
	}
	if ownerID == reporterID {
		return 0, "", fmt.Errorf("cannot report your own media")
	}

	exists, err := uc.reportRepo.Exists(ctx, reporterID, mediaID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to check existing report")
	}
	if exists {
		return 0, "", fmt.Errorf("you have already reported this media")
	}

	_, err = uc.reportRepo.Create(ctx, mediaID, reporterID, reason, description)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create report: %w", err)
	}

	reportCount, err := uc.moderationRepo.IncrementReportCount(ctx, mediaID)
	if err != nil {
		uc.log.Warnf("failed to increment report count for media %s: %v", mediaID, err)
	}

	return reportCount, "reported", nil
}
