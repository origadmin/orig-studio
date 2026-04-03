package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
)

type CategoryHandler struct {
	client *entity.Client
}

func NewCategoryHandler(client *entity.Client) *CategoryHandler {
	return &CategoryHandler{client: client}
}

func (h *CategoryHandler) Register(group *gin.RouterGroup) {
	categories := group.Group("/categories")
	{
		categories.GET("", h.listCategories())
		categories.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			cat, err := h.client.Category.Get(c.Request.Context(), id)
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

			cat, err := h.client.Category.Create().
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
			h.client.Category.DeleteOneID(id).Exec(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}
}

func (h *CategoryHandler) listCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := h.client.Category.Query().All(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}
