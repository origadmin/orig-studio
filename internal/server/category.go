package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
)

func RegisterCategoryRoutes(group *gin.RouterGroup, client *entity.Client) {
	categories := group.Group("/categories")
	{
		categories.GET("", func(c *gin.Context) {
			items, err := client.Category.Query().All(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, items)
		})

		categories.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			cat, err := client.Category.Get(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
				return
			}
			c.JSON(http.StatusOK, cat)
		})

		categories.POST("", func(c *gin.Context) {
			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				IsGlobal    bool   `json:"is_global"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			cat, err := client.Category.Create().
				SetName(input.Name).
				SetDescription(input.Description).
				SetIsGlobal(input.IsGlobal).
				Save(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, cat)
		})

		categories.DELETE("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			client.Category.DeleteOneID(id).Exec(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}
}
