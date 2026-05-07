/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"fmt"
	"math/rand"
	"strconv"

	ginhttp "github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/helpers/hashtag"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/features/content/biz"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemservice "origadmin/application/origcms/internal/features/system/service"
)

type ArticleHandler struct {
	uc        *biz.ArticleUseCase
	jwt       *auth.Manager
	settingUC *systembiz.SettingUseCase
}

func NewArticleHandler(uc *biz.ArticleUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase) *ArticleHandler {
	return &ArticleHandler{uc: uc, jwt: jwt, settingUC: settingUC}
}

// extractUserIDCtx extracts the user ID from JWT claims in an http2.Context.
func extractUserIDCtx(ctx http2.Context) string {
	if claims, ok := server.GetClaimsCtx(ctx); ok {
		return claims.GetUserID()
	}
	return ""
}

// extractUserID extracts the user ID from JWT claims in the Gin context.
// Deprecated: Use extractUserIDCtx instead.
func extractUserID(c *ginhttp.Context) string {
	if claims, exists := c.Get("claims"); exists {
		if cl, ok := claims.(*auth.Claims); ok {
			return cl.GetUserID()
		}
	}
	return ""
}

func (h *ArticleHandler) RegisterRoutes(r http2.Router) {
	g := r.Group("/articles")
	g.Use(systemservice.ModuleGuardCtx(h.settingUC, "module_articles"))

	// ================================
	// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
	// ================================

	// GET /articles - List articles (public, defaults to published only)
	g.GET("", h.listArticles())

	// POST /articles - Create article (requires auth)
	g.POST("", server.WithJWTCtx(h.jwt, h.createArticle()))

	// GET /articles/featured - Get featured articles
	g.GET("/featured", h.listFeaturedArticles())

	// GET /articles/latest - Get latest articles
	g.GET("/latest", h.listLatestArticles())

	// GET /articles/me - List current user's articles (requires auth)
	g.GET("/me", server.WithJWTCtx(h.jwt, h.listMyArticles()))

	// GET /articles/slug/:slug - Get article by slug
	g.GET("/slug/:slug", h.getArticleBySlug())

	// ================================
	// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
	// ================================

	// GET /articles/:id - Get article by ID
	g.GET("/:id", h.getArticle())

	// PUT /articles/:id - Update article (requires auth)
	g.PUT("/:id", server.WithJWTCtx(h.jwt, h.updateArticle()))

	// DELETE /articles/:id - Delete article (requires auth)
	g.DELETE("/:id", server.WithJWTCtx(h.jwt, h.deleteArticle()))

	// PATCH /articles/:id/state - Update article state (requires auth)
	g.PATCH("/:id/state", server.WithJWTCtx(h.jwt, h.updateArticleState()))
}

func (h *ArticleHandler) listArticles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		page, _ := strconv.Atoi(gc.Query("page"))
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(gc.Query("page_size"))
		if pageSize == 0 {
			pageSize = 20
		}
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		filters := map[string]interface{}{}

		state := gc.Query("state")
		if state == "" {
			state = "published"
		}
		filters["state"] = state

		if categoryIDStr := gc.Query("category_id"); categoryIDStr != "" {
			if catID, err := strconv.ParseInt(categoryIDStr, 10, 64); err == nil {
				filters["category_id"] = catID
			}
		}
		if keyword := gc.Query("keyword"); keyword != "" {
			_ = keyword
		}

		items, total, err := h.uc.List(ctx.Request().Context(), page, pageSize, filters)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.Page(ctx, items, int64(total), page, pageSize)
		return nil
	}
}

func (h *ArticleHandler) getArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id := gc.Param("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "article id is required")
			return nil
		}

		article, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}

		http2.OK(ctx, article)
		return nil
	}
}

func (h *ArticleHandler) getArticleBySlug() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		slug := gc.Param("slug")
		if slug == "" {
			http2.Fail(ctx, server.ErrBadRequest, "slug is required")
			return nil
		}

		article, err := h.uc.GetBySlug(ctx.Request().Context(), slug)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}

		http2.OK(ctx, article)
		return nil
	}
}

func (h *ArticleHandler) listFeaturedArticles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		limit, _ := strconv.Atoi(gc.Query("limit"))
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

		items, _, err := h.uc.List(ctx.Request().Context(), 1, limit, filters)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, items)
		return nil
	}
}

func (h *ArticleHandler) listLatestArticles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		limit, _ := strconv.Atoi(gc.Query("limit"))
		if limit == 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}

		filters := map[string]interface{}{
			"state": "published",
		}

		items, _, err := h.uc.List(ctx.Request().Context(), 1, limit, filters)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, items)
		return nil
	}
}

func (h *ArticleHandler) createArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		var input struct {
			Title      string   `json:"title" binding:"required"`
			Slug       string   `json:"slug"`
			Content    string   `json:"content" binding:"required"`
			Summary    string   `json:"summary"`
			CategoryID int64    `json:"category_id"`
			MediaID    string   `json:"media_id"`
			Thumbnail  string   `json:"thumbnail"`
			Tags       []string `json:"tags"`
			Featured   bool     `json:"featured"`
			State      string   `json:"state"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		// Reject archived state from user-side requests
		if input.State == "archived" {
			http2.Fail(ctx, server.ErrBadRequest, "invalid state, must be draft or published")
			return nil
		}

		userID := extractUserIDCtx(ctx)

		slug := input.Slug
		if slug == "" {
			slug = hashtag.GenerateTagSlug(input.Title)
			// Append random suffix to auto-generated slugs to avoid collisions
			// (article slug is Unique in schema; same-titled articles would conflict)
			slug = fmt.Sprintf("%s-%s", slug, randomSuffix(4))
		}

		// Determine state: default to draft, only allow draft or published
		state := "draft"
		if input.State == "published" {
			state = "published"
		}

		article := &biz.Article{
			Title:      input.Title,
			Slug:       slug,
			Content:    input.Content,
			Summary:    input.Summary,
			State:      state,
			Featured:   false, // Always false for user-side; ignore user input
			Tags:       input.Tags,
			UserID:     userID,
			CategoryID: input.CategoryID,
			MediaID:    input.MediaID,
			Thumbnail:  input.Thumbnail,
		}

		created, err := h.uc.Create(ctx.Request().Context(), article)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.Created(ctx, created)
		return nil
	}
}

func (h *ArticleHandler) updateArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		id := gc.Param("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "article id is required")
			return nil
		}

		var input struct {
			Title      string   `json:"title"`
			Slug       string   `json:"slug"`
			Content    string   `json:"content"`
			Summary    string   `json:"summary"`
			CategoryID int64    `json:"category_id"`
			MediaID    string   `json:"media_id"`
			Thumbnail  string   `json:"thumbnail"`
			Tags       []string `json:"tags"`
			Featured   bool     `json:"featured"`
			State      string   `json:"state"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		// Reject archived state from user-side requests
		if input.State == "archived" {
			http2.Fail(ctx, server.ErrBadRequest, "invalid state, must be draft or published")
			return nil
		}

		existing, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}

		// Ownership check: only the article owner can update
		userID := extractUserIDCtx(ctx)
		if existing.UserID != userID {
			http2.Fail(ctx, server.ErrForbidden, "you can only edit your own articles")
			return nil
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
		// Preserve existing featured value; ignore user input
		// existing.Featured is NOT modified from input.Featured
		if input.State != "" {
			existing.State = input.State
		}

		updated, err := h.uc.Update(ctx.Request().Context(), existing)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, updated)
		return nil
	}
}

func (h *ArticleHandler) deleteArticle() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		id := gc.Param("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "article id is required")
			return nil
		}

		article, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}

		// Ownership check: only the article owner can delete
		userID := extractUserIDCtx(ctx)
		if article.UserID != userID {
			http2.Fail(ctx, server.ErrForbidden, "you can only delete your own articles")
			return nil
		}

		// Only allow deleting draft articles
		if article.State == "published" {
			http2.Fail(ctx, server.ErrBadRequest, "published articles cannot be deleted, please contact admin")
			return nil
		}

		if err := h.uc.Delete(ctx.Request().Context(), id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, nil)
		return nil
	}
}

func (h *ArticleHandler) updateArticleState() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		id := gc.Param("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "article id is required")
			return nil
		}

		var input struct {
			State string `json:"state" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		// Only allow draft and published states for user-side
		validUserStates := map[string]bool{"draft": true, "published": true}
		if !validUserStates[input.State] {
			http2.Fail(ctx, server.ErrBadRequest, "invalid state, must be draft or published")
			return nil
		}

		// Ownership check: only the article owner can change state
		article, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}
		userID := extractUserIDCtx(ctx)
		if article.UserID != userID {
			http2.Fail(ctx, server.ErrForbidden, "you can only modify your own articles")
			return nil
		}

		if err := h.uc.UpdateState(ctx.Request().Context(), id, input.State); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		updated, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.OK(ctx, nil)
			return nil
		}

		http2.OK(ctx, updated)
		return nil
	}
}

// listMyArticles returns the current authenticated user's articles (all states).
func (h *ArticleHandler) listMyArticles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		userID := extractUserIDCtx(ctx)
		if userID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "authentication required")
			return nil
		}

		page, _ := strconv.Atoi(gc.Query("page"))
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(gc.Query("page_size"))
		if pageSize == 0 {
			pageSize = 20
		}
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		filters := map[string]interface{}{
			"user_id": userID,
		}

		if state := gc.Query("state"); state != "" {
			filters["state"] = state
		}

		items, total, err := h.uc.List(ctx.Request().Context(), page, pageSize, filters)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.Page(ctx, items, int64(total), page, pageSize)
		return nil
	}
}

// randomSuffix generates a random alphanumeric suffix of the given length
// to avoid slug collisions for auto-generated slugs.
func randomSuffix(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
