/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"fmt"
	"math/rand"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/api/gen/v1/types"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
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

		req := &types.ListArticlesRequest{
			Page:     1,
			PageSize: 20,
		}

		if page := gc.Query("page"); page != "" {
			fmt.Sscanf(page, "%d", &req.Page)
		}
		if req.Page <= 0 {
			req.Page = 1
		}

		if pageSize := gc.Query("page_size"); pageSize != "" {
			fmt.Sscanf(pageSize, "%d", &req.PageSize)
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}

		state := gc.Query("state")
		if state == "" {
			state = "published"
		}
		req.State = state

		if categoryIDStr := gc.Query("category_id"); categoryIDStr != "" {
			fmt.Sscanf(categoryIDStr, "%d", &req.CategoryId)
		}
		req.Keyword = gc.Query("keyword")

		resp, err := h.uc.List(ctx.Request().Context(), req)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, resp)
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

		limit := 10
		if limitStr := gc.Query("limit"); limitStr != "" {
			fmt.Sscanf(limitStr, "%d", &limit)
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}

		req := &types.ListArticlesRequest{
			Page:     1,
			PageSize: int32(limit),
			State:    "published",
			Featured: true,
		}

		resp, err := h.uc.List(ctx.Request().Context(), req)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, resp.Articles)
		return nil
	}
}

func (h *ArticleHandler) listLatestArticles() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		limit := 10
		if limitStr := gc.Query("limit"); limitStr != "" {
			fmt.Sscanf(limitStr, "%d", &limit)
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}

		req := &types.ListArticlesRequest{
			Page:     1,
			PageSize: int32(limit),
			State:    "published",
		}

		resp, err := h.uc.List(ctx.Request().Context(), req)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, resp.Articles)
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

		if input.State == "archived" {
			http2.Fail(ctx, server.ErrBadRequest, "invalid state, must be draft or published")
			return nil
		}

		userID := extractUserIDCtx(ctx)

		slug := input.Slug
		if slug == "" {
			slug = hashtag.GenerateTagSlug(input.Title)
			slug = fmt.Sprintf("%s-%s", slug, randomSuffix(4))
		}

		state := "draft"
		if input.State == "published" {
			state = "published"
		}

		article := &types.Article{
			Title:      input.Title,
			Slug:       slug,
			Content:    input.Content,
			Summary:    input.Summary,
			State:      state,
			Featured:   false,
			Tags:       input.Tags,
			UserId:     userID,
			CategoryId: input.CategoryID,
			MediaId:    input.MediaID,
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

		if input.State == "archived" {
			http2.Fail(ctx, server.ErrBadRequest, "invalid state, must be draft or published")
			return nil
		}

		existing, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}

		userID := extractUserIDCtx(ctx)
		if existing.UserId != userID {
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
			existing.CategoryId = input.CategoryID
		}
		if input.MediaID != "" {
			existing.MediaId = input.MediaID
		}
		existing.Thumbnail = input.Thumbnail
		if input.Tags != nil {
			existing.Tags = input.Tags
		}
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

		userID := extractUserIDCtx(ctx)
		if article.UserId != userID {
			http2.Fail(ctx, server.ErrForbidden, "you can only delete your own articles")
			return nil
		}

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

		validUserStates := map[string]bool{"draft": true, "published": true}
		if !validUserStates[input.State] {
			http2.Fail(ctx, server.ErrBadRequest, "invalid state, must be draft or published")
			return nil
		}

		article, err := h.uc.Get(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "article not found")
			return nil
		}

		userID := extractUserIDCtx(ctx)
		if article.UserId != userID {
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

		req := &types.ListArticlesRequest{
			Page:     1,
			PageSize: 20,
			UserId:   userID,
		}

		if page := gc.Query("page"); page != "" {
			fmt.Sscanf(page, "%d", &req.Page)
		}
		if req.Page <= 0 {
			req.Page = 1
		}

		if pageSize := gc.Query("page_size"); pageSize != "" {
			fmt.Sscanf(pageSize, "%d", &req.PageSize)
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}

		if state := gc.Query("state"); state != "" {
			req.State = state
		}

		resp, err := h.uc.List(ctx.Request().Context(), req)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, resp)
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
