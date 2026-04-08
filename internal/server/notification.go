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

		notifs.GET("", h.listNotifications)
		notifs.POST("/:id/read", h.markAsRead)
		notifs.POST("/read-all", h.markAllRead)
		notifs.GET("/unread-count", h.unreadCount)

		// Admin: create notification (internal use or admin)
		notifs.POST("", h.createNotification)
	}
}

// listNotifications returns notifications for the authenticated user,
// ordered by most recent, with pagination support.
func (h *NotificationHandler) listNotifications(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListUserNotifications(c.Request.Context(), int(claims.UserID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	unread, _ := h.uc.GetUnreadCount(c.Request.Context(), int(claims.UserID))

	c.JSON(http.StatusOK, gin.H{
		"list":   items,
		"total":  total,
		"unread": unread,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID := input.UserID
	if targetUserID == 0 {
		val, ok := c.Get("claims")
		if !ok || val == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// markAsRead marks a specific notification as read.
func (h *NotificationHandler) markAsRead(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.uc.MarkAsRead(c.Request.Context(), id, int(claims.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

// markAllRead marks all notifications as read for the current user.
func (h *NotificationHandler) markAllRead(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	err := h.uc.MarkAllAsRead(c.Request.Context(), int(claims.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all marked as read"})
}

// unreadCount returns the count of unread notifications for the current user.
func (h *NotificationHandler) unreadCount(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	count, err := h.uc.GetUnreadCount(c.Request.Context(), int(claims.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
