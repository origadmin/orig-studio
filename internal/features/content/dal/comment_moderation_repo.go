package dal

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/comment"
	"origadmin/application/origstudio/internal/data/entity/commentreport"
	"origadmin/application/origstudio/internal/features/content/biz"
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
	query := r.data.db.Comment.Query()
	if mediaID != "" {
		query = query.Where(comment.MediaIDEQ(mediaID))
	}
	if status != "" {
		// Normalize status to uppercase for DB query (frontend may send lowercase)
		upperStatus := strings.ToUpper(status)
		query = query.Where(comment.StatusEQ(comment.Status(upperStatus)))
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
		WithParent().
		WithReplies().
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
	blockedQuery := r.data.db.Comment.Query().Where(comment.StatusEQ(comment.StatusBLOCKED))
	reportedPendingQuery := r.data.db.Comment.Query().Where(comment.ReportCountGT(0))

	if mediaID != "" {
		pendingQuery.Where(comment.MediaIDEQ(mediaID))
		approvedQuery.Where(comment.MediaIDEQ(mediaID))
		rejectedQuery.Where(comment.MediaIDEQ(mediaID))
		blockedQuery.Where(comment.MediaIDEQ(mediaID))
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

	blockedCount, err := blockedQuery.Count(ctx)
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
		Blocked:         blockedCount,
		Total:           pendingCount + approvedCount + rejectedCount + blockedCount,
		ReportedPending: reportedPendingCount,
	}, nil
}

func (r *commentModerationRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.Comment.DeleteOneID(id).Exec(ctx)
}

// ListAdminComments returns comments with optional tree structure and report status filtering.
// When tree=true, only root-level comments are paginated, and children are loaded via WithReplies.
// When reportStatus is set, comments are filtered by their report status.
func (r *commentModerationRepo) ListAdminComments(ctx context.Context, mediaID string, status string, reportStatus string, tree bool, page, pageSize int) ([]*biz.CommentModerationItem, int, error) {
	query := r.data.db.Comment.Query()

	// Apply media filter
	if mediaID != "" {
		query = query.Where(comment.MediaIDEQ(mediaID))
	}

	// Apply status filter
	if status != "" {
		upperStatus := strings.ToUpper(status)
		query = query.Where(comment.StatusEQ(comment.Status(upperStatus)))
	}

	// Apply report status filter using subqueries
	switch strings.ToLower(reportStatus) {
	case "reported":
		// Comments with at least one report
		query = query.Where(comment.ReportCountGT(0))
	case "pending_reports":
		// Comments that have at least one PENDING report
		query = query.Where(
			comment.HasReportsWith(commentreport.StatusEQ(commentreport.StatusPENDING)),
		)
	case "reviewed_reports":
		// Comments that have reports but none are PENDING
		query = query.Where(
			comment.ReportCountGT(0),
			comment.Not(
				comment.HasReportsWith(commentreport.StatusEQ(commentreport.StatusPENDING)),
			),
		)
	case "no_reports":
		// Comments with no reports
		query = query.Where(comment.ReportCountEQ(0))
	}

	if tree {
		// In tree mode, only fetch root-level comments (no parent)
		query = query.Where(comment.Not(comment.HasParent()))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Build the query with eager loading
	if tree {
		query = query.
			WithUser().
			WithMedia().
			WithReplies(func(q *entity.CommentQuery) {
				q.WithUser().WithMedia().Order(entity.Asc(comment.FieldAddDate))
			}).
			Order(entity.Desc(comment.FieldAddDate))
	} else {
		query = query.
			WithUser().
			WithMedia().
			WithParent().
			WithReplies().
			Order(entity.Desc(comment.FieldAddDate))
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.CommentModerationItem, len(ents))
	for i, ent := range ents {
		if tree {
			res[i] = mapCommentToTreeItem(ent, 0)
		} else {
			res[i] = mapCommentToModerationItem(ent)
		}
	}
	return res, total, nil
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

// UpdateStatusByComment updates all reports for a comment that match fromStatus to toStatus.
// Returns the number of reports updated.
func (r *commentReportRepo) UpdateStatusByComment(ctx context.Context, commentID string, fromStatus string, toStatus string) (int, error) {
	n, err := r.data.db.CommentReport.Update().
		Where(
			commentreport.CommentID(commentID),
			commentreport.StatusEQ(commentreport.Status(fromStatus)),
		).
		SetStatus(commentreport.Status(toStatus)).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// CountPendingByComment counts the number of PENDING reports for a comment.
func (r *commentReportRepo) CountPendingByComment(ctx context.Context, commentID string) (int, error) {
	return r.data.db.CommentReport.Query().
		Where(
			commentreport.CommentID(commentID),
			commentreport.StatusEQ(commentreport.StatusPENDING),
		).
		Count(ctx)
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
		LikeCount:   ent.LikeCount,
	}

	if ent.ModeratedBy != "" {
		item.ModeratedBy = &ent.ModeratedBy
	}

	if !ent.ModeratedAt.IsZero() {
		item.ModeratedAt = &ent.ModeratedAt
	}

	if ent.Edges.User != nil {
		item.Username = ent.Edges.User.Username
		item.Avatar = ent.Edges.User.Logo
	}

	if ent.Edges.Media != nil {
		item.MediaTitle = ent.Edges.Media.Title
	}

	if len(ent.Edges.Replies) > 0 {
		item.ReplyCount = len(ent.Edges.Replies)
		item.HasReplies = true
	}

	// Set parent_id from the parent edge
	if ent.Edges.Parent != nil {
		item.ParentID = ent.Edges.Parent.ID
	}

	// Check for pending reports
	if len(ent.Edges.Reports) > 0 {
		for _, r := range ent.Edges.Reports {
			if r.Status == commentreport.StatusPENDING {
				item.HasPendingReports = true
				break
			}
		}
	}

	return item
}

// mapCommentToTreeItem maps a Comment entity to a CommentModerationItem with tree structure.
// It recursively maps children and calculates depth.
func mapCommentToTreeItem(ent *entity.Comment, depth int) *biz.CommentModerationItem {
	item := &biz.CommentModerationItem{
		ID:          ent.ID,
		Text:        ent.Text,
		Status:      string(ent.Status),
		MediaID:     ent.MediaID,
		UserID:      ent.UserID,
		ReportCount: ent.ReportCount,
		AddDate:     ent.AddDate,
		LikeCount:   ent.LikeCount,
		Depth:       depth,
		HasReplies:  len(ent.Edges.Replies) > 0,
	}

	if ent.ModeratedBy != "" {
		item.ModeratedBy = &ent.ModeratedBy
	}

	if !ent.ModeratedAt.IsZero() {
		item.ModeratedAt = &ent.ModeratedAt
	}

	if ent.Edges.User != nil {
		item.Username = ent.Edges.User.Username
		item.Avatar = ent.Edges.User.Logo
	}

	if ent.Edges.Media != nil {
		item.MediaTitle = ent.Edges.Media.Title
	}

	if len(ent.Edges.Replies) > 0 {
		item.ReplyCount = len(ent.Edges.Replies)
		item.Children = make([]*biz.CommentModerationItem, len(ent.Edges.Replies))
		for i, reply := range ent.Edges.Replies {
			item.Children[i] = mapCommentToTreeItem(reply, depth+1)
		}
	}

	// Set parent_id from the parent edge
	if ent.Edges.Parent != nil {
		item.ParentID = ent.Edges.Parent.ID
	}

	// Check for pending reports
	if len(ent.Edges.Reports) > 0 {
		for _, r := range ent.Edges.Reports {
			if r.Status == commentreport.StatusPENDING {
				item.HasPendingReports = true
				break
			}
		}
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
		Status:      string(ent.Status),
		CreateTime:   ent.CreateTime,
	}

	if ent.Edges.Reporter != nil {
		item.Username = ent.Edges.Reporter.Username
	}

	return item
}
