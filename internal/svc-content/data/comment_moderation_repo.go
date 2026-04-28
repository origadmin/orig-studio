package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/comment"
	"origadmin/application/origcms/internal/data/entity/commentreport"
	"origadmin/application/origcms/internal/svc-content/biz"
)

type commentModerationRepo struct {
	data *Data
	log  *log.Helper
}

type commentReportRepo struct {
	data *Data
	log  *log.Helper
}

func NewCommentModerationRepo(data *Data, logger log.Logger) biz.CommentModerationRepo {
	return &commentModerationRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "comment_moderation.data")),
	}
}

func NewCommentReportRepo(data *Data, logger log.Logger) biz.CommentReportRepo {
	return &commentReportRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "comment_report.data")),
	}
}

func (r *commentModerationRepo) Get(ctx context.Context, id string) (*biz.CommentModerationItem, error) {
	ent, err := r.data.db.Comment.Query().
		Where(comment.ID(id)).
		WithUser().
		WithMedia().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapCommentToModerationItem(ent), nil
}

func (r *commentModerationRepo) UpdateStatus(ctx context.Context, commentID string, status string, moderatedBy string) error {
	builder := r.data.db.Comment.UpdateOneID(commentID).
		SetStatus(comment.Status(status)).
		SetModeratedAt(time.Now())

	if moderatedBy != "" {
		builder.SetModeratedBy(moderatedBy)
	} else {
		builder.ClearModeratedBy()
	}

	_, err := builder.Save(ctx)
	return err
}

func (r *commentModerationRepo) BatchUpdateStatus(ctx context.Context, commentIDs []string, status string, moderatedBy string) (int, error) {
	successCount := 0
	for _, id := range commentIDs {
		builder := r.data.db.Comment.UpdateOneID(id).
			SetStatus(comment.Status(status)).
			SetModeratedAt(time.Now())

		if moderatedBy != "" {
			builder.SetModeratedBy(moderatedBy)
		} else {
			builder.ClearModeratedBy()
		}

		_, err := builder.Save(ctx)
		if err != nil {
			r.log.Warnf("failed to update comment %s status: %v", id, err)
			continue
		}
		successCount++
	}
	return successCount, nil
}

func (r *commentModerationRepo) IncrementReportCount(ctx context.Context, commentID string) (int, error) {
	ent, err := r.data.db.Comment.UpdateOneID(commentID).
		AddReportCount(1).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return ent.ReportCount, nil
}

func (r *commentModerationRepo) ResetReportCount(ctx context.Context, commentID string) error {
	_, err := r.data.db.Comment.UpdateOneID(commentID).
		SetReportCount(0).
		Save(ctx)
	return err
}

func (r *commentModerationRepo) ListByMedia(ctx context.Context, mediaID string, status string, page, pageSize int) ([]*biz.CommentModerationItem, int, error) {
	query := r.data.db.Comment.Query().Where(comment.MediaIDEQ(mediaID))
	if status != "" {
		query.Where(comment.StatusEQ(comment.Status(status)))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(comment.FieldAddDate)).
		WithUser().
		WithMedia().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.CommentModerationItem, len(ents))
	for i, ent := range ents {
		res[i] = mapCommentToModerationItem(ent)
	}
	return res, total, nil
}

func (r *commentModerationRepo) ListPending(ctx context.Context, mediaID string, page, pageSize int) ([]*biz.CommentModerationItem, int, error) {
	query := r.data.db.Comment.Query().Where(comment.StatusEQ(comment.StatusPENDING))
	if mediaID != "" {
		query.Where(comment.MediaIDEQ(mediaID))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(comment.FieldAddDate)).
		WithUser().
		WithMedia().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.CommentModerationItem, len(ents))
	for i, ent := range ents {
		res[i] = mapCommentToModerationItem(ent)
	}
	return res, total, nil
}

func (r *commentModerationRepo) CountByStatus(ctx context.Context, mediaID string) (*biz.CommentStatusCounts, error) {
	pendingQuery := r.data.db.Comment.Query().Where(comment.StatusEQ(comment.StatusPENDING))
	approvedQuery := r.data.db.Comment.Query().Where(comment.StatusEQ(comment.StatusAPPROVED))
	rejectedQuery := r.data.db.Comment.Query().Where(comment.StatusEQ(comment.StatusREJECTED))
	reportedPendingQuery := r.data.db.Comment.Query().Where(comment.StatusEQ(comment.StatusPENDING), comment.ReportCountGT(0))

	if mediaID != "" {
		pendingQuery.Where(comment.MediaIDEQ(mediaID))
		approvedQuery.Where(comment.MediaIDEQ(mediaID))
		rejectedQuery.Where(comment.MediaIDEQ(mediaID))
		reportedPendingQuery.Where(comment.MediaIDEQ(mediaID))
	}

	pendingCount, err := pendingQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	approvedCount, err := approvedQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	rejectedCount, err := rejectedQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	reportedPendingCount, err := reportedPendingQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.CommentStatusCounts{
		Pending:         pendingCount,
		Approved:        approvedCount,
		Rejected:        rejectedCount,
		Total:           pendingCount + approvedCount + rejectedCount,
		ReportedPending: reportedPendingCount,
	}, nil
}

func (r *commentReportRepo) Create(ctx context.Context, commentID string, reporterID string, reason string, description string) (*biz.CommentReportItem, error) {
	ent, err := r.data.db.CommentReport.Create().
		SetCommentID(commentID).
		SetReporterID(reporterID).
		SetReason(commentreport.Reason(reason)).
		SetDescription(description).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapCommentReportToItem(ent), nil
}

func (r *commentReportRepo) Exists(ctx context.Context, reporterID, commentID string) (bool, error) {
	return r.data.db.CommentReport.Query().
		Where(commentreport.ReporterID(reporterID), commentreport.CommentID(commentID)).
		Exist(ctx)
}

func (r *commentReportRepo) ListByComment(ctx context.Context, commentID string) ([]*biz.CommentReportItem, error) {
	ents, err := r.data.db.CommentReport.Query().
		Where(commentreport.CommentID(commentID)).
		WithReporter().
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*biz.CommentReportItem, len(ents))
	for i, ent := range ents {
		res[i] = mapCommentReportToItem(ent)
	}
	return res, nil
}

func mapCommentToModerationItem(ent *entity.Comment) *biz.CommentModerationItem {
	item := &biz.CommentModerationItem{
		ID:          ent.ID,
		Text:        ent.Text,
		Status:      string(ent.Status),
		MediaID:     ent.MediaID,
		UserID:      ent.UserID,
		ReportCount: ent.ReportCount,
		AddDate:     ent.AddDate,
	}

	if ent.ModeratedBy != "" {
		item.ModeratedBy = &ent.ModeratedBy
	}

	if !ent.ModeratedAt.IsZero() {
		item.ModeratedAt = &ent.ModeratedAt
	}

	if ent.Edges.User != nil {
		item.Username = ent.Edges.User.Username
	}

	if ent.Edges.Media != nil {
		item.MediaTitle = ent.Edges.Media.Title
	}

	return item
}

func mapCommentReportToItem(ent *entity.CommentReport) *biz.CommentReportItem {
	item := &biz.CommentReportItem{
		ID:          ent.ID,
		CommentID:   ent.CommentID,
		ReporterID:  ent.ReporterID,
		Reason:      string(ent.Reason),
		Description: ent.Description,
		CreatedAt:   ent.CreatedAt,
	}

	if ent.Edges.Reporter != nil {
		item.Username = ent.Edges.Reporter.Username
	}

	return item
}
