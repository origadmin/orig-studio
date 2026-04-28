package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
)

type CommentModerationHandler struct {
	moderationUC *contentbiz.CommentModerationUseCase
	client       *entity.Client
	jwtMgr       *auth.Manager
}

func NewCommentModerationHandler(moderationUC *contentbiz.CommentModerationUseCase, client *entity.Client, jwtMgr *auth.Manager) *CommentModerationHandler {
	return &CommentModerationHandler{
		moderationUC: moderationUC,
		client:       client,
		jwtMgr:       jwtMgr,
	}
}

func (h *CommentModerationHandler) RegisterRoutes(apiV1 *gin.RouterGroup) {
	adminComments := apiV1.Group("/admin/comments")
	adminComments.Use(JWTMiddleware(h.jwtMgr), AdminMiddleware(h.jwtMgr))
	{
		adminComments.GET("", h.listAdminComments())
		adminComments.GET("/stats", h.getCommentStats())
		adminComments.POST("/:id/approve", h.approveComment())
		adminComments.POST("/:id/reject", h.rejectComment())
		adminComments.POST("/batch-approve", h.batchApproveComments())
		adminComments.POST("/batch-reject", h.batchRejectComments())
		adminComments.GET("/:id/reports", h.getCommentReports())
	}

	apiV1.POST("/comments/:id/report", JWTMiddleware(h.jwtMgr), h.reportComment())
}

func (h *CommentModerationHandler) listAdminComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		status := c.Query("status")
		mediaID := c.Query("media_id")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		items, total, err := h.moderationUC.ListByMedia(ctx, mediaID, status, page, pageSize)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		result := make([]gin.H, len(items))
		for i, item := range items {
			entry := gin.H{
				"id":           item.ID,
				"text":         item.Text,
				"status":       item.Status,
				"media_id":     item.MediaID,
				"user_id":      item.UserID,
				"report_count": item.ReportCount,
				"add_date":     item.AddDate.Format(time.RFC3339),
			}
			if item.ModeratedBy != nil {
				entry["moderated_by"] = *item.ModeratedBy
			}
			if item.ModeratedAt != nil {
				entry["moderated_at"] = item.ModeratedAt.Format(time.RFC3339)
			}
			if item.Username != "" {
				entry["username"] = item.Username
			}
			if item.MediaTitle != "" {
				entry["media_title"] = item.MediaTitle
			}
			result[i] = entry
		}

		OK(c, gin.H{
			"items":     result,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *CommentModerationHandler) getCommentStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		mediaID := c.Query("media_id")

		stats, err := h.moderationUC.GetCommentStats(ctx, mediaID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"pending":          stats.Pending,
			"approved":         stats.Approved,
			"rejected":         stats.Rejected,
			"total":            stats.Total,
			"reported_pending": stats.ReportedPending,
		})
	}
}

func (h *CommentModerationHandler) approveComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.ModerateComment(ctx, id, "approve", adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				Fail(c, ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				Fail(c, ErrCommentNotFound, "comment not found")
			} else {
				Fail(c, ErrInternal, err.Error())
			}
			return
		}

		commentObj, getErr := h.moderationUC.GetComment(ctx, id)
		if getErr != nil {
			OK(c, gin.H{
				"id":           id,
				"status":       "APPROVED",
				"moderated_by": adminID,
				"moderated_at": time.Now().Format(time.RFC3339),
			})
			return
		}

		resp := gin.H{
			"id":           id,
			"status":       commentObj.Status,
			"moderated_by": adminID,
			"moderated_at": time.Now().Format(time.RFC3339),
			"report_count": commentObj.ReportCount,
		}
		if commentObj.ModeratedAt != nil {
			resp["moderated_at"] = commentObj.ModeratedAt.Format(time.RFC3339)
		}
		if commentObj.ModeratedBy != nil {
			resp["moderated_by"] = *commentObj.ModeratedBy
		}

		OK(c, resp)
	}
}

func (h *CommentModerationHandler) rejectComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		err := h.moderationUC.ModerateComment(ctx, id, "reject", adminID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid status transition") {
				Fail(c, ErrBadRequest, err.Error())
			} else if strings.Contains(err.Error(), "failed to get comment") {
				Fail(c, ErrCommentNotFound, "comment not found")
			} else {
				Fail(c, ErrInternal, err.Error())
			}
			return
		}

		commentObj, getErr := h.moderationUC.GetComment(ctx, id)
		if getErr != nil {
			OK(c, gin.H{
				"id":           id,
				"status":       "REJECTED",
				"moderated_by": adminID,
				"moderated_at": time.Now().Format(time.RFC3339),
			})
			return
		}

		resp := gin.H{
			"id":     id,
			"status": commentObj.Status,
		}
		if commentObj.ModeratedBy != nil {
			resp["moderated_by"] = *commentObj.ModeratedBy
		} else {
			resp["moderated_by"] = adminID
		}
		if commentObj.ModeratedAt != nil {
			resp["moderated_at"] = commentObj.ModeratedAt.Format(time.RFC3339)
		} else {
			resp["moderated_at"] = time.Now().Format(time.RFC3339)
		}

		OK(c, resp)
	}
}

func (h *CommentModerationHandler) batchApproveComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if len(req.IDs) == 0 {
			Fail(c, ErrBadRequest, "ids is required")
			return
		}
		if len(req.IDs) > 100 {
			Fail(c, ErrBadRequest, "batch size cannot exceed 100")
			return
		}

		updatedCount, skippedCount, err := h.moderationUC.BatchModerateComments(ctx, req.IDs, "approve", adminID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"updated_count": updatedCount,
			"skipped_count": skippedCount,
			"message":       "batch approve completed",
		})
	}
}

func (h *CommentModerationHandler) batchRejectComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		var req struct {
			IDs []string `json:"ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if len(req.IDs) == 0 {
			Fail(c, ErrBadRequest, "ids is required")
			return
		}
		if len(req.IDs) > 100 {
			Fail(c, ErrBadRequest, "batch size cannot exceed 100")
			return
		}

		updatedCount, skippedCount, err := h.moderationUC.BatchModerateComments(ctx, req.IDs, "reject", adminID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"updated_count": updatedCount,
			"skipped_count": skippedCount,
			"message":       "batch reject completed",
		})
	}
}

func (h *CommentModerationHandler) getCommentReports() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "comment ID is required")
			return
		}

		reports, err := h.moderationUC.GetCommentReports(ctx, id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		reportItems := make([]gin.H, len(reports))
		for i, r := range reports {
			entry := gin.H{
				"id":          r.ID,
				"comment_id":  r.CommentID,
				"reporter_id": r.ReporterID,
				"reason":      r.Reason,
				"created_at":  r.CreatedAt.Format(time.RFC3339),
			}
			if r.Description != "" {
				entry["description"] = r.Description
			}
			if r.Username != "" {
				entry["username"] = r.Username
			}
			reportItems[i] = entry
		}

		OK(c, gin.H{
			"comment_id":   id,
			"report_count": len(reports),
			"reports":      reportItems,
		})
	}
}

func (h *CommentModerationHandler) reportComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "comment ID is required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}
		userID := claims.GetUserID()

		var req struct {
			Reason      string `json:"reason"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if req.Reason == "" {
			Fail(c, ErrBadRequest, "reason is required")
			return
		}

		validReasons := map[string]bool{
			"SPAM":          true,
			"HARASSMENT":    true,
			"INAPPROPRIATE": true,
			"OTHER":         true,
		}
		if !validReasons[req.Reason] {
			Fail(c, ErrBadRequest, "invalid report reason, must be one of: SPAM, HARASSMENT, INAPPROPRIATE, OTHER")
			return
		}

		reportCount, _, err := h.moderationUC.ReportComment(ctx, id, userID, req.Reason, req.Description)
		if err != nil {
			if strings.Contains(err.Error(), "already reported") {
				c.JSON(http.StatusConflict, Response[interface{}]{Code: ErrConflict, Message: err.Error()})
				return
			}
			if strings.Contains(err.Error(), "cannot report your own comment") {
				Fail(c, ErrBadRequest, err.Error())
				return
			}
			if strings.Contains(err.Error(), "failed to get comment") {
				Fail(c, ErrCommentNotFound, "comment not found")
				return
			}
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"message":      "report submitted",
			"report_count": reportCount,
			"status":       "reported",
		})
	}
}
