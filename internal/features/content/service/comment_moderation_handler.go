package service

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/infra/auth"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/helpers/repo"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/server"
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

func (h *CommentModerationHandler) RegisterRoutes(apiV1 *gin.RouterGroup) {
	adminComments := apiV1.Group("/admin/comments")
	adminComments.Use(server.JWTMiddleware(h.jwtMgr), server.AdminMiddleware(h.jwtMgr))
	{
		r := ginadapter.NewStdRouterAdapter(adminComments)
		r.GET("", h.listAdminComments())
		r.GET("/stats", h.getCommentStats())
		r.DELETE("/:id", h.deleteComment())
		r.POST("/:id/approve", h.approveComment())
		r.POST("/:id/reject", h.rejectComment())
		r.POST("/:id/block", h.blockComment())
		r.POST("/:id/unblock", h.unblockComment())
		r.POST("/:id/dismiss-reports", h.dismissReports())
		r.POST("/batch-approve", h.batchApproveComments())
		r.POST("/batch-reject", h.batchRejectComments())
		r.GET("/:id/reports", h.getCommentReports())
	}

	// Report comment (authenticated user)
	reportR := ginadapter.NewStdRouterAdapter(apiV1)
	reportR.POST("/comments/:id/report", server.WithJWT(h.jwtMgr, h.reportComment()))
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

func (h *CommentModerationHandler) deleteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.DeleteComment(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
			} else {
				server.Fail(gc, server.ErrInternal, err.Error())
			}
			return
		}

		server.OK(gc, gin.H{"id": id, "deleted": true})
	}
}

func (h *CommentModerationHandler) listAdminComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		status := gc.Query("status")
		mediaID := gc.Query("media_id")
		reportStatus := gc.Query("report_status")
		tree := gc.Query("tree") == "true"
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.moderationUC.ListAdminComments(ctx, mediaID, status, reportStatus, tree, page, pageSize)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		result := make([]CommentListItem, len(items))
		for i, item := range items {
			result[i] = mapBizItemToDTO(item)
		}

		server.Page(gc, result, int64(total), page, pageSize)
	}
}

func (h *CommentModerationHandler) getCommentStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		mediaID := gc.Query("media_id")

		stats, err := h.moderationUC.GetCommentStats(ctx, mediaID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, CommentStatsDTO{
			Pending:         stats.Pending,
			Approved:        stats.Approved,
			Rejected:        stats.Rejected,
			Blocked:         stats.Blocked,
			Total:           stats.Total,
			ReportedPending: stats.ReportedPending,
		})
	}
}

func (h *CommentModerationHandler) approveComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.ModerateComment(ctx, id, "approve", adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				server.Fail(gc, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
			} else {
				server.Fail(gc, server.ErrInternal, err.Error())
			}
			return
		}

		commentObj, getErr := h.moderationUC.GetComment(ctx, id)
		if getErr != nil {
			server.OK(gc, ModerationResultDTO{
				ID:          id,
				Status:      "APPROVED",
				ModeratedBy: adminID,
				ModeratedAt: time.Now().Format(time.RFC3339),
			})
			return
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

		server.OK(gc, resp)
	}
}

func (h *CommentModerationHandler) rejectComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.ModerateComment(ctx, id, "reject", adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				server.Fail(gc, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
			} else {
				server.Fail(gc, server.ErrInternal, err.Error())
			}
			return
		}

		commentObj, getErr := h.moderationUC.GetComment(ctx, id)
		if getErr != nil {
			server.OK(gc, ModerationResultDTO{
				ID:          id,
				Status:      "REJECTED",
				ModeratedBy: adminID,
				ModeratedAt: time.Now().Format(time.RFC3339),
			})
			return
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

		server.OK(gc, resp)
	}
}

func (h *CommentModerationHandler) batchApproveComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if len(req.IDs) == 0 {
			server.Fail(gc, server.ErrBadRequest, "ids is required")
			return
		}
		if len(req.IDs) > 100 {
			server.Fail(gc, server.ErrBadRequest, "batch size cannot exceed 100")
			return
		}

		updatedCount, skippedCount, err := h.moderationUC.BatchModerateComments(ctx, req.IDs, "approve", adminID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, BatchResultDTO{
			UpdatedCount: updatedCount,
			SkippedCount: skippedCount,
			Message:      "batch approve completed",
		})
	}
}

func (h *CommentModerationHandler) batchRejectComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if len(req.IDs) == 0 {
			server.Fail(gc, server.ErrBadRequest, "ids is required")
			return
		}
		if len(req.IDs) > 100 {
			server.Fail(gc, server.ErrBadRequest, "batch size cannot exceed 100")
			return
		}

		updatedCount, skippedCount, err := h.moderationUC.BatchModerateComments(ctx, req.IDs, "reject", adminID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, BatchResultDTO{
			UpdatedCount: updatedCount,
			SkippedCount: skippedCount,
			Message:      "batch reject completed",
		})
	}
}

func (h *CommentModerationHandler) getCommentReports() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		reports, err := h.moderationUC.GetCommentReports(ctx, id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
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

		server.OK(gc, CommentReportsResultDTO{
			CommentID:   id,
			ReportCount: len(reports),
			Reports:     reportItems,
		})
	}
}

func (h *CommentModerationHandler) reportComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		userID := claims.GetUserID()

		var req struct {
			Reason      string `json:"reason"`
			Description string `json:"description"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if req.Reason == "" {
			server.Fail(gc, server.ErrBadRequest, "reason is required")
			return
		}

		validReasons := map[string]bool{
			"SPAM":          true,
			"HARASSMENT":    true,
			"INAPPROPRIATE": true,
			"OTHER":         true,
		}
		if !validReasons[req.Reason] {
			server.Fail(gc, server.ErrBadRequest, "invalid report reason, must be one of: SPAM, HARASSMENT, INAPPROPRIATE, OTHER")
			return
		}

		reportCount, _, err := h.moderationUC.ReportComment(ctx, id, userID, req.Reason, req.Description)
		if err != nil {
			if strings.Contains(err.Error(), "already reported") {
				server.Fail(gc, server.ErrConflict, err.Error())
				return
			}
			if strings.Contains(err.Error(), "cannot report your own comment") {
				server.Fail(gc, server.ErrBadRequest, err.Error())
				return
			}
			if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
				return
			}
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, ReportResultDTO{
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

func (h *CommentModerationHandler) blockComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		result, err := h.moderationUC.BlockComment(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				server.Fail(gc, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
			} else {
				server.Fail(gc, server.ErrInternal, err.Error())
			}
			return
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

		server.OK(gc, resp)
	}
}

func (h *CommentModerationHandler) unblockComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		result, err := h.moderationUC.UnblockComment(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				server.Fail(gc, server.ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
			} else {
				server.Fail(gc, server.ErrInternal, err.Error())
			}
			return
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

		server.OK(gc, resp)
	}
}

func (h *CommentModerationHandler) dismissReports() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		result, err := h.moderationUC.DismissReports(ctx, id, adminID)
		if err != nil {
			if strings.Contains(err.Error(), "failed to get comment") {
				server.Fail(gc, server.ErrCommentNotFound, "comment not found")
			} else {
				server.Fail(gc, server.ErrInternal, err.Error())
			}
			return
		}

		server.OK(gc, DismissReportsResultDTO{
			CommentID:      result.CommentID,
			DismissedCount: result.DismissedCount,
			ReportCount:    result.ReportCount,
			Message:        "reports dismissed",
		})
	}
}
