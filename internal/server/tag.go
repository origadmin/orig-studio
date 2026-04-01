package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
)

func RegisterTagRoutes(group *gin.RouterGroup, client *entity.Client) {
	tags := group.Group("/tags")
	{
		tags.GET("", func(c *gin.Context) {
			items, err := client.Tag.Query().All(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, items)
		})

		tags.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			tag, err := client.Tag.Get(c.Request.Context(), id)
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

			tag, err := client.Tag.Create().
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
