/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
	"origadmin/application/origcms/internal/server"
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
		adminPerms.GET("", server.HTTPToHandlerFunc(h.listGroups()))
		adminPerms.POST("", server.HTTPToHandlerFunc(h.createGroup()))
		adminPerms.GET("/:id", server.HTTPToHandlerFunc(h.getGroup()))
		adminPerms.PUT("/:id", server.HTTPToHandlerFunc(h.updateGroup()))
		adminPerms.DELETE("/:id", server.HTTPToHandlerFunc(h.deleteGroup()))
		adminPerms.POST("/:id/toggle", server.HTTPToHandlerFunc(h.toggleGroup()))
		adminPerms.GET("/:id/members", server.HTTPToHandlerFunc(h.listMembers()))
		adminPerms.POST("/:id/members", server.HTTPToHandlerFunc(h.addMembers()))
		adminPerms.DELETE("/:id/members/:user_id", server.HTTPToHandlerFunc(h.removeMember()))
	}

	adminUsers := r.Group("/admin/users")
	adminUsers.Use(server.AdminMiddlewareCtx(h.jwtMgr))
	{
		adminUsers.GET("/:id/permissions", server.HTTPToHandlerFunc(h.getUserPermissions()))
	}

	// Public endpoint
	r.GET("/permissions", server.HTTPToHandlerFunc(h.listPermissionEnums()))
}

func (h *PermissionHandler) listGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		var isActive *bool
		if v := gc.Query("is_active"); v != "" {
			parsed, err := strconv.ParseBool(v)
			if err == nil {
				isActive = &parsed
			}
		}

		items, total, err := h.permUC.ListGroup(ctx, isActive, page, pageSize)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PermissionHandler) createGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if req.Name == "" {
			server.Fail(gc, server.ErrBadRequest, "name is required")
			return
		}

		for _, perm := range req.Permissions {
			if !authbiz.IsValidPermission(perm) {
				server.Fail(gc, server.ErrBadRequest, "invalid permission: "+perm)
				return
			}
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		adminID := claims.GetUserID()

		group, err := h.permUC.CreateGroup(ctx, req.Name, req.Description, req.Permissions, req.CategoryScope, adminID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		gc.JSON(201, server.Response[interface{}]{Code: 0, Message: "ok", Data: group})
	}
}

func (h *PermissionHandler) getGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		group, err := h.permUC.GetGroup(ctx, id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "permission group not found")
			return
		}

		server.OK(gc, group)
	}
}

func (h *PermissionHandler) updateGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		var req struct {
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Permissions   []string `json:"permissions"`
			CategoryScope []string `json:"category_scope"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		for _, perm := range req.Permissions {
			if !authbiz.IsValidPermission(perm) {
				server.Fail(gc, server.ErrBadRequest, "invalid permission: "+perm)
				return
			}
		}

		group, err := h.permUC.UpdateGroup(ctx, id, req.Name, req.Description, req.Permissions, req.CategoryScope)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, group)
	}
}

func (h *PermissionHandler) deleteGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		err := h.permUC.DeleteGroup(ctx, id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"message": "permission group deleted"})
	}
}

func (h *PermissionHandler) toggleGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		var req struct {
			IsActive bool `json:"is_active"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		err := h.permUC.ToggleGroup(ctx, id, req.IsActive)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"id":        id,
			"is_active": req.IsActive,
		})
	}
}

func (h *PermissionHandler) listMembers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "50"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}

		items, total, err := h.permUC.ListMembers(ctx, id, page, pageSize)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PermissionHandler) addMembers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		var req struct {
			UserIDs []string `json:"user_ids"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if len(req.UserIDs) == 0 {
			server.Fail(gc, server.ErrBadRequest, "user_ids is required")
			return
		}
		if len(req.UserIDs) > 100 {
			server.Fail(gc, server.ErrBadRequest, "user_ids cannot exceed 100")
			return
		}

		added, skipped, err := h.permUC.AddMembers(ctx, id, req.UserIDs)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"added":   added,
			"skipped": skipped,
		})
	}
}

func (h *PermissionHandler) removeMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		userID := gc.Param("user_id")
		if userID == "" {
			server.Fail(gc, server.ErrBadRequest, "user_id is required")
			return
		}

		err := h.permUC.RemoveMember(ctx, id, userID)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{"message": "member removed"})
	}
}

func (h *PermissionHandler) getUserPermissions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		detail, err := h.permUC.GetUserPermissions(ctx, id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, detail)
	}
}

func (h *PermissionHandler) listPermissionEnums() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"permissions":   authbiz.AllPermissions,
			"role_defaults": authbiz.RoleDefaultPermissions,
		})
	}
}
