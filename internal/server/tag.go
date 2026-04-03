package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
	entitymedia "origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/mediatag"
	"origadmin/application/origcms/internal/data/entity/tag"
)

type TagHandler struct {
	client *entity.Client
}

func NewTagHandler(client *entity.Client) *TagHandler {
	return &TagHandler{client: client}
}

func (h *TagHandler) Register(group *gin.RouterGroup) {
	tags := group.Group("/tags")
	{
		tags.GET("", h.listTags())
		tags.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			tag, err := h.client.Tag.Get(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
				return
			}
			c.JSON(http.StatusOK, tag)
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

			t, err := h.client.Tag.Create().
				SetTitle(input.Title).
				Save(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, t)
		})

		tags.DELETE("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			err := h.client.Tag.DeleteOneID(id).Exec(c.Request.Context())
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
		items, err := h.client.Tag.Query().All(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

// getMediaByTarget returns all media associated with a specific tag.
// GET /api/v1/tags/:id/media
func (h *TagHandler) getMediaByTag() gin.HandlerFunc {
	return func(c *gin.Context) {
		tagID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
			return
		}

		ctx := c.Request.Context()

		// Verify tag exists
		_, err = h.client.Tag.Get(ctx, tagID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		offset := (page - 1) * pageSize

		// Query media through MediaTag junction table
		items, err := h.client.Media.Query().
			Where(entitymedia.HasTagsWith(tag.IDEQ(tagID))).
			Where(entitymedia.StateEQ("active")).
			WithUser().
			WithCategory().
			Limit(pageSize).
			Offset(offset).
			Order(entity.Desc("created_at")).
			All(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		total, _ := h.client.Media.Query().
			Where(entitymedia.HasTagsWith(tag.IDEQ(tagID))).
			Where(entitymedia.StateEQ("active")).
			Count(ctx)

		c.JSON(http.StatusOK, gin.H{
			"list":      items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}
