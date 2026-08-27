package service

import (
	"strconv"

	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"
)

type NotificationHandler struct {
	uc  *biz.NotificationUseCase
	jwt *auth.Manager
}

func NewNotificationHandler(uc *biz.NotificationUseCase, jwt *auth.Manager) *NotificationHandler {
	return &NotificationHandler{uc: uc, jwt: jwt}
}

func (h *NotificationHandler) RegisterRoutes(r http2.Router) {
	userNotifs := r.Group("/notifications")
	userNotifs.Use(server.JWTMiddlewareCtx(h.jwt))
	{
		userNotifs.GET("", h.listNotifications)
		userNotifs.POST("/read-all", h.markAllRead)
		userNotifs.DELETE("/read", h.deleteReadNotifications)
		userNotifs.DELETE("", h.deleteAllNotifications)
		userNotifs.GET("/unread-count", h.unreadCount)
		userNotifs.DELETE("/:id", h.deleteNotification)
		userNotifs.POST("/:id/read", h.markAsRead)
	}

	// Permission groups: content service owns the group data; the gateway proxies
	// this path to content (the user-service PermissionService is a stub).
	pgGroup := r.Group("/admin/permission-groups")
	pgGroup.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		pgGroup.GET("", h.listPermissionGroups)
	}

	// Admin notification management — BUG-262: restored after BUG-123 removed the
	// legacy admin* methods (design anchor: docs/modules/content/notification/
	// 00-INDEX.md §3.2). Gateway proxies /api/v1/admin/notifications to content.
	adminNotifs := r.Group("/admin/notifications")
	adminNotifs.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminNotifs.GET("", h.adminListNotifications)
		adminNotifs.POST("", h.adminCreateNotification)
		adminNotifs.POST("/broadcast", h.adminBroadcastNotification)
		adminNotifs.POST("/test", h.adminSendTestNotification)
		adminNotifs.DELETE("/:id", h.adminDeleteNotification)
		adminNotifs.GET("/types", h.adminListNotificationTypes)
	}
}

func getUserIDFromGin(ctx http2.Context) (string, bool) {
	val, exists := ctx.Get("claims")
	if !exists || val == nil {
		return "", false
	}
	claims := val.(*auth.Claims)
	return claims.GetUserID(), true
}

func (h *NotificationHandler) listNotifications(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	limit, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	page, limit = types.NormalizeHTTPPagination(page, limit)

	items, total, err := h.uc.ListUserNotifications(ctx, userID, page, limit)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to fetch notifications: "+err.Error())
	}

	unread, _ := h.uc.GetUnreadCount(ctx, userID)

	return server.OKCtx(ctx, map[string]any{
		"items":        items,
		"total":        total,
		"unread_count": unread,
		"page":         page,
		"page_size":    limit,
	})
}

func (h *NotificationHandler) listPermissionGroups(ctx http2.Context) error {
	items, err := h.uc.ListPermissionGroups(ctx.Request().Context())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "Failed to list permission groups: "+err.Error())
	}
	server.PageCtx(ctx, items, int64(len(items)), 1, len(items))
	return nil
}

func (h *NotificationHandler) markAsRead(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	id, err := strconv.Atoi(ctx.Var("id"))
	if err != nil {
		return server.FailCtx(ctx, 400, "Invalid ID")
	}

	err = h.uc.MarkAsRead(ctx, id, userID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to mark as read: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "marked as read"})
}

func (h *NotificationHandler) markAllRead(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	err := h.uc.MarkAllAsRead(ctx, userID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to mark all as read: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "all marked as read"})
}

func (h *NotificationHandler) unreadCount(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	count, err := h.uc.GetUnreadCount(ctx, userID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to get unread count: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"unread_count": count})
}

func (h *NotificationHandler) deleteNotification(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	id, err := strconv.Atoi(ctx.Var("id"))
	if err != nil {
		return server.FailCtx(ctx, 400, "Invalid ID")
	}

	err = h.uc.DeleteNotification(ctx, id, userID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to delete notification: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "deleted"})
}

func (h *NotificationHandler) deleteAllNotifications(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	count, err := h.uc.DeleteAllNotifications(ctx, userID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to delete all notifications: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "all deleted", "deleted_count": count})
}

func (h *NotificationHandler) deleteReadNotifications(ctx http2.Context) error {
	userID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	count, err := h.uc.DeleteReadNotifications(ctx, userID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to delete read notifications: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "read notifications deleted", "deleted_count": count})
}

// --- Admin notification management (BUG-262 restored) ---

func (h *NotificationHandler) adminListNotifications(ctx http2.Context) error {
	limit, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	page, limit = types.NormalizeHTTPPagination(page, limit)

	items, total, err := h.uc.ListAllNotifications(ctx, page, limit)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to fetch notifications: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

func (h *NotificationHandler) adminCreateNotification(ctx http2.Context) error {
	var input struct {
		Action      string   `json:"action" binding:"required,max=30"`
		Notify      bool     `json:"notify"`
		Method      string   `json:"method"`
		UserIDs     []string `json:"user_ids"`
		RoleList    []string `json:"role_list"`
		GroupIDList []string `json:"group_id_list"`
		Title       string   `json:"title" binding:"required,max=200"`
		Body        string   `json:"body" binding:"required"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, 400, "Invalid request body: "+err.Error())
	}

	n := &biz.Notification{
		Action: input.Action,
		Notify: input.Notify,
		Method: input.Method,
		Title:  input.Title,
		Body:   input.Body,
	}

	// Resolve target users with priority: explicit user_ids > roles > groups > active users
	userIDs, err := h.uc.ResolveTargetUsers(ctx, input.UserIDs, input.RoleList, input.GroupIDList)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to resolve target users: "+err.Error())
	}
	if len(userIDs) == 0 {
		return server.FailCtx(ctx, 400, "no target users matched the provided filters")
	}

	created, err := h.uc.BatchCreateNotifications(ctx, userIDs, n)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to send notifications: "+err.Error())
	}

	return server.CreatedCtx(ctx, map[string]any{"items": created, "sent_count": len(created)})
}

func (h *NotificationHandler) adminBroadcastNotification(ctx http2.Context) error {
	var input struct {
		Action      string   `json:"action" binding:"required,max=30"`
		Notify      bool     `json:"notify"`
		Method      string   `json:"method"`
		Title       string   `json:"title" binding:"required,max=200"`
		Body        string   `json:"body" binding:"required"`
		RoleList    []string `json:"role_list"`
		GroupIDList []string `json:"group_id_list"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, 400, "Invalid request body: "+err.Error())
	}

	n := &biz.Notification{
		Action: input.Action,
		Notify: input.Notify,
		Method: input.Method,
		Title:  input.Title,
		Body:   input.Body,
	}

	// Role/group filters take precedence; otherwise broadcast to all active users.
	var created []*biz.Notification
	var err error

	if len(input.RoleList) > 0 {
		created, err = h.uc.BroadcastByRole(ctx, input.RoleList, n)
	} else if len(input.GroupIDList) > 0 {
		created, err = h.uc.BroadcastByGroup(ctx, input.GroupIDList, n)
	} else {
		created, err = h.uc.BroadcastToAll(ctx, n)
	}
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to broadcast notification: "+err.Error())
	}

	return server.CreatedCtx(ctx, map[string]any{"items": created, "sent_count": len(created)})
}

func (h *NotificationHandler) adminSendTestNotification(ctx http2.Context) error {
	adminID, ok := getUserIDFromGin(ctx)
	if !ok {
		return server.FailCtx(ctx, 401, "unauthorized")
	}

	created, err := h.uc.SendTestNotification(ctx, adminID)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to send test notification: "+err.Error())
	}

	return server.CreatedCtx(ctx, created)
}

func (h *NotificationHandler) adminDeleteNotification(ctx http2.Context) error {
	id, err := strconv.Atoi(ctx.Var("id"))
	if err != nil {
		return server.FailCtx(ctx, 400, "Invalid ID")
	}

	err = h.uc.AdminDeleteNotification(ctx, id)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to delete notification: "+err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "deleted"})
}

// adminListNotificationTypes 返回受控通知类型清单（供管理端发送下拉 + 类型开关设置页）。
func (h *NotificationHandler) adminListNotificationTypes(ctx http2.Context) error {
	return server.OKCtx(ctx, map[string]any{"items": biz.NotificationTypes()})
}
