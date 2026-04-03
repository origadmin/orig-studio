package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
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

		tags.POST("", func(c *gin.Context) {
			var input struct {
				Title string `json:"title"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			tag, err := h.client.Tag.Create().
				SetTitle(input.Title).
				Save(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, tag)
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
