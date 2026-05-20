package service

import (
	"net/http"
	"regexp"
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/features/admin/dto"

	"github.com/gin-gonic/gin"
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
		tags.GET("", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.listTags())))
		tags.GET("/:id", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.getTag())))
		tags.POST("", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.createTag())))
		tags.PUT("/:id", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.updateTag())))
		tags.DELETE("/:id", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.deleteTag())))
		tags.POST("/bulk", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.bulkTagOperation())))
		tags.GET("/export", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.exportTags())))
		tags.POST("/import", server.WithAdminCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.importTags())))
	}
}

// listTags handles GET /admin/tags
func (h *AdminTagHandler) listTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))

		search := gc.Query("search")
		if search == "" {
			search = gc.Query("keyword")
		}

		status := gc.Query("status")
		sortBy := gc.DefaultQuery("sort_by", "create_time")
		sortOrder := gc.DefaultQuery("sort_order", "desc")

		page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

		tags, total, err := h.service.List(r.Context(), page, pageSize, search, status, sortBy, sortOrder)
		if err != nil {
			server.Fail(gc, 10000, "Failed to list tags")
			return
		}

		tagResponses := ToTagResponseList(tags)

		totalPages := (int(total) + pageSize - 1) / pageSize

		server.OK(gc, gin.H{
			"items":       tagResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		})
	}
}

// getTag handles GET /admin/tags/:id
func (h *AdminTagHandler) getTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		tag, err := h.service.Get(r.Context(), id)
		if err != nil {
			server.Fail(gc, 10001, "Tag not found")
			return
		}

		server.OK(gc, ToTagResponse(tag))
	}
}

// createTag handles POST /admin/tags
func (h *AdminTagHandler) createTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, 10004, "Invalid request")
			return
		}

		if req.Color != "" && !hexColorRegex.MatchString(req.Color) {
			server.Fail(gc, 10004, "Invalid color format, expected #RRGGBB")
			return
		}

		tag := &dto.TagDTO{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			Status:      ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			tag.Slug = req.Slug
		} else {
			tag.Slug = hashtag.GenerateTagSlug(req.Name)
		}

		createdTag, err := h.service.Create(r.Context(), tag)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(gc, ToTagResponse(createdTag))
	}
}

// updateTag handles PUT /admin/tags/:id
func (h *AdminTagHandler) updateTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		var req struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, 10004, "Invalid request")
			return
		}

		if req.Color != "" && !hexColorRegex.MatchString(req.Color) {
			server.Fail(gc, 10004, "Invalid color format, expected #RRGGBB")
			return
		}

		updates := &dto.TagDTO{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			Status:      ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			updates.Slug = req.Slug
		}

		updatedTag, err := h.service.Update(r.Context(), id, updates)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(gc, ToTagResponse(updatedTag))
	}
}

// deleteTag handles DELETE /admin/tags/:id
func (h *AdminTagHandler) deleteTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		if err := h.service.Delete(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"code":    0,
			"message": "Tag deleted successfully",
		})
	}
}

// bulkTagOperation handles POST /admin/tags/bulk
func (h *AdminTagHandler) bulkTagOperation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, 10004, "Invalid request")
			return
		}

		if req.Action != "delete" {
			server.Fail(gc, 10004, "Unsupported action")
			return
		}

		count, err := h.service.BulkDelete(r.Context(), req.IDs)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"success": count,
			"failed":  len(req.IDs) - count,
		})
	}
}

// exportTags handles GET /admin/tags/export
func (h *AdminTagHandler) exportTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"message": "Export functionality not implemented yet",
		})
	}
}

// importTags handles POST /admin/tags/import
func (h *AdminTagHandler) importTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"message": "Import functionality not implemented yet",
		})
	}
}
