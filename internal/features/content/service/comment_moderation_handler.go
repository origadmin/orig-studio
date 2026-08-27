package service

import (
	"strconv"
	"strings"
	"time"

	"origadmin/application/origstudio/internal/domain/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/server"
)

type CommentModerationHandler struct {
	moderationUC *contentbiz.CommentModerationUseCase
	jwtMgr       *auth.Manager
}

func NewCommentModerationHandler(moderationUC *contentbiz.CommentModerationUseCase, jwtMgr *auth.Manager) *CommentModerationHandler {
	return &CommentModerationHandler{
		moderationUC: moderationUC,
		jwtMgr:       jwtMgr,
	}
}

func (h *CommentModerationHandler) RegisterRoutes(r http2.Router) {
	adminComments := r.Group("/admin/comments")
	adminComments.Use(server.JWTMiddlewareCtx(h.jwtMgr), server.AdminMiddlewareCtx(h.jwtMgr))
	{
		adminComments.GET("", h.listAdminComments())
		adminComments.GET("/stats", h.getCommentStats())
		adminComments.DELETE("/:id", h.deleteComment())
		adminComments.POST("/:id/approve", h.approveComment())
		adminComments.POST("/:id/reject", h.rejectComment())
		adminComments.POST("/:id/block", h.blockComment())
		adminComments.POST("/:id/unblock", h.unblockComment())
		adminComments.POST("/:id/dismiss-reports", h.dismissReports())
		adminComments.POST("/batch-approve", h.batchApproveComments())
		adminComments.POST("/batch-reject", h.batchRejectComments())
		adminComments.GET("/:id/reports", h.getCommentReports())
	}

	// Report comment (authenticated user)
	r.POST("/comments/:id/report", server.WithJWTCtx(h.jwtMgr, h.reportComment()))
}

// CommentListItem is the DTO for a comment in admin list responses.
// Field names align with the frontend admin Comments page expectations (B087).
type CommentListItem struct {
	ID                string            `json:"id"`
	Content           string            `json:"content"`
	Status            string            `json:"status"`
	MediaID           string            `json:"media_id"`
	UserID            string            `json:"user_id"`
	Username          string            `json:"username,omitempty"`
	Avatar            string            `json:"avatar,omitempty"`
	LikeCount         int               `json:"like_count"`
	ReplyCount        int               `json:"reply_count"`
	ReportCount       int               `json:"report_count"`
	IsSpam            bool              `json:"is_spam"`
	CreateTime        string            `json:"create_time"`
	Media             *CommentMediaItem `json:"media,omitempty"`
	ModeratedBy       string            `json:"moderated_by,omitempty"`
	ModeratedAt       string            `json:"moderated_at,omitempty"`
	ParentID          string            `json:"parent_id,omitempty"`
	Depth             int               `json:"depth"`
	HasReplies        bool              `json:"has_replies"`
	Children          []CommentListItem `json:"children,omitempty"`
	HasPendingReports bool              `json:"has_pending_reports"`
}

// CommentMediaItem is the nested media object in admin comment list responses.
type CommentMediaItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// CommentStatsDTO is the DTO for comment statistics.
type CommentStatsDTO struct {
	Pending         int `json:"pending"`
	Approved        int `json:"approved"`
	Rejected        int `json:"rejected"`
	Blocked         int `json:"blocked"`
	Total           int `json:"total"`
	ReportedPending int `json:"reported_pending"`
}

// ModerationResultDTO is the DTO for moderation action results.
type ModerationResultDTO struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	ModeratedBy  string `json:"moderated_by"`
	ModeratedAt  string `json:"moderated_at"`
	ReportCount  int    `json:"report_count,omitempty"`
}

// BatchResultDTO is the DTO for batch moderation results.
type BatchResultDTO struct {
	UpdatedCount int    `json:"updated_count"`
	SkippedCount int    `json:"skipped_count"`
	Message      string `json:"message"`
}

// CommentReportDTO is the DTO for a comment report.
type CommentReportDTO struct {
	ID          string `json:"id"`
	CommentID   string `json:"comment_id"`
	ReporterID  string `json:"reporter_id"`
	Reason      string `json:"reason"`
	CreateTime   string `json:"create_time"`
	Description string `json:"description,omitempty"`
	Username    string `json:"username,omitempty"`
	Status      string `json:"status"`
}

// CommentReportsResultDTO is the DTO for comment reports response.
type CommentReportsResultDTO struct {
	CommentID   string              `json:"comment_id"`
	ReportCount int                 `json:"report_count"`
	Reports     []CommentReportDTO  `json:"reports"`
}

// ReportResultDTO is the DTO for report submission result.
type ReportResultDTO struct {
	Message     string `json:"message"`
	ReportCount int    `json:"report_count"`
	Status      string `json:"status"`
}

func (h *CommentModerationHandler) deleteComment() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.DeleteComment(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			}
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"id": id, "deleted": true})
	}
}

func (h *CommentModerationHandler) listAdminComments() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		status := ctx.QueryVar("status")
		mediaID := ctx.QueryVar("media_id")
		reportStatus := ctx.QueryVar("report_status")
		keyword := ctx.QueryVar("keyword")
		tree := ctx.QueryVar("tree") == "true"
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.moderationUC.ListAdminComments(ctx, mediaID, status, reportStatus, tree, keyword, page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		result := make([]CommentListItem, len(items))
		for i, item := range items {
			result[i] = mapBizItemToDTO(item)
		}

		server.PageCtx(ctx, result, int64(total), page, pageSize)
		return nil
	}
}

func (h *CommentModerationHandler) getCommentStats() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		mediaID := ctx.QueryVar("media_id")

		stats, err := h.moderationUC.GetCommentStats(ctx, mediaID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, CommentStatsDTO{
			Pending:         stats.Pending,
			Approved:        stats.Approved,
			Rejected:        stats.Rejected,
			Blocked:         stats.Blocked,
			Total:           stats.Total,
			ReportedPending: stats.ReportedPending,
		})
	}
}

func (h *CommentModerationHandler) approveComment() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.ModerateComment(ctx, id, "approve", adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			} else {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
		}

		commentObj, getErr := h.moderationUC.GetComment(ctx, id)
		if getErr != nil {
			return server.OKCtx(ctx, ModerationResultDTO{
				ID:          id,
				Status:      "APPROVED",
				ModeratedBy: adminID,
				ModeratedAt: time.Now().Format(time.RFC3339),
			})
		}

		resp := ModerationResultDTO{
			ID:          id,
			Status:      commentObj.Status,
			ModeratedBy: adminID,
			ModeratedAt: time.Now().Format(time.RFC3339),
			ReportCount: commentObj.ReportCount,
		}
		if commentObj.ModeratedAt != nil {
			resp.ModeratedAt = commentObj.ModeratedAt.Format(time.RFC3339)
		}
		if commentObj.ModeratedBy != nil {
			resp.ModeratedBy = *commentObj.ModeratedBy
		}

		return server.OKCtx(ctx, resp)
	}
}

func (h *CommentModerationHandler) rejectComment() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.ModerateComment(ctx, id, "reject", adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			} else {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
		}

		commentObj, getErr := h.moderationUC.GetComment(ctx, id)
		if getErr != nil {
			return server.OKCtx(ctx, ModerationResultDTO{
				ID:          id,
				Status:      "REJECTED",
				ModeratedBy: adminID,
				ModeratedAt: time.Now().Format(time.RFC3339),
			})
		}

		resp := ModerationResultDTO{
			ID:     id,
			Status: commentObj.Status,
		}
		if commentObj.ModeratedBy != nil {
			resp.ModeratedBy = *commentObj.ModeratedBy
		} else {
			resp.ModeratedBy = adminID
		}
		if commentObj.ModeratedAt != nil {
			resp.ModeratedAt = commentObj.ModeratedAt.Format(time.RFC3339)
		} else {
			resp.ModeratedAt = time.Now().Format(time.RFC3339)
		}

		return server.OKCtx(ctx, resp)
	}
}

func (h *CommentModerationHandler) batchApproveComments() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if len(req.IDs) == 0 {
			return server.FailCtx(ctx, server.ErrBadRequest, "ids is required")
		}
		if len(req.IDs) > 100 {
			return server.FailCtx(ctx, server.ErrBadRequest, "batch size cannot exceed 100")
		}

		updatedCount, skippedCount, err := h.moderationUC.BatchModerateComments(ctx, req.IDs, "approve", adminID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, BatchResultDTO{
			UpdatedCount: updatedCount,
			SkippedCount: skippedCount,
			Message:      "batch approve completed",
		})
	}
}

func (h *CommentModerationHandler) batchRejectComments() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if len(req.IDs) == 0 {
			return server.FailCtx(ctx, server.ErrBadRequest, "ids is required")
		}
		if len(req.IDs) > 100 {
			return server.FailCtx(ctx, server.ErrBadRequest, "batch size cannot exceed 100")
		}

		updatedCount, skippedCount, err := h.moderationUC.BatchModerateComments(ctx, req.IDs, "reject", adminID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, BatchResultDTO{
			UpdatedCount: updatedCount,
			SkippedCount: skippedCount,
			Message:      "batch reject completed",
		})
	}
}

func (h *CommentModerationHandler) getCommentReports() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		reports, err := h.moderationUC.GetCommentReports(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		reportItems := make([]CommentReportDTO, len(reports))
		for i, r := range reports {
			entry := CommentReportDTO{
				ID:         r.ID,
				CommentID:  r.CommentID,
				ReporterID: r.ReporterID,
				Reason:     r.Reason,
				CreateTime:  r.CreateTime.Format(time.RFC3339),
				Status:     strings.ToLower(r.Status),
			}
			if r.Description != "" {
				entry.Description = r.Description
			}
			if r.Username != "" {
				entry.Username = r.Username
			}
			reportItems[i] = entry
		}

		return server.OKCtx(ctx, CommentReportsResultDTO{
			CommentID:   id,
			ReportCount: len(reports),
			Reports:     reportItems,
		})
	}
}

func (h *CommentModerationHandler) reportComment() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID := claims.GetUserID()

		var req struct {
			Reason      string `json:"reason"`
			Description string `json:"description"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if req.Reason == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "reason is required")
		}

		validReasons := map[string]bool{
			"SPAM":            true,
			"HARASSMENT":      true,
			"INAPPROPRIATE":   true,
			"PLAYBACK_ERROR":  true,
			"SUBTITLE_ERROR":  true,
			"QUALITY_ISSUE":   true,
			"BROKEN_LINK":     true,
			"OTHER":           true,
		}
		if !validReasons[req.Reason] {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid report reason, must be one of: SPAM, HARASSMENT, INAPPROPRIATE, PLAYBACK_ERROR, SUBTITLE_ERROR, QUALITY_ISSUE, BROKEN_LINK, OTHER")
		}

		reportCount, _, err := h.moderationUC.ReportComment(ctx, id, userID, req.Reason, req.Description)
		if err != nil {
			if strings.Contains(err.Error(), "already reported") {
				return server.FailCtx(ctx, server.ErrConflict, err.Error())
			}
			if strings.Contains(err.Error(), "cannot report your own comment") {
				return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			}
			if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			}
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, ReportResultDTO{
			Message:     "report submitted",
			ReportCount: reportCount,
			Status:      "reported",
		})
	}
}

// mapBizItemToDTO converts a biz.CommentModerationItem to a CommentListItem DTO.
// It recursively maps children for tree structure.
func mapBizItemToDTO(item *contentbiz.CommentModerationItem) CommentListItem {
	entry := CommentListItem{
		ID:                item.ID,
		Content:           item.Text,
		Status:            strings.ToLower(item.Status),
		MediaID:           item.MediaID,
		UserID:            item.UserID,
		Username:          item.Username,
		Avatar:            item.Avatar,
		LikeCount:         item.LikeCount,
		ReplyCount:        item.ReplyCount,
		ReportCount:       item.ReportCount,
		IsSpam:            item.ReportCount >= 3,
		CreateTime:        item.AddDate.Format(time.RFC3339),
		ParentID:          item.ParentID,
		Depth:             item.Depth,
		HasReplies:        item.HasReplies,
		HasPendingReports: item.HasPendingReports,
	}
	if item.MediaID != "" || item.MediaTitle != "" {
		entry.Media = &CommentMediaItem{
			ID:    item.MediaID,
			Title: item.MediaTitle,
		}
	}
	if item.ModeratedBy != nil {
		entry.ModeratedBy = *item.ModeratedBy
	}
	if item.ModeratedAt != nil {
		entry.ModeratedAt = item.ModeratedAt.Format(time.RFC3339)
	}
	// Map children recursively for tree structure
	if len(item.Children) > 0 {
		entry.Children = make([]CommentListItem, len(item.Children))
		for i, child := range item.Children {
			entry.Children[i] = mapBizItemToDTO(child)
		}
	}
	return entry
}

// DismissReportsResultDTO is the DTO for dismiss reports result.
type DismissReportsResultDTO struct {
	CommentID      string `json:"comment_id"`
	DismissedCount int    `json:"dismissed_count"`
	ReportCount    int    `json:"report_count"`
	Message        string `json:"message"`
}

func (h *CommentModerationHandler) blockComment() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		result, err := h.moderationUC.BlockComment(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			} else {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
		}

		resp := ModerationResultDTO{
			ID:     id,
			Status: "blocked",
		}
		if result != nil {
			resp.Status = strings.ToLower(result.Status)
			resp.ReportCount = result.ReportCount
			if result.ModeratedBy != nil {
				resp.ModeratedBy = *result.ModeratedBy
			}
			if result.ModeratedAt != nil {
				resp.ModeratedAt = result.ModeratedAt.Format(time.RFC3339)
			}
		}

		return server.OKCtx(ctx, resp)
	}
}

func (h *CommentModerationHandler) unblockComment() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		result, err := h.moderationUC.UnblockComment(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			} else {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
		}

		resp := ModerationResultDTO{
			ID:     id,
			Status: "approved",
		}
		if result != nil {
			resp.Status = strings.ToLower(result.Status)
			resp.ReportCount = result.ReportCount
			if result.ModeratedBy != nil {
				resp.ModeratedBy = *result.ModeratedBy
			}
			if result.ModeratedAt != nil {
				resp.ModeratedAt = result.ModeratedAt.Format(time.RFC3339)
			}
		}

		return server.OKCtx(ctx, resp)
	}
}

func (h *CommentModerationHandler) dismissReports() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "comment ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		result, err := h.moderationUC.DismissReports(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "failed to get comment") {
				return server.FailCtx(ctx, server.ErrCommentNotFound, "comment not found")
			} else {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
		}

		return server.OKCtx(ctx, DismissReportsResultDTO{
			CommentID:      result.CommentID,
			DismissedCount: result.DismissedCount,
			ReportCount:    result.ReportCount,
			Message:        "reports dismissed",
		})
	}
}
