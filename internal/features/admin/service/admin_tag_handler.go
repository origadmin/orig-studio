package service

import (
	"net/http"
	"regexp"
	"strconv"

	"origadmin/application/origcms/internal/data/entity"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/helpers/hashtag"
	"origadmin/application/origcms/internal/helpers/repo"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/server"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// AdminTagHandler handles tag HTTP requests in admin panel
type AdminTagHandler struct {
	service *TagService
}

// NewAdminTagHandler creates a new AdminTagHandler
func NewAdminTagHandler(service *TagService) *AdminTagHandler {
	return &AdminTagHandler{service: service}
}

// RegisterRoutes registers tag routes
func (h *AdminTagHandler) RegisterRoutes(r http2.Router) {
	tags := r.Group("/admin/tags")
	{
		tags.GET("", server.HTTPToHandlerFunc(h.listTags()))
		tags.GET("/:id", server.HTTPToHandlerFunc(h.getTag()))
		tags.POST("", server.HTTPToHandlerFunc(h.createTag()))
		tags.PUT("/:id", server.HTTPToHandlerFunc(h.updateTag()))
		tags.DELETE("/:id", server.HTTPToHandlerFunc(h.deleteTag()))
		tags.POST("/bulk", server.HTTPToHandlerFunc(h.bulkTagOperation()))
		tags.GET("/export", server.HTTPToHandlerFunc(h.exportTags()))
		tags.POST("/import", server.HTTPToHandlerFunc(h.importTags()))
	}
}

// listTags handles GET /admin/tags
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
// Also supports both "search" and "keyword" query parameters.
func (h *AdminTagHandler) listTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// Parse query parameters
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))

		// B087-R2 Fix: Support both "search" and "keyword" parameters.
		// Frontend sends "keyword", backend originally expected "search".
		search := gc.Query("search")
		if search == "" {
			search = gc.Query("keyword")
		}

		status := gc.Query("status")
		sortBy := gc.DefaultQuery("sort_by", "create_time")
		sortOrder := gc.DefaultQuery("sort_order", "desc")

		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		// Get tags
		tags, total, err := h.service.List(r.Context(), page, pageSize, search, status, sortBy, sortOrder)
		if err != nil {
			server.Fail(gc, 10000, "Failed to list tags")
			return
		}

		// B087-R2 Fix: Convert entity.Tag to TagResponse DTO
		tagResponses := ToTagResponseList(tags)

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		// Return response with frontend-compatible field names
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
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
func (h *AdminTagHandler) getTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")

		tag, err := h.service.Get(r.Context(), id)
		if err != nil {
			server.Fail(gc, 10001, "Tag not found")
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(gc, ToTagResponse(tag))
	}
}

// createTag handles POST /admin/tags
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
// Also maps frontend "status" (lowercase) to DB enum.
func (h *AdminTagHandler) createTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug"` // Optional: auto-generated from name when empty
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

		tag := &entity.Tag{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			// B087-R2 Fix: Parse frontend status string to DB enum
			Status: ParseTagStatus(req.Status),
		}

		// Auto-generate slug from name when not provided
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

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(gc, ToTagResponse(createdTag))
	}
}

// updateTag handles PUT /admin/tags/:id
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
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

		updates := &entity.Tag{
			Title:       req.Name,
			Description: req.Description,
			Color:       req.Color,
			// B087-R2 Fix: Parse frontend status string to DB enum
			Status: ParseTagStatus(req.Status),
		}

		if req.Slug != "" {
			updates.Slug = req.Slug
		}

		updatedTag, err := h.service.Update(r.Context(), id, updates)
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
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
		// TODO: Implement export functionality
		server.OK(gc, gin.H{
			"message": "Export functionality not implemented yet",
		})
	}
}

// importTags handles POST /admin/tags/import
func (h *AdminTagHandler) importTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement import functionality
		server.OK(gc, gin.H{
			"message": "Import functionality not implemented yet",
		})
	}
}
