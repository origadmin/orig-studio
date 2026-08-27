/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"net/http"
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	"origadmin/application/origstudio/internal/server"
)

// PermissionHandler handles permission-related HTTP endpoints.
type PermissionHandler struct {
	permUC *authbiz.PermissionUseCase
	jwtMgr *auth.Manager
}

// NewPermissionHandler creates a new PermissionHandler.
func NewPermissionHandler(permUC *authbiz.PermissionUseCase, jwtMgr *auth.Manager) *PermissionHandler {
	return &PermissionHandler{permUC: permUC, jwtMgr: jwtMgr}
}

// RegisterRoutes registers the handler's routes.
func (h *PermissionHandler) RegisterRoutes(r http2.Router) {
	adminPerms := r.Group("/admin/permission-groups")
	adminPerms.Use(server.AdminMiddlewareCtx(h.jwtMgr))
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

	adminUsers := r.Group("/admin/users")
	adminUsers.Use(server.AdminMiddlewareCtx(h.jwtMgr))
	{
		adminUsers.GET("/:id/permissions", h.getUserPermissions())
	}

	// Public endpoint
	r.GET("/permissions", h.listPermissionEnums())
}

func (h *PermissionHandler) listGroups() http2.HandlerFunc {
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
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PermissionHandler) createGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if req.Name == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "name is required")
		}

		for _, perm := range req.Permissions {
			if !authbiz.IsValidPermission(perm) {
				return server.FailCtx(ctx, server.ErrBadRequest, "invalid permission: "+perm)
			}
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		adminID := claims.GetUserID()

		group, err := h.permUC.CreateGroup(ctx, req.Name, req.Description, req.Permissions, req.CategoryScope, adminID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return ctx.JSON(http.StatusCreated, group)
	}
}

func (h *PermissionHandler) getGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		group, err := h.permUC.GetGroup(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "permission group not found")
		}

		return server.OKCtx(ctx, group)
	}
}

func (h *PermissionHandler) updateGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		for _, perm := range req.Permissions {
			if !authbiz.IsValidPermission(perm) {
				return server.FailCtx(ctx, server.ErrBadRequest, "invalid permission: "+perm)
			}
		}

		group, err := h.permUC.UpdateGroup(ctx, id, req.Name, req.Description, req.Permissions, req.CategoryScope)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, group)
	}
}

func (h *PermissionHandler) deleteGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		err := h.permUC.DeleteGroup(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"message": "permission group deleted"})
	}
}

func (h *PermissionHandler) toggleGroup() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var req struct {
			IsActive bool `json:"is_active"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		err := h.permUC.ToggleGroup(ctx, id, req.IsActive)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"id":        id,
			"is_active": req.IsActive,
		})
	}
}

func (h *PermissionHandler) listMembers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
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
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PermissionHandler) addMembers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var req struct {
			UserIDs []string `json:"user_ids"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if len(req.UserIDs) == 0 {
			return server.FailCtx(ctx, server.ErrBadRequest, "user_ids is required")
		}
		if len(req.UserIDs) > 100 {
			return server.FailCtx(ctx, server.ErrBadRequest, "user_ids cannot exceed 100")
		}

		added, skipped, err := h.permUC.AddMembers(ctx, id, req.UserIDs)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"added":   added,
			"skipped": skipped,
		})
	}
}

func (h *PermissionHandler) removeMember() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		userID := ctx.Var("user_id")
		if userID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "user_id is required")
		}

		err := h.permUC.RemoveMember(ctx, id, userID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{"message": "member removed"})
	}
}

func (h *PermissionHandler) getUserPermissions() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		detail, err := h.permUC.GetUserPermissions(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, detail)
	}
}

func (h *PermissionHandler) listPermissionEnums() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"permissions":   authbiz.AllPermissions,
			"role_defaults": authbiz.RoleDefaultPermissions,
		})
	}
}
