package service

import (
	"regexp"
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/features/admin/dto"

	"origadmin/application/origstudio/internal/server"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// AdminTagHandler handles tag HTTP requests in admin panel
type AdminTagHandler struct {
	service *TagService
	jwtMgr  *auth.Manager
}

// NewAdminTagHandler creates a new AdminTagHandler
func NewAdminTagHandler(service *TagService, jwtMgr *auth.Manager) *AdminTagHandler {
	return &AdminTagHandler{service: service, jwtMgr: jwtMgr}
}

// RegisterRoutes registers tag routes
func (h *AdminTagHandler) RegisterRoutes(r http2.Router) {
	tags := r.Group("/admin/tags")
	{
		tags.GET("", server.WithAdminCtx(h.jwtMgr, h.listTags()))
		tags.GET("/:id", server.WithAdminCtx(h.jwtMgr, h.getTag()))
		tags.POST("", server.WithAdminCtx(h.jwtMgr, h.createTag()))
		tags.PUT("/:id", server.WithAdminCtx(h.jwtMgr, h.updateTag()))
		tags.DELETE("/:id", server.WithAdminCtx(h.jwtMgr, h.deleteTag()))
		tags.POST("/bulk", server.WithAdminCtx(h.jwtMgr, h.bulkTagOperation()))
		tags.GET("/export", server.WithAdminCtx(h.jwtMgr, h.exportTags()))
		tags.POST("/import", server.WithAdminCtx(h.jwtMgr, h.importTags()))
	}
}

// listTags handles GET /admin/tags
func (h *AdminTagHandler) listTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))

		search := ctx.QueryVar("search")
		if search == "" {
			search = ctx.QueryVar("keyword")
		}

		status := ctx.QueryVar("status")
		sortBy := ctx.QueryVarDefault("sort_by", "create_time")
		sortOrder := ctx.QueryVarDefault("sort_order", "desc")

		page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

		tags, total, err := h.service.List(ctx.Request().Context(), page, pageSize, search, status, sortBy, sortOrder)
		if err != nil {
			return server.FailCtx(ctx, 10000, "Failed to list tags")
		}

		tagResponses := ToTagResponseList(tags)

		totalPages := (int(total) + pageSize - 1) / pageSize

		return server.OKCtx(ctx, map[string]any{
			"items":       tagResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		})
	}
}

// getTag handles GET /admin/tags/:id
func (h *AdminTagHandler) getTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")

		tag, err := h.service.Get(ctx.Request().Context(), id)
		if err != nil {
			return server.FailCtx(ctx, 10001, "Tag not found")
		}

		return server.OKCtx(ctx, ToTagResponse(tag))
	}
}

// createTag handles POST /admin/tags
func (h *AdminTagHandler) createTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			Title       string `json:"title" binding:"required"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, 10004, "Invalid request")
		}

		if req.Color != "" && !hexColorRegex.MatchString(req.Color) {
			return server.FailCtx(ctx, 10004, "Invalid color format, expected #RRGGBB")
		}

		tag := &dto.TagDTO{
			Title:       req.Title,
			Description: req.Description,
			Color:       req.Color,
			Status:      ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			tag.Slug = req.Slug
		} else {
			tag.Slug = hashtag.GenerateTagSlug(req.Title)
		}

		createdTag, err := h.service.Create(ctx.Request().Context(), tag)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		return server.OKCtx(ctx, ToTagResponse(createdTag))
	}
}

// updateTag handles PUT /admin/tags/:id
func (h *AdminTagHandler) updateTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")

		var req struct {
			Title       string `json:"title"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, 10004, "Invalid request")
		}

		if req.Color != "" && !hexColorRegex.MatchString(req.Color) {
			return server.FailCtx(ctx, 10004, "Invalid color format, expected #RRGGBB")
		}

		updates := &dto.TagDTO{
			Title:       req.Title,
			Description: req.Description,
			Color:       req.Color,
			Status:      ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			updates.Slug = req.Slug
		}

		updatedTag, err := h.service.Update(ctx.Request().Context(), id, updates)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		return server.OKCtx(ctx, ToTagResponse(updatedTag))
	}
}

// deleteTag handles DELETE /admin/tags/:id
func (h *AdminTagHandler) deleteTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")

		if err := h.service.Delete(ctx.Request().Context(), id); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"code":    0,
			"message": "Tag deleted successfully",
		})
	}
}

// bulkTagOperation handles POST /admin/tags/bulk
func (h *AdminTagHandler) bulkTagOperation() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, 10004, "Invalid request")
		}

		if req.Action != "delete" {
			return server.FailCtx(ctx, 10004, "Unsupported action")
		}

		count, err := h.service.BulkDelete(ctx.Request().Context(), req.IDs)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"success": count,
			"failed":  len(req.IDs) - count,
		})
	}
}

// exportTags handles GET /admin/tags/export
func (h *AdminTagHandler) exportTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"message": "Export functionality not implemented yet",
		})
	}
}

// importTags handles POST /admin/tags/import
func (h *AdminTagHandler) importTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.OKCtx(ctx, map[string]any{
			"message": "Import functionality not implemented yet",
		})
	}
}
