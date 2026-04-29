package service

import (
	"net/http"
	"strconv"

	"origadmin/application/origcms/internal/data/entity"

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

// listTags handles GET /tags
func (h *AdminTagHandler) listTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		search := c.Query("search")
		status := c.Query("status")
		sortBy := c.DefaultQuery("sort_by", "created_at")
		sortOrder := c.DefaultQuery("sort_order", "desc")

		// Get tags
		tags, total, err := h.service.List(c.Request.Context(), page, pageSize, search, status, sortBy, sortOrder)
		if err != nil {
			server.Fail(c, 10000, "Failed to list tags")
			return
		}

		// Calculate total pages
		totalPages := (int(total) + pageSize - 1) / pageSize

		// Return response
		server.OK(c, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"items":       tags,
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": totalPages,
			},
		})
	}
}

// getTag handles GET /tags/:id
func (h *AdminTagHandler) getTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		tag, err := h.service.Get(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, 10001, "Tag not found")
			return
		}

		server.OK(c, gin.H{
			"code":    0,
			"message": "ok",
			"data":    tag,
		})
	}
}

// createTag handles POST /tags
func (h *AdminTagHandler) createTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug" binding:"required"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Status      string `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			server.Fail(c, 10004, "Invalid request")
			return
		}

		tag := &entity.Tag{
			Title: req.Name,
		}

		createdTag, err := h.service.Create(c.Request.Context(), tag)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		server.OK(c, gin.H{
			"code":    0,
			"message": "Tag created successfully",
			"data":    createdTag,
		})
	}
}

// updateTag handles PUT /tags/:id
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
			Title: req.Name,
		}

		updatedTag, err := h.service.Update(c.Request.Context(), id, updates)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		server.OK(c, gin.H{
			"code":    0,
			"message": "Tag updated successfully",
			"data":    updatedTag,
		})
	}
}

// deleteTag handles DELETE /tags/:id
func (h *AdminTagHandler) deleteTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := h.service.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		server.OK(c, gin.H{
			"code":    0,
			"message": "Tag deleted successfully",
		})
	}
}

// bulkTagOperation handles POST /tags/bulk
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
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": err.Error()})
			return
		}

		server.OK(c, gin.H{
			"code":    0,
			"message": "Bulk operation completed",
			"data": gin.H{
				"success": count,
				"failed":  len(req.IDs) - count,
			},
		})
	}
}

// exportTags handles GET /tags/export
func (h *AdminTagHandler) exportTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement export functionality
		server.OK(c, gin.H{
			"code":    0,
			"message": "Export functionality not implemented yet",
		})
	}
}

// importTags handles POST /tags/import
func (h *AdminTagHandler) importTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement import functionality
		server.OK(c, gin.H{
			"code":    0,
			"message": "Import functionality not implemented yet",
		})
	}
}
