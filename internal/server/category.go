package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-content/biz"
)

type CategoryHandler struct {
	uc *biz.CategoryTagUseCase
}

func NewCategoryHandler(uc *biz.CategoryTagUseCase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

func (h *CategoryHandler) Register(r handler.Router) {
	categories := r.Group("/categories")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		categories.GET("", h.listCategories())
		categories.POST("", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Slug        string `json:"slug"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
				return
			}

			cat, err := h.uc.CreateCategory(r.Context(), &biz.Category{
				Name:        input.Name,
				Description: input.Description,
				Slug:        input.Slug,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": cat})
		})

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		categories.GET("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
				return
			}
			cat, err := h.uc.GetCategory(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "category not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": cat})
		})

		categories.PUT("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
				return
			}
			
			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Slug        string `json:"slug"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
				return
			}

			cat, err := h.uc.UpdateCategory(r.Context(), &biz.Category{
				ID:          id,
				Name:        input.Name,
				Description: input.Description,
				Slug:        input.Slug,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": cat})
		})

		categories.DELETE("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid category id"})
				return
			}
			err = h.uc.DeleteCategory(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"message": "deleted"}})
		})
	}
}

func (h *CategoryHandler) listCategories() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)
		page, _ := strconv.Atoi(c.Query("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(c.Query("page_size"))
		if limit == 0 {
			limit = 100
		}
		items, err := h.uc.ListCategories(r.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "ok",
			"data": gin.H{
				"items":     items,
				"total":     int64(len(items)),
				"page":      page,
				"page_size": limit,
			},
		})
	}
}
