package server

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	authbiz "origadmin/application/origcms/internal/svc-auth/biz"
)

type PermissionHandler struct {
	permUC *authbiz.PermissionUseCase
	jwtMgr *auth.Manager
}

func NewPermissionHandler(permUC *authbiz.PermissionUseCase, jwtMgr *auth.Manager) *PermissionHandler {
	return &PermissionHandler{permUC: permUC, jwtMgr: jwtMgr}
}

func (h *PermissionHandler) RegisterRoutes(apiV1 *gin.RouterGroup) {
	adminPerms := apiV1.Group("/admin/permission-groups")
	adminPerms.Use(AdminMiddleware(h.jwtMgr))
	{
		adminPerms.GET("", h.listGroups())
		adminPerms.POST("", h.createGroup())
		adminPerms.GET("/:id", h.getGroup())
		adminPerms.PUT("/:id", h.updateGroup())
		adminPerms.DELETE("/:id", h.deleteGroup())
		adminPerms.POST("/:id/toggle", h.toggleGroup())
		adminPerms.GET("/:id/members", h.listMembers())
		adminPerms.POST("/:id/members", h.addMembers())
		adminPerms.DELETE("/:id/members/:user_id", h.removeMember())
	}

	adminUsers := apiV1.Group("/admin/users")
	adminUsers.Use(AdminMiddleware(h.jwtMgr))
	{
		adminUsers.GET("/:id/permissions", h.getUserPermissions())
	}

	apiV1.GET("/permissions", h.listPermissionEnums())
}

func (h *PermissionHandler) listGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		var isActive *bool
		if v := c.Query("is_active"); v != "" {
			parsed, err := strconv.ParseBool(v)
			if err == nil {
				isActive = &parsed
			}
		}

		items, total, err := h.permUC.ListGroup(ctx, isActive, page, pageSize)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PermissionHandler) createGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if req.Name == "" {
			Fail(c, ErrBadRequest, "name is required")
			return
		}

		for _, perm := range req.Permissions {
			if !authbiz.IsValidPermission(perm) {
				Fail(c, ErrBadRequest, "invalid permission: "+perm)
				return
			}
		}

		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		group, err := h.permUC.CreateGroup(ctx, req.Name, req.Description, req.Permissions, req.CategoryScope, adminID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		c.JSON(201, Response[interface{}]{Code: 0, Message: "ok", Data: group})
	}
}

func (h *PermissionHandler) getGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		group, err := h.permUC.GetGroup(ctx, id)
		if err != nil {
			Fail(c, ErrNotFound, "permission group not found")
			return
		}

		OK(c, group)
	}
}

func (h *PermissionHandler) updateGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		for _, perm := range req.Permissions {
			if !authbiz.IsValidPermission(perm) {
				Fail(c, ErrBadRequest, "invalid permission: "+perm)
				return
			}
		}

		group, err := h.permUC.UpdateGroup(ctx, id, req.Name, req.Description, req.Permissions, req.CategoryScope)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, group)
	}
}

func (h *PermissionHandler) deleteGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		err := h.permUC.DeleteGroup(ctx, id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{"message": "permission group deleted"})
	}
}

func (h *PermissionHandler) toggleGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		var req struct {
			IsActive bool `json:"is_active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		err := h.permUC.ToggleGroup(ctx, id, req.IsActive)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"id":        id,
			"is_active": req.IsActive,
		})
	}
}

func (h *PermissionHandler) listMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}

		items, total, err := h.permUC.ListMembers(ctx, id, page, pageSize)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PermissionHandler) addMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		var req struct {
			UserIDs []string `json:"user_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, ErrBadRequest, err.Error())
			return
		}

		if len(req.UserIDs) == 0 {
			Fail(c, ErrBadRequest, "user_ids is required")
			return
		}
		if len(req.UserIDs) > 100 {
			Fail(c, ErrBadRequest, "user_ids cannot exceed 100")
			return
		}

		added, skipped, err := h.permUC.AddMembers(ctx, id, req.UserIDs)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{
			"added":   added,
			"skipped": skipped,
		})
	}
}

func (h *PermissionHandler) removeMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		userID := c.Param("user_id")
		if userID == "" {
			Fail(c, ErrBadRequest, "user_id is required")
			return
		}

		err := h.permUC.RemoveMember(ctx, id, userID)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, gin.H{"message": "member removed"})
	}
}

func (h *PermissionHandler) getUserPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id := c.Param("id")
		if id == "" {
			Fail(c, ErrBadRequest, "id is required")
			return
		}

		detail, err := h.permUC.GetUserPermissions(ctx, id)
		if err != nil {
			Fail(c, ErrInternal, err.Error())
			return
		}

		OK(c, detail)
	}
}

func (h *PermissionHandler) listPermissionEnums() gin.HandlerFunc {
	return func(c *gin.Context) {
		OK(c, gin.H{
			"permissions":    authbiz.AllPermissions,
			"role_defaults":  authbiz.RoleDefaultPermissions,
		})
	}
}
