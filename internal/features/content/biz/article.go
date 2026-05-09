/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"

	"origadmin/application/origcms/api/gen/v1/types"

	"github.com/go-kratos/kratos/v2/log"
)

// Article represents a content article (legacy internal type, migrating to Proto).
// Deprecated: Use types.Article instead. This struct is kept for backward compatibility.
type Article struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Summary      string   `json:"summary,omitempty"`
	Slug         string   `json:"slug,omitempty"`
	ShortToken   string   `json:"short_token,omitempty"`
	State        string   `json:"state"`
	ViewCount    int64    `json:"view_count"`
	CommentCount int64    `json:"comment_count"`
	Featured     bool     `json:"featured"`
	Tags         []string `json:"tags,omitempty"`
	UserID       string   `json:"user_id"`
	CategoryID   int64    `json:"category_id,omitempty"`
	MediaID      string   `json:"media_id,omitempty"`
	Thumbnail    string   `json:"thumbnail,omitempty"`
}

// ArticleRepo defines storage operations for articles using Proto types.
type ArticleRepo interface {
	Create(ctx context.Context, article *types.Article) (*types.Article, error)
	Update(ctx context.Context, article *types.Article) (*types.Article, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*types.Article, error)
	GetBySlug(ctx context.Context, slug string) (*types.Article, error)
	GetByShortToken(ctx context.Context, shortToken string) (*types.Article, error)
	List(ctx context.Context, filters *types.ListArticlesRequest) (*types.ListArticlesResponse, error)
	UpdateState(ctx context.Context, id string, state string) error
}

// ArticleUseCase handles article business logic using Proto types.
type ArticleUseCase struct {
	repo ArticleRepo
	log  *log.Helper
}

// NewArticleUseCase creates a new ArticleUseCase.
func NewArticleUseCase(repo ArticleRepo, logger log.Logger) *ArticleUseCase {
	return &ArticleUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "article.biz")),
	}
}

// Create creates a new article.
func (uc *ArticleUseCase) Create(ctx context.Context, article *types.Article) (*types.Article, error) {
	if article.State == "" {
		article.State = "draft"
	}
	return uc.repo.Create(ctx, article)
}

// Update updates an existing article.
func (uc *ArticleUseCase) Update(ctx context.Context, article *types.Article) (*types.Article, error) {
	return uc.repo.Update(ctx, article)
}

// Delete deletes an article by ID.
func (uc *ArticleUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// Get gets an article by ID.
func (uc *ArticleUseCase) Get(ctx context.Context, id string) (*types.Article, error) {
	return uc.repo.Get(ctx, id)
}

// GetBySlug gets an article by slug.
func (uc *ArticleUseCase) GetBySlug(ctx context.Context, slug string) (*types.Article, error) {
	return uc.repo.GetBySlug(ctx, slug)
}

// GetByShortToken gets an article by short token.
func (uc *ArticleUseCase) GetByShortToken(ctx context.Context, shortToken string) (*types.Article, error) {
	return uc.repo.GetByShortToken(ctx, shortToken)
}

// List lists articles with pagination and filters.
func (uc *ArticleUseCase) List(ctx context.Context, filters *types.ListArticlesRequest) (*types.ListArticlesResponse, error) {
	return uc.repo.List(ctx, filters)
}

// UpdateState updates the state of an article.
func (uc *ArticleUseCase) UpdateState(ctx context.Context, id string, state string) error {
	return uc.repo.UpdateState(ctx, id, state)
}
