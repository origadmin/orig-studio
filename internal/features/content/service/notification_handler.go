package service

import (
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/features/content/biz"
)

// NotificationHandler handles /api/v1/notifications routes.
type NotificationHandler struct {
	uc  *biz.NotificationUseCase
	jwt *auth.Manager
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(uc *biz.NotificationUseCase, jwt *auth.Manager) *NotificationHandler {
	return &NotificationHandler{uc: uc, jwt: jwt}
}

func (h *NotificationHandler) RegisterRoutes(r http2.Router) {
	notifs := r.Group("/notifications")
	{
		// Protected routes — all notification operations require auth

		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		notifs.GET("", server.WithJWTCtx(h.jwt, h.listNotifications()))
		notifs.POST("", server.WithJWTCtx(h.jwt, h.createNotification()))
		notifs.POST("/read-all", server.WithJWTCtx(h.jwt, h.markAllRead()))
		notifs.GET("/unread-count", server.WithJWTCtx(h.jwt, h.unreadCount()))

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		notifs.POST("/:id/read", server.WithJWTCtx(h.jwt, h.markAsRead()))
	}
}

// listNotifications returns notifications for the authenticated user,
// ordered by most recent, with pagination support.
func (h *NotificationHandler) listNotifications() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists || val == nil {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}
		claims := val.(*auth.Claims)

		limit, _ := strconv.Atoi(gc.Query("limit"))
		if limit == 0 {
			limit = 20
		}
		page, _ := strconv.Atoi(gc.Query("page"))
		if page == 0 {
			page = 1
		}
		// Normalize pagination parameters
		page, limit = repo.NormalizeHTTPPagination(page, limit)

		userID, _ := strconv.Atoi(claims.GetUserID())
		items, total, err := h.uc.ListUserNotifications(
			ctx.Request().Context(),
			userID,
			page,
			limit,
		)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		unread, _ := h.uc.GetUnreadCount(ctx.Request().Context(), userID)

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{
			"items":        items,
			"total":        total,
			"unread_count": unread,
			"page":         page,
			"page_size":    limit,
		}})
		return nil
	}
}

// createNotification creates a new notification.
// POST body: {"action": string, "notify": bool, "method": string, "user_id": int}
func (h *NotificationHandler) createNotification() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		var input struct {
			Action string `json:"action" binding:"required,max=30"`
			Notify bool   `json:"notify"`
			Method string `json:"method"`
			UserID int    `json:"user_id"` // optional; defaults to current user
		}
		if err := gc.Bind(&input); err != nil {
			gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: err.Error()})
			return nil
		}

		targetUserID := input.UserID
		if targetUserID == 0 {
			val, exists := gc.Get("claims")
			if !exists || val == nil {
				gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "user_id required"})
				return nil
			}
			claims := val.(*auth.Claims)
			userID, _ := strconv.Atoi(claims.GetUserID())
			targetUserID = userID
		}

		n := &biz.Notification{
			Action: input.Action,
			Notify: input.Notify,
			Method: input.Method,
			UserID: targetUserID,
		}

		created, err := h.uc.CreateNotification(ctx.Request().Context(), n)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		server.CreatedCtx(ctx, created)
		return nil
	}
}

// markAsRead marks a specific notification as read.
func (h *NotificationHandler) markAsRead() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists || val == nil {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}
		claims := val.(*auth.Claims)

		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "Invalid ID"})
			return nil
		}

		userID, _ := strconv.Atoi(claims.GetUserID())
		err = h.uc.MarkAsRead(ctx.Request().Context(), id, userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "marked as read"}})
		return nil
	}
}

// markAllRead marks all notifications as read for the current user.
func (h *NotificationHandler) markAllRead() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists || val == nil {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}
		claims := val.(*auth.Claims)

		userID, _ := strconv.Atoi(claims.GetUserID())
		err := h.uc.MarkAllAsRead(ctx.Request().Context(), userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "all marked as read"}})
		return nil
	}
}

// unreadCount returns the count of unread notifications for the current user.
func (h *NotificationHandler) unreadCount() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists || val == nil {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}
		claims := val.(*auth.Claims)

		userID, _ := strconv.Atoi(claims.GetUserID())
		count, err := h.uc.GetUnreadCount(ctx.Request().Context(), userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"unread_count": count}})
		return nil
	}
}
