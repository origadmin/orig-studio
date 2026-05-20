package dal

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/mediareviewlog"
	"origadmin/application/origstudio/internal/features/media/biz"
)

type reviewLogRepo struct {
	db *entity.Client
}

func NewReviewLogRepo(db *entity.Client) biz.ReviewLogRepo {
	return &reviewLogRepo{db: db}
}

func (r *reviewLogRepo) Create(ctx context.Context, mediaID string, reviewerID string, action string, comment string, previousStatus string, newStatus string) (*biz.ReviewLog, error) {
	create := r.db.MediaReviewLog.Create().
		SetMediaID(mediaID).
		SetReviewerID(reviewerID).
		SetAction(action).
		SetPreviousStatus(previousStatus).
		SetNewStatus(newStatus)

	if comment != "" {
		create = create.SetComment(comment)
	}

	m, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create review log: %w", err)
	}

	return convertReviewLog(m), nil
}

func (r *reviewLogRepo) ListByMedia(ctx context.Context, mediaID string) ([]*biz.ReviewLog, error) {
	items, err := r.db.MediaReviewLog.Query().
		Where(mediareviewlog.MediaIDEQ(mediaID)).
		Order(entity.Desc(mediareviewlog.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list review logs: %w", err)
	}

	result := make([]*biz.ReviewLog, len(items))
	for i, item := range items {
		result[i] = convertReviewLog(item)
	}
	return result, nil
}

func convertReviewLog(m *entity.MediaReviewLog) *biz.ReviewLog {
	return &biz.ReviewLog{
		ID:             m.ID,
		MediaID:        m.MediaID,
		ReviewerID:     m.ReviewerID,
		Action:         m.Action,
		Comment:        m.Comment,
		PreviousStatus: m.PreviousStatus,
		NewStatus:      m.NewStatus,
		CreateTime:      m.CreateTime.Format("2006-01-02T15:04:05Z07:00"),
	}
}
