package service

import (
	"strconv"


	http2 "origadmin/application/origstudio/internal/pkg/http"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

type Handler struct {
	permUC      *biz.UseCase
	jwtMgr      *auth.Manager
	permChecker authbiz.PermissionChecker
}

func NewHandler(permUC *biz.UseCase, jwtMgr *auth.Manager, permChecker authbiz.PermissionChecker) *Handler {
	return &Handler{permUC: permUC, jwtMgr: jwtMgr, permChecker: permChecker}
}

func (h *Handler) RegisterRoutes(r http2.Router) {
	adminPerms := r.Group("/admin/permission-groups")
	adminPerms.Use(server.JWTMiddlewareCtx(h.jwtMgr), server.RequirePermissionCtx(h.permChecker, "permission:read"))
	{
		adminPerms.GET("", h.listGroups())
		adminPerms.POST("", server.RequirePermissionCtx(h.permChecker, "permission:write")(h.createGroup()))
		adminPerms.GET("/:id", h.getGroup())
		adminPerms.PUT("/:id", server.RequirePermissionCtx(h.permChecker, "permission:write")(h.updateGroup()))
		adminPerms.DELETE("/:id", server.RequirePermissionCtx(h.permChecker, "permission:delete")(h.deleteGroup()))
		adminPerms.POST("/:id/toggle", server.RequirePermissionCtx(h.permChecker, "permission:write")(h.toggleGroup()))
		adminPerms.GET("/:id/members", h.listMembers())
		adminPerms.POST("/:id/members", server.RequirePermissionCtx(h.permChecker, "permission:manage")(h.addMembers()))
		adminPerms.DELETE("/:id/members/:user_id", server.RequirePermissionCtx(h.permChecker, "permission:manage")(h.removeMember()))
	}

	adminUsers := r.Group("/admin/users")
	adminUsers.Use(server.JWTMiddlewareCtx(h.jwtMgr), server.RequirePermissionCtx(h.permChecker, "permission:read"))
	{
		adminUsers.GET("/:id/permissions", h.getUserPermissions())
	}

	r.GET("/permissions", h.listPermissionEnums())
}

func (h *Handler) listGroups() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		var isActive *bool
		if v := ctx.QueryVar("is_active"); v != "" {
			parsed, err := strconv.ParseBool(v)
			if err == nil {
				isActive = &parsed
			}
		}
		items, total, err := h.permUC.ListGroup(ctx, isActive, page, pageSize)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total, "page": page, "page_size": pageSize})
		return nil
	}
}

func (h *Handler) createGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		if req.Name == "" {
			http2.Fail(ctx, server.ErrBadRequest, "name is required")
			return nil
		}
		for _, perm := range req.Permissions {
			if !biz.IsValidPermission(perm) {
				http2.Fail(ctx, server.ErrBadRequest, "invalid permission: "+perm)
				return nil
			}
		}
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}
		adminID := claims.GetUserID()
		group, err := h.permUC.CreateGroup(ctx, req.Name, req.Description, req.Permissions, req.CategoryScope, adminID)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.Created(ctx, group)
		return nil
	}
}

func (h *Handler) getGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		group, err := h.permUC.GetGroup(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "permission group not found")
			return nil
		}
		http2.OK(ctx, group)
		return nil
	}
}

func (h *Handler) updateGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		for _, perm := range req.Permissions {
			if !biz.IsValidPermission(perm) {
				http2.Fail(ctx, server.ErrBadRequest, "invalid permission: "+perm)
				return nil
			}
		}
		group, err := h.permUC.UpdateGroup(ctx, id, req.Name, req.Description, req.Permissions, req.CategoryScope)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, group)
		return nil
	}
}

func (h *Handler) deleteGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		err := h.permUC.DeleteGroup(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"message": "permission group deleted"})
		return nil
	}
}

func (h *Handler) toggleGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		var req struct {
			IsActive bool `json:"is_active"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		err := h.permUC.ToggleGroup(ctx, id, req.IsActive)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"id": id, "is_active": req.IsActive})
		return nil
	}
}

func (h *Handler) listMembers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "50"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}
		items, total, err := h.permUC.ListMembers(ctx, id, page, pageSize)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total, "page": page, "page_size": pageSize})
		return nil
	}
}

func (h *Handler) addMembers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		var req struct {
			UserIDs []string `json:"user_ids"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		if len(req.UserIDs) == 0 {
			http2.Fail(ctx, server.ErrBadRequest, "user_ids is required")
			return nil
		}
		if len(req.UserIDs) > 100 {
			http2.Fail(ctx, server.ErrBadRequest, "user_ids cannot exceed 100")
			return nil
		}
		added, skipped, err := h.permUC.AddMembers(ctx, id, req.UserIDs)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"added": added, "skipped": skipped})
		return nil
	}
}

func (h *Handler) removeMember() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		userID := ctx.Var("user_id")
		if userID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "user_id is required")
			return nil
		}
		err := h.permUC.RemoveMember(ctx, id, userID)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"message": "member removed"})
		return nil
	}
}

func (h *Handler) getUserPermissions() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		detail, err := h.permUC.GetUserPermissions(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, detail)
		return nil
	}
}

func (h *Handler) listPermissionEnums() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, map[string]interface{}{
			"permissions":   biz.AllPermissions,
			"role_defaults": biz.RoleDefaultPermissions,
		})
		return nil
	}
}
