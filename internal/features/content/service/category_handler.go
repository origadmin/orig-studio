package service

import (
	"origadmin/application/origcms/internal/handler"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/features/content/biz"
)

type CategoryHandler struct {
	uc  *biz.CategoryTagUseCase
	jwt *auth.Manager
}

func NewCategoryHandler(uc *biz.CategoryTagUseCase, jwt *auth.Manager) *CategoryHandler {
	return &CategoryHandler{uc: uc, jwt: jwt}
}

func (h *CategoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	categories := r.Group("/categories")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		categories.GET("", h.listCategories())
		categories.POST("", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
												c := handler.NewGinContextAdapterFromHTTP(w, r)
var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Slug        string `json:"slug"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": err.Error()})
				return
			}

			cat, err := h.uc.CreateCategory(r.Context(), &biz.Category{
				Name:        input.Name,
				Description: input.Description,
				Slug:        input.Slug,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": server.ErrInternal, "message": err.Error()})
				return
			}

			server.OK(c.GinContext(), gin.H{"code": 0, "message": "ok", "data": cat})
		}))

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		categories.GET("/:id", func(w http.ResponseWriter, r *http.Request) {
												c := handler.NewGinContextAdapterFromHTTP(w, r)
id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": "invalid category id"})
				return
			}
			cat, err := h.uc.GetCategory(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": server.ErrNotFound, "message": "category not found"})
				return
			}
			server.OK(c.GinContext(), gin.H{"code": 0, "message": "ok", "data": cat})
		})

		categories.PUT("/:id", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
												c := handler.NewGinContextAdapterFromHTTP(w, r)
id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": "invalid category id"})
				return
			}

			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Slug        string `json:"slug"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": err.Error()})
				return
			}

			cat, err := h.uc.UpdateCategory(r.Context(), &biz.Category{
				ID:          id,
				Name:        input.Name,
				Description: input.Description,
				Slug:        input.Slug,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": server.ErrInternal, "message": err.Error()})
				return
			}

			server.OK(c.GinContext(), gin.H{"code": 0, "message": "ok", "data": cat})
		}))

		categories.PATCH("/:id", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
												c := handler.NewGinContextAdapterFromHTTP(w, r)
id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": "invalid category id"})
				return
			}

			var input biz.UpdateCategoryInput
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": err.Error()})
				return
			}

			cat, err := h.uc.UpdateCategoryPartial(r.Context(), id, &input)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": server.ErrInternal, "message": err.Error()})
				return
			}

			server.OK(c.GinContext(), gin.H{"code": 0, "message": "ok", "data": cat})
		}))

		categories.DELETE("/:id", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
												c := handler.NewGinContextAdapterFromHTTP(w, r)
id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": server.ErrBadRequest, "message": "invalid category id"})
				return
			}
			err = h.uc.DeleteCategory(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": server.ErrInternal, "message": err.Error()})
				return
			}
			server.OK(c.GinContext(), gin.H{"code": 0, "message": "ok", "data": gin.H{"message": "deleted"}})
		}))
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
			c.JSON(http.StatusInternalServerError, gin.H{"code": server.ErrInternal, "message": err.Error()})
			return
		}
		server.OK(c.GinContext(), gin.H{
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
