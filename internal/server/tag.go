package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/svc-content/biz"
)

type TagHandler struct {
	uc *biz.CategoryTagUseCase
}

func NewTagHandler(uc *biz.CategoryTagUseCase) *TagHandler {
	return &TagHandler{uc: uc}
}

func (h *TagHandler) Register(group *gin.RouterGroup) {
	tags := group.Group("/tags")
	{
		tags.GET("", h.listTags())
		tags.GET("/:id", func(c *gin.Context) {
			_, _ = strconv.Atoi(c.Param("id"))
			// UseCase GetTag implementation needed?
			c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented in UseCase"})
		})

		// GET /tags/:tag_id/media — list media by tag
		tags.GET("/:id/media", h.getMediaByTag())

		tags.POST("", func(c *gin.Context) {
			var input struct {
				Title string `json:"title"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			t, err := h.uc.CreateTag(c.Request.Context(), &biz.Tag{
				Title: input.Title,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, t)
		})

		tags.DELETE("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			err := h.uc.DeleteTag(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}
}

func (h *TagHandler) listTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
		items, total, err := h.uc.ListTags(c.Request.Context(), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"list":  items,
			"total": total,
		})
	}
}

// getMediaByTag returns all media associated with a specific tag.
// GET /api/v1/tags/:id/media
func (h *TagHandler) getMediaByTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This requires MediaUseCase or a more complex query in UseCase
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented in UseCase"})
	}
}
