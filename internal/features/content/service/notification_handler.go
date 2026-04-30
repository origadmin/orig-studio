package service

import (
	"origadmin/application/origcms/internal/handler"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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

func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	notifs := r.Group("/notifications")
	{
		// Protected routes — all notification operations require auth
		// Note: We can't use Use() directly with the Router interface
		// We'll need to apply middleware to each route individually

		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		notifs.GET("", server.WithJWT(h.jwt, h.listNotifications))
		notifs.POST("", server.WithJWT(h.jwt, h.createNotification))
		notifs.POST("/read-all", server.WithJWT(h.jwt, h.markAllRead))
		notifs.GET("/unread-count", server.WithJWT(h.jwt, h.unreadCount))

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		notifs.POST("/:id/read", server.WithJWT(h.jwt, h.markAsRead))
	}
}

// listNotifications returns notifications for the authenticated user,
// ordered by most recent, with pagination support.
func (h *NotificationHandler) listNotifications(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	// Normalize pagination parameters
	page, limit = repo.NormalizeHTTPPagination(page, limit)

	userID, _ := strconv.Atoi(claims.GetUserID())
	items, total, err := h.uc.ListUserNotifications(
		r.Context(),
		userID,
		page,
		limit,
	)
	if err != nil {
		c.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
		return
	}

	unread, _ := h.uc.GetUnreadCount(r.Context(), userID)

	c.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{
		"items":        items,
		"total":        total,
		"unread_count": unread,
		"page":         page,
		"page_size":    limit,
	}})
}

// createNotification creates a new notification.
// POST body: {"action": string, "notify": bool, "method": string, "user_id": int}
func (h *NotificationHandler) createNotification(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	var input struct {
		Action string `json:"action" binding:"required,max=30"`
		Notify bool   `json:"notify"`
		Method string `json:"method"`
		UserID int    `json:"user_id"` // optional; defaults to current user
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: err.Error()})
		return
	}

	targetUserID := input.UserID
	if targetUserID == 0 {
		val := c.Get("claims")
		if val == nil {
			c.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "user_id required"})
			return
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

	created, err := h.uc.CreateNotification(r.Context(), n)
	if err != nil {
		c.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
		return
	}

	server.Created(c.GinContext(), created)
}

// markAsRead marks a specific notification as read.
func (h *NotificationHandler) markAsRead(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, server.Response[interface{}]{Code: server.ErrBadRequest, Message: "Invalid ID"})
		return
	}

	userID, _ := strconv.Atoi(claims.GetUserID())
	err = h.uc.MarkAsRead(r.Context(), id, userID)
	if err != nil {
		c.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
		return
	}

	c.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "marked as read"}})
}

// markAllRead marks all notifications as read for the current user.
func (h *NotificationHandler) markAllRead(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	userID, _ := strconv.Atoi(claims.GetUserID())
	err := h.uc.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		c.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
		return
	}

	c.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"message": "all marked as read"}})
}

// unreadCount returns the count of unread notifications for the current user.
func (h *NotificationHandler) unreadCount(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	val := c.Get("claims")
	if val == nil {
		c.JSON(401, server.Response[interface{}]{Code: server.ErrUnauthorized, Message: "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	userID, _ := strconv.Atoi(claims.GetUserID())
	count, err := h.uc.GetUnreadCount(r.Context(), userID)
	if err != nil {
		c.JSON(500, server.Response[interface{}]{Code: server.ErrInternal, Message: err.Error()})
		return
	}

	c.JSON(200, server.Response[interface{}]{Code: 0, Message: "ok", Data: gin.H{"unread_count": count}})
}
