package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/svc-content/biz"
)

type CategoryHandler struct {
	uc *biz.CategoryTagUseCase
}

func NewCategoryHandler(uc *biz.CategoryTagUseCase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

func (h *CategoryHandler) Register(group *gin.RouterGroup) {
	categories := group.Group("/categories")
	{
		categories.GET("", h.listCategories())
		categories.GET("/:id", func(c *gin.Context) {
			_, _ = strconv.Atoi(c.Param("id"))
			// UseCase GetCategory implementation needed?
			// cat, err := h.uc.GetCategory(c.Request.Context(), id)
			// For now, CategoryHandler refactoring is minimal.
			c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented in UseCase"})
		})

		categories.POST("", func(c *gin.Context) {
			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			cat, err := h.uc.CreateCategory(c.Request.Context(), &biz.Category{
				Name:        input.Name,
				Description: input.Description,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, cat)
		})

		categories.DELETE("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			err := h.uc.DeleteCategory(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}
}

func (h *CategoryHandler) listCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := h.uc.ListCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}
