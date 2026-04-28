package biz

import (
	"context"
	"fmt"
	"time"

	systembiz "origadmin/application/origcms/internal/svc-system/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type CommentModerationItem struct {
	ID          string     `json:"id"`
	Text        string     `json:"text"`
	Status      string     `json:"status"`
	MediaID     string     `json:"media_id"`
	UserID      string     `json:"user_id"`
	ReportCount int        `json:"report_count"`
	ModeratedBy *string    `json:"moderated_by,omitempty"`
	ModeratedAt *time.Time `json:"moderated_at,omitempty"`
	AddDate     time.Time  `json:"add_date"`
	Username    string     `json:"username,omitempty"`
	MediaTitle  string     `json:"media_title,omitempty"`
}

type CommentReportItem struct {
	ID          string    `json:"id"`
	CommentID   string    `json:"comment_id"`
	ReporterID  string    `json:"reporter_id"`
	Reason      string    `json:"reason"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Username    string    `json:"username,omitempty"`
}

type CommentStatusCounts struct {
	Pending         int `json:"pending"`
	Approved        int `json:"approved"`
	Rejected        int `json:"rejected"`
	Total           int `json:"total"`
	ReportedPending int `json:"reported_pending"`
}

type CommentConfig struct {
	AutoApprove     bool     `json:"auto_approve"`
	ReportThreshold int      `json:"report_threshold"`
	ReportReasons   []string `json:"report_reasons"`
}

type CommentModerationRepo interface {
	Get(ctx context.Context, id string) (*CommentModerationItem, error)
	UpdateStatus(ctx context.Context, commentID string, status string, moderatedBy string) error
	BatchUpdateStatus(ctx context.Context, commentIDs []string, status string, moderatedBy string) (int, error)
	IncrementReportCount(ctx context.Context, commentID string) (int, error)
	ResetReportCount(ctx context.Context, commentID string) error
	ListByMedia(ctx context.Context, mediaID string, status string, page, pageSize int) ([]*CommentModerationItem, int, error)
	ListPending(ctx context.Context, mediaID string, page, pageSize int) ([]*CommentModerationItem, int, error)
	CountByStatus(ctx context.Context, mediaID string) (*CommentStatusCounts, error)
}

type CommentReportRepo interface {
	Create(ctx context.Context, commentID string, reporterID string, reason string, description string) (*CommentReportItem, error)
	Exists(ctx context.Context, reporterID, commentID string) (bool, error)
	ListByComment(ctx context.Context, commentID string) ([]*CommentReportItem, error)
}

type CommentModerationUseCase struct {
	commentRepo    CommentModerationRepo
	reportRepo     CommentReportRepo
	configProvider systembiz.ConfigProvider
	log            *log.Helper
}

func NewCommentModerationUseCase(
	commentRepo CommentModerationRepo,
	reportRepo CommentReportRepo,
	configProvider systembiz.ConfigProvider,
	logger log.Logger,
) *CommentModerationUseCase {
	return &CommentModerationUseCase{
		commentRepo:    commentRepo,
		reportRepo:     reportRepo,
		configProvider: configProvider,
		log:            log.NewHelper(log.With(logger, "module", "comment_moderation.biz")),
	}
}

var validModerationActions = map[string]bool{
	"approve": true,
	"reject":  true,
}

var validStatusTransitions = map[string]map[string]bool{
	"PENDING":  {"APPROVED": true, "REJECTED": true},
	"APPROVED": {"REJECTED": true},
	"REJECTED": {"APPROVED": true},
}

func (uc *CommentModerationUseCase) ModerateComment(ctx context.Context, commentID string, action string, moderatorID string) error {
	if !validModerationActions[action] {
		return fmt.Errorf("invalid moderation action: %s", action)
	}

	item, err := uc.commentRepo.Get(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	targetStatus := "APPROVED"
	if action == "reject" {
		targetStatus = "REJECTED"
	}

	allowed, ok := validStatusTransitions[item.Status]
	if !ok || !allowed[targetStatus] {
		return fmt.Errorf("invalid status transition from %s to %s", item.Status, targetStatus)
	}

	err = uc.commentRepo.UpdateStatus(ctx, commentID, targetStatus, moderatorID)
	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}

	if action == "approve" && item.ReportCount > 0 {
		if err := uc.commentRepo.ResetReportCount(ctx, commentID); err != nil {
			uc.log.Warnf("failed to reset report count for comment %s: %v", commentID, err)
		}
	}

	uc.log.Infof("comment %s moderated: %s -> %s by %s", commentID, item.Status, targetStatus, moderatorID)
	return nil
}

func (uc *CommentModerationUseCase) BatchModerateComments(ctx context.Context, commentIDs []string, action string, moderatorID string) (int, int, error) {
	if !validModerationActions[action] {
		return 0, 0, fmt.Errorf("invalid moderation action: %s", action)
	}

	if len(commentIDs) == 0 {
		return 0, 0, fmt.Errorf("comment IDs cannot be empty")
	}

	if len(commentIDs) > 100 {
		return 0, 0, fmt.Errorf("batch size cannot exceed 100")
	}

	targetStatus := "APPROVED"
	if action == "reject" {
		targetStatus = "REJECTED"
	}

	updatedCount, err := uc.commentRepo.BatchUpdateStatus(ctx, commentIDs, targetStatus, moderatorID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to batch update comment status: %w", err)
	}

	skippedCount := len(commentIDs) - updatedCount

	if action == "approve" {
		for _, id := range commentIDs {
			item, getErr := uc.commentRepo.Get(ctx, id)
			if getErr != nil {
				continue
			}
			if item.ReportCount > 0 {
				if resetErr := uc.commentRepo.ResetReportCount(ctx, id); resetErr != nil {
					uc.log.Warnf("failed to reset report count for comment %s: %v", id, resetErr)
				}
			}
		}
	}

	uc.log.Infof("batch moderated %d comments (action: %s) by %s, skipped: %d", updatedCount, action, moderatorID, skippedCount)
	return updatedCount, skippedCount, nil
}

var validReportReasons = map[string]bool{
	"SPAM":          true,
	"HARASSMENT":    true,
	"INAPPROPRIATE": true,
	"OTHER":         true,
}

func (uc *CommentModerationUseCase) ReportComment(ctx context.Context, commentID string, reporterID string, reason string, description string) (int, bool, error) {
	if !validReportReasons[reason] {
		return 0, false, fmt.Errorf("invalid report reason: %s", reason)
	}

	item, err := uc.commentRepo.Get(ctx, commentID)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get comment: %w", err)
	}

	if item.UserID == reporterID {
		return 0, false, fmt.Errorf("cannot report your own comment")
	}

	exists, err := uc.reportRepo.Exists(ctx, reporterID, commentID)
	if err != nil {
		return 0, false, fmt.Errorf("failed to check existing report: %w", err)
	}
	if exists {
		return 0, false, fmt.Errorf("already reported this comment")
	}

	_, err = uc.reportRepo.Create(ctx, commentID, reporterID, reason, description)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create report: %w", err)
	}

	reportCount, err := uc.commentRepo.IncrementReportCount(ctx, commentID)
	if err != nil {
		return 0, false, fmt.Errorf("failed to increment report count: %w", err)
	}

	config := uc.GetCommentConfig(ctx)

	var statusChanged bool
	if reportCount >= config.ReportThreshold && item.Status == "APPROVED" {
		err = uc.commentRepo.UpdateStatus(ctx, commentID, "PENDING", "")
		if err != nil {
			uc.log.Warnf("failed to change comment %s status to PENDING: %v", commentID, err)
		} else {
			statusChanged = true
			uc.log.Infof("comment %s status changed to PENDING due to report threshold (%d >= %d)", commentID, reportCount, config.ReportThreshold)
		}
	}

	return reportCount, statusChanged, nil
}

func (uc *CommentModerationUseCase) GetCommentReports(ctx context.Context, commentID string) ([]*CommentReportItem, error) {
	return uc.reportRepo.ListByComment(ctx, commentID)
}

func (uc *CommentModerationUseCase) ListByMedia(ctx context.Context, mediaID string, status string, page, pageSize int) ([]*CommentModerationItem, int, error) {
	return uc.commentRepo.ListByMedia(ctx, mediaID, status, page, pageSize)
}

func (uc *CommentModerationUseCase) GetComment(ctx context.Context, commentID string) (*CommentModerationItem, error) {
	return uc.commentRepo.Get(ctx, commentID)
}

func (uc *CommentModerationUseCase) GetCommentStats(ctx context.Context, mediaID string) (*CommentStatusCounts, error) {
	return uc.commentRepo.CountByStatus(ctx, mediaID)
}

func (uc *CommentModerationUseCase) GetInitialStatus(ctx context.Context) string {
	config := uc.GetCommentConfig(ctx)
	if config.AutoApprove {
		return "APPROVED"
	}
	return "PENDING"
}

func (uc *CommentModerationUseCase) GetCommentConfig(ctx context.Context) *CommentConfig {
	autoApprove := uc.configProvider.GetBool(ctx, "comment.auto_approve")
	if !autoApprove {
		val := uc.configProvider.Get(ctx, "comment.auto_approve")
		if val == "" {
			autoApprove = true
		}
	}

	reportThreshold := uc.configProvider.GetInt(ctx, "comment.report_threshold")
	if reportThreshold <= 0 {
		reportThreshold = 3
	}

	reportReasons := []string{"SPAM", "HARASSMENT", "INAPPROPRIATE", "OTHER"}

	return &CommentConfig{
		AutoApprove:     autoApprove,
		ReportThreshold: reportThreshold,
		ReportReasons:   reportReasons,
	}
}
