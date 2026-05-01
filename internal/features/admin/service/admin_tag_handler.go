package service

import (
	"strconv"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/helpers/hashtag"
	"origadmin/application/origcms/internal/helpers/repo"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/server"
)

// AdminTagHandler handles tag HTTP requests in admin panel
type AdminTagHandler struct {
	service *TagService
}

// NewAdminTagHandler creates a new AdminTagHandler
func NewAdminTagHandler(service *TagService) *AdminTagHandler {
	return &AdminTagHandler{service: service}
}

// RegisterRoutes registers tag routes
func (h *AdminTagHandler) RegisterRoutes(r *gin.RouterGroup) {
	tags := r.Group("/admin/tags")
	{
		tags.GET("", h.listTags())
		tags.GET("/:id", h.getTag())
		tags.POST("", h.createTag())
		tags.PUT("/:id", h.updateTag())
		tags.DELETE("/:id", h.deleteTag())
		tags.POST("/bulk", h.bulkTagOperation())
		tags.GET("/export", h.exportTags())
		tags.POST("/import", h.importTags())
	}
}

// listTags handles GET /admin/tags
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
// Also supports both "search" and "keyword" query parameters.
func (h *AdminTagHandler) listTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		// B087-R2 Fix: Support both "search" and "keyword" parameters.
		// Frontend sends "keyword", backend originally expected "search".
		search := c.Query("search")
		if search == "" {
			search = c.Query("keyword")
		}

		status := c.Query("status")
		sortBy := c.DefaultQuery("sort_by", "create_time")
		sortOrder := c.DefaultQuery("sort_order", "desc")

		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		// Get tags
		tags, total, err := h.service.List(c.Request.Context(), page, pageSize, search, status, sortBy, sortOrder)
		if err != nil {
			server.Fail(c, 10000, "Failed to list tags")
			return
		}

		// B087-R2 Fix: Convert entity.Tag to TagResponse DTO
		tagResponses := ToTagResponseList(tags)

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		// Return response with frontend-compatible field names
		server.OK(c, gin.H{
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
func (h *AdminTagHandler) getTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		tag, err := h.service.Get(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, 10001, "Tag not found")
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(c, ToTagResponse(tag))
	}
}

// createTag handles POST /admin/tags
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
// Also maps frontend "status" (lowercase) to DB enum.
func (h *AdminTagHandler) createTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug"` // Optional: auto-generated from name when empty
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, 10004, "Invalid request")
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

		createdTag, err := h.service.Create(c.Request.Context(), tag)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(c, ToTagResponse(createdTag))
	}
}

// updateTag handles PUT /admin/tags/:id
// B087-R2 Fix: Uses TagResponse DTO for frontend-compatible field names.
func (h *AdminTagHandler) updateTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, 10004, "Invalid request")
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

		updatedTag, err := h.service.Update(c.Request.Context(), id, updates)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		// B087-R2 Fix: Convert to TagResponse DTO
		server.OK(c, ToTagResponse(updatedTag))
	}
}

// deleteTag handles DELETE /admin/tags/:id
func (h *AdminTagHandler) deleteTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := h.service.Delete(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(c, gin.H{
			"code":    0,
			"message": "Tag deleted successfully",
		})
	}
}

// bulkTagOperation handles POST /admin/tags/bulk
func (h *AdminTagHandler) bulkTagOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IDs    []string `json:"ids" binding:"required"`
			Action string   `json:"action" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, 10004, "Invalid request")
			return
		}

		if req.Action != "delete" {
			server.Fail(c, 10004, "Unsupported action")
			return
		}

		count, err := h.service.BulkDelete(c.Request.Context(), req.IDs)
		if err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		server.OK(c, gin.H{
			"success": count,
			"failed":  len(req.IDs) - count,
		})
	}
}

// exportTags handles GET /admin/tags/export
func (h *AdminTagHandler) exportTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement export functionality
		server.OK(c, gin.H{
			"message": "Export functionality not implemented yet",
		})
	}
}

// importTags handles POST /admin/tags/import
func (h *AdminTagHandler) importTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement import functionality
		server.OK(c, gin.H{
			"message": "Import functionality not implemented yet",
		})
	}
}
