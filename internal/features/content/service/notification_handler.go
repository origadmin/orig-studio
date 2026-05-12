package service

import (
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/features/content/biz"
)

type NotificationHandler struct {
	uc  *biz.NotificationUseCase
	jwt *auth.Manager
}

func NewNotificationHandler(uc *biz.NotificationUseCase, jwt *auth.Manager) *NotificationHandler {
	return &NotificationHandler{uc: uc, jwt: jwt}
}

func (h *NotificationHandler) RegisterRoutes(r http2.Router) {
	notifs := r.Group("/notifications")
	{
		notifs.GET("", server.WithJWTCtx(h.jwt, h.listNotifications()))
		notifs.POST("", server.WithJWTCtx(h.jwt, h.createNotification()))
		notifs.POST("/read-all", server.WithJWTCtx(h.jwt, h.markAllRead()))
		notifs.GET("/unread-count", server.WithJWTCtx(h.jwt, h.unreadCount()))
		notifs.DELETE("/:id", server.WithJWTCtx(h.jwt, h.deleteNotification()))
		notifs.POST("/:id/read", server.WithJWTCtx(h.jwt, h.markAsRead()))
	}
}

func getUserID(gc *gin.Context) (string, bool) {
	val, exists := gc.Get("claims")
	if !exists || val == nil {
		return "", false
	}
	claims := val.(*auth.Claims)
	return claims.GetUserID(), true
}

func (h *NotificationHandler) listNotifications() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		userID, ok := getUserID(gc)
		if !ok {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}

		limit, _ := strconv.Atoi(gc.Query("page_size"))
		if limit == 0 {
			limit = 20
		}
		page, _ := strconv.Atoi(gc.Query("page"))
		if page == 0 {
			page = 1
		}
		page, limit = repo.NormalizeHTTPPagination(page, limit)

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

func (h *NotificationHandler) createNotification() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		var input struct {
			Action string `json:"action" binding:"required,max=30"`
			Notify bool   `json:"notify"`
			Method string `json:"method"`
			UserID string `json:"user_id"`
			Title  string `json:"title" binding:"required,max=200"`
			Body   string `json:"body" binding:"required"`
		}
		if err := gc.Bind(&input); err != nil {
			gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: err.Error()})
			return nil
		}

		targetUserID := input.UserID
		if targetUserID == "" {
			id, ok := getUserID(gc)
			if !ok {
				gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "user_id required"})
				return nil
			}
			targetUserID = id
		}

		n := &biz.Notification{
			Action: input.Action,
			Notify: input.Notify,
			Method: input.Method,
			UserID: targetUserID,
			Title:  input.Title,
			Body:   input.Body,
		}

		created, err := h.uc.CreateNotification(ctx.Request().Context(), n)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		http2.Created(ctx, created)
		return nil
	}
}

func (h *NotificationHandler) markAsRead() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		userID, ok := getUserID(gc)
		if !ok {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}

		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "Invalid ID"})
			return nil
		}

		err = h.uc.MarkAsRead(ctx.Request().Context(), id, userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "marked as read"}})
		return nil
	}
}

func (h *NotificationHandler) markAllRead() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		userID, ok := getUserID(gc)
		if !ok {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}

		err := h.uc.MarkAllAsRead(ctx.Request().Context(), userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "all marked as read"}})
		return nil
	}
}

func (h *NotificationHandler) unreadCount() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		userID, ok := getUserID(gc)
		if !ok {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}

		count, err := h.uc.GetUnreadCount(ctx.Request().Context(), userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"unread_count": count}})
		return nil
	}
}

func (h *NotificationHandler) deleteNotification() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		userID, ok := getUserID(gc)
		if !ok {
			gc.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
			return nil
		}

		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			gc.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "Invalid ID"})
			return nil
		}

		err = h.uc.DeleteNotification(ctx.Request().Context(), id, userID)
		if err != nil {
			gc.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
			return nil
		}

		gc.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "deleted"}})
		return nil
	}
}