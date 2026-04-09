package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
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

func (h *NotificationHandler) Register(group *gin.RouterGroup) {
	notifs := group.Group("/notifications")
	{
		// Protected routes — all notification operations require auth
		notifs.Use(JWTMiddleware(h.jwt))

		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		notifs.GET("", h.listNotifications)
		notifs.POST("", h.createNotification)
		notifs.POST("/read-all", h.markAllRead)
		notifs.GET("/unread-count", h.unreadCount)

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		notifs.POST("/:id/read", h.markAsRead)
	}
}

// listNotifications returns notifications for the authenticated user,
// ordered by most recent, with pagination support.
func (h *NotificationHandler) listNotifications(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListUserNotifications(
		c.Request.Context(),
		int(claims.UserID),
		page,
		limit,
	)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	unread, _ := h.uc.GetUnreadCount(c.Request.Context(), int(claims.UserID))

	OK(c, gin.H{
		"items":        items,
		"total":        total,
		"unread_count": unread,
		"page":         page,
		"page_size":    limit,
	})
}

// createNotification creates a new notification.
// POST body: {"action": string, "notify": bool, "method": string, "user_id": int}
func (h *NotificationHandler) createNotification(c *gin.Context) {
	var input struct {
		Action string `json:"action" binding:"required,max=30"`
		Notify bool   `json:"notify"`
		Method string `json:"method"`
		UserID int    `json:"user_id"` // optional; defaults to current user
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, ErrBadRequest, err.Error())
		return
	}

	targetUserID := input.UserID
	if targetUserID == 0 {
		val, ok := c.Get("claims")
		if !ok || val == nil {
			Fail(c, ErrBadRequest, "user_id required")
			return
		}
		claims := val.(*auth.Claims)
		targetUserID = int(claims.UserID)
	}

	n := &biz.Notification{
		Action: input.Action,
		Notify: input.Notify,
		Method: input.Method,
		UserID: targetUserID,
	}

	created, err := h.uc.CreateNotification(c.Request.Context(), n)
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: created})
}

// markAsRead marks a specific notification as read.
func (h *NotificationHandler) markAsRead(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, ErrBadRequest, "Invalid ID")
		return
	}

	err = h.uc.MarkAsRead(c.Request.Context(), id, int(claims.UserID))
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{"message": "marked as read"})
}

// markAllRead marks all notifications as read for the current user.
func (h *NotificationHandler) markAllRead(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.MarkAllAsRead(c.Request.Context(), int(claims.UserID))
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{"message": "all marked as read"})
}

// unreadCount returns the count of unread notifications for the current user.
func (h *NotificationHandler) unreadCount(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		Fail(c, ErrUnauthorized, "unauthorized")
		return
	}
	claims := val.(*auth.Claims)

	count, err := h.uc.GetUnreadCount(c.Request.Context(), int(claims.UserID))
	if err != nil {
		Fail(c, ErrInternal, err.Error())
		return
	}

	OK(c, gin.H{"unread_count": count})
}
