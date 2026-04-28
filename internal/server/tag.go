package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-content/biz"
)

type TagHandler struct {
	uc  *biz.CategoryTagUseCase
	jwt *auth.Manager
}

func NewTagHandler(uc *biz.CategoryTagUseCase, jwt *auth.Manager) *TagHandler {
	return &TagHandler{uc: uc, jwt: jwt}
}

func (h *TagHandler) Register(r handler.Router) {
	tags := r.Group("/tags")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		tags.GET("", h.listTags())
		tags.POST("", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			var input struct {
				Title string `json:"title"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
				return
			}

			t, err := h.uc.CreateTag(r.Context(), &biz.Tag{
				Title: input.Title,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": t})
		}))

		// ================================
		// 2. NESTED RESOURCE ROUTES (WITH :id) - MUST BE BEFORE MAIN :id ROUTES
		// ================================
		// GET /tags/:tag_id/media — list media by tag
		tags.GET("/:id/media", h.getMediaByTag())

		// ================================
		// 3. MAIN RESOURCE PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		tags.GET("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid tag id"})
				return
			}
			t, err := h.uc.GetTag(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": t})
		})

		tags.PUT("/:id", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": "invalid tag id"})
				return
			}
			var input struct {
				Title string `json:"title"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
				return
			}

			t, err := h.uc.UpdateTag(r.Context(), &biz.Tag{
				ID:    id,
				Title: input.Title,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": t})
		}))

		tags.DELETE("/:id", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, _ := strconv.Atoi(c.Param("id"))
			err := h.uc.DeleteTag(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"message": "deleted"}})
		}))
	}
}

func (h *TagHandler) listTags() http.HandlerFunc {
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
		items, total, err := h.uc.ListTags(r.Context(), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": ErrInternal, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "ok",
			"data": gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": limit,
			},
		})
	}
}

// getMediaByTag returns all media associated with a specific tag.
// GET /api/v1/tags/:id/media
func (h *TagHandler) getMediaByTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)
		// This requires MediaUseCase or a more complex query in UseCase
		c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "not implemented in UseCase"})
	}
}
