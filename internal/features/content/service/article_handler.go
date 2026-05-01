/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/features/content/biz"
)

type ArticleHandler struct {
	uc  *biz.ArticleUseCase
	jwt *auth.Manager
}

func NewArticleHandler(uc *biz.ArticleUseCase, jwt *auth.Manager) *ArticleHandler {
	return &ArticleHandler{uc: uc, jwt: jwt}
}

func (h *ArticleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	articles := r.Group("/articles")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================

		// GET /articles - List articles (public, defaults to published only)
		articles.GET("", h.listArticles())

		// POST /articles - Create article (requires auth)
		articles.POST("", server.WithJWT(h.jwt, h.createArticle()))

		// GET /articles/featured - Get featured articles
		articles.GET("/featured", h.listFeaturedArticles())

		// GET /articles/latest - Get latest articles
		articles.GET("/latest", h.listLatestArticles())

		// GET /articles/slug/:slug - Get article by slug
		articles.GET("/slug/:slug", h.getArticleBySlug())

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================

		// GET /articles/:id - Get article by ID
		articles.GET("/:id", h.getArticle())

		// PUT /articles/:id - Update article (requires auth)
		articles.PUT("/:id", server.WithJWT(h.jwt, h.updateArticle()))

		// DELETE /articles/:id - Delete article (requires auth)
		articles.DELETE("/:id", server.WithJWT(h.jwt, h.deleteArticle()))

		// PATCH /articles/:id/state - Update article state (requires auth)
		articles.PATCH("/:id/state", server.WithJWT(h.jwt, h.updateArticleState()))
	}
}

func (h *ArticleHandler) listArticles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)

		page, _ := strconv.Atoi(c.Query("page"))
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(c.Query("page_size"))
		if pageSize == 0 {
			pageSize = 20
		}
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		filters := map[string]interface{}{}

		state := c.Query("state")
		if state == "" {
			state = "published"
		}
		filters["state"] = state

		if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
			if catID, err := strconv.ParseInt(categoryIDStr, 10, 64); err == nil {
				filters["category_id"] = catID
			}
		}
		if keyword := c.Query("keyword"); keyword != "" {
			_ = keyword
		}

		items, total, err := h.uc.List(r.Context(), page, pageSize, filters)
		if err != nil {
			server.Fail(c.GinContext(), server.ErrInternal, err.Error())
			return
		}

		server.Page(c.GinContext(), items, int64(total), page, pageSize)
	}
}

func (h *ArticleHandler) getArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)
		id := c.Param("id")
		if id == "" {
			server.Fail(c.GinContext(), server.ErrBadRequest, "article id is required")
			return
		}

		article, err := h.uc.Get(r.Context(), id)
		if err != nil {
			server.Fail(c.GinContext(), server.ErrNotFound, "article not found")
			return
		}

		server.OK(c.GinContext(), article)
	}
}

func (h *ArticleHandler) getArticleBySlug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)
		slug := c.Param("slug")
		if slug == "" {
			server.Fail(c.GinContext(), server.ErrBadRequest, "slug is required")
			return
		}

		article, err := h.uc.GetBySlug(r.Context(), slug)
		if err != nil {
			server.Fail(c.GinContext(), server.ErrNotFound, "article not found")
			return
		}

		server.OK(c.GinContext(), article)
	}
}

func (h *ArticleHandler) listFeaturedArticles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)

		limit, _ := strconv.Atoi(c.Query("limit"))
		if limit == 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}

		filters := map[string]interface{}{
			"state":    "published",
			"featured": true,
		}

		items, _, err := h.uc.List(r.Context(), 1, limit, filters)
		if err != nil {
			server.Fail(c.GinContext(), server.ErrInternal, err.Error())
			return
		}

		server.OK(c.GinContext(), items)
	}
}

func (h *ArticleHandler) listLatestArticles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)

		limit, _ := strconv.Atoi(c.Query("limit"))
		if limit == 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}

		filters := map[string]interface{}{
			"state": "published",
		}

		items, _, err := h.uc.List(r.Context(), 1, limit, filters)
		if err != nil {
			server.Fail(c.GinContext(), server.ErrInternal, err.Error())
			return
		}

		server.OK(c.GinContext(), items)
	}
}

func (h *ArticleHandler) createArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title       string   `json:"title" binding:"required"`
			Slug        string   `json:"slug"`
			Content     string   `json:"content" binding:"required"`
			Summary     string   `json:"summary"`
			CategoryID  int64    `json:"category_id"`
			MediaID     string   `json:"media_id"`
			Thumbnail   string   `json:"thumbnail"`
			Tags        []string `json:"tags"`
			Featured    bool     `json:"featured"`
			PublishedAt string   `json:"published_at"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		userID := ""
		if claims, exists := c.Get("claims"); exists {
			if cl, ok := claims.(*auth.Claims); ok {
				userID = cl.GetUserID()
			}
		}

		article := &biz.Article{
			Title:      input.Title,
			Slug:       input.Slug,
			Content:    input.Content,
			Summary:    input.Summary,
			State:      "draft",
			Featured:   input.Featured,
			Tags:       input.Tags,
			UserID:     userID,
			CategoryID: input.CategoryID,
			MediaID:    input.MediaID,
			Thumbnail:  input.Thumbnail,
		}

		created, err := h.uc.Create(c.Request.Context(), article)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.Created(c, created)
	}
}

func (h *ArticleHandler) updateArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		var input struct {
			Title       string   `json:"title"`
			Slug        string   `json:"slug"`
			Content     string   `json:"content"`
			Summary     string   `json:"summary"`
			CategoryID  int64    `json:"category_id"`
			MediaID     string   `json:"media_id"`
			Thumbnail   string   `json:"thumbnail"`
			Tags        []string `json:"tags"`
			Featured    bool     `json:"featured"`
			State       string   `json:"state"`
			PublishedAt string   `json:"published_at"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		existing, err := h.uc.Get(c.Request.Context(), id)
		if err != nil {
			server.Fail(c, server.ErrNotFound, "article not found")
			return
		}

		if input.Title != "" {
			existing.Title = input.Title
		}
		if input.Slug != "" {
			existing.Slug = input.Slug
		}
		if input.Content != "" {
			existing.Content = input.Content
		}
		if input.Summary != "" {
			existing.Summary = input.Summary
		}
		if input.CategoryID != 0 {
			existing.CategoryID = input.CategoryID
		}
		if input.MediaID != "" {
			existing.MediaID = input.MediaID
		}
		existing.Thumbnail = input.Thumbnail // Allow empty string to clear
		if input.Tags != nil {
			existing.Tags = input.Tags
		}
		existing.Featured = input.Featured
		if input.State != "" {
			existing.State = input.State
		}

		updated, err := h.uc.Update(c.Request.Context(), existing)
		if err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, updated)
	}
}

func (h *ArticleHandler) deleteArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		if err := h.uc.Delete(c.Request.Context(), id); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		server.OK(c, nil)
	}
}

func (h *ArticleHandler) updateArticleState() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			server.Fail(c, server.ErrBadRequest, "article id is required")
			return
		}

		var input struct {
			State string `json:"state" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			server.Fail(c, server.ErrBadRequest, err.Error())
			return
		}

		validStates := map[string]bool{"draft": true, "published": true, "archived": true}
		if !validStates[input.State] {
			server.Fail(c, server.ErrBadRequest, "invalid state, must be one of: draft, published, archived")
			return
		}

		if err := h.uc.UpdateState(c.Request.Context(), id, input.State); err != nil {
			server.Fail(c, server.ErrInternal, err.Error())
			return
		}

		article, err := h.uc.Get(c.Request.Context(), id)
		if err != nil {
			server.OK(c, nil)
			return
		}

		server.OK(c, article)
	}
}
