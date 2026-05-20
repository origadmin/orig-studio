package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/dal/entity/mediareport"
	"origadmin/application/origstudio/internal/dal/entity/user"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/pkg/idutil"
)

type mediaReportRepo struct {
	data *Data
	log  *log.Helper
}

type mediaReportModerationRepo struct {
	data *Data
	log  *log.Helper
}

func NewMediaReportRepo(data *Data, logger log.Logger) biz.MediaReportRepo {
	return &mediaReportRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "media_report.data")),
	}
}

func NewMediaReportModerationRepo(data *Data, logger log.Logger) biz.MediaReportModerationRepo {
	return &mediaReportModerationRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "media_report_moderation.data")),
	}
}

func (r *mediaReportRepo) Create(ctx context.Context, mediaID string, reporterID string, reason string, description string) (*biz.MediaReportItem, error) {
	ent, err := r.data.db.MediaReport.Create().
		SetID(idutil.GenUUID()).
		SetMediaID(mediaID).
		SetReporterID(reporterID).
		SetReason(mediareport.Reason(reason)).
		SetDescription(description).
		SetStatus(mediareport.StatusPENDING).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create media report: %w", err)
	}

	return &biz.MediaReportItem{
		ID:          ent.ID,
		MediaID:     ent.MediaID,
		ReporterID:  ent.ReporterID,
		Reason:      string(ent.Reason),
		Description: ent.Description,
		Status:      string(ent.Status),
		CreateTime:  ent.CreateTime,
	}, nil
}

func (r *mediaReportRepo) Exists(ctx context.Context, reporterID, mediaID string) (bool, error) {
	count, err := r.data.db.MediaReport.Query().
		Where(
			mediareport.And(
				mediareport.HasMediaWith(media.IDEQ(mediaID)),
				mediareport.HasReporterWith(user.IDEQ(reporterID)),
			),
		).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check existing media report: %w", err)
	}
	return count > 0, nil
}

func (r *mediaReportModerationRepo) IncrementReportCount(ctx context.Context, mediaID string) (int, error) {
	ent, err := r.data.db.Media.UpdateOneID(mediaID).
		AddReportedTimes(1).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment media report count: %w", err)
	}
	return ent.ReportedTimes, nil
}

func (r *mediaReportModerationRepo) GetMediaOwnerID(ctx context.Context, mediaID string) (string, error) {
	ent, err := r.data.db.Media.Query().
		Where(media.IDEQ(mediaID)).
		Only(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get media: %w", err)
	}
	return ent.UserID, nil
}
