/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Article represents a content article.
type Article struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Summary      string    `json:"summary,omitempty"`
	Slug         string    `json:"slug,omitempty"`
	State        string    `json:"state"` // draft / published / archived
	ViewCount    int64     `json:"view_count"`
	CommentCount int64     `json:"comment_count"`
	Featured     bool      `json:"featured"`
	Tags         []string  `json:"tags,omitempty"`
	UserID       string    `json:"user_id"`
	CategoryID   int64     `json:"category_id,omitempty"`
	PublishedAt  time.Time `json:"published_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// ArticleRepo defines storage operations for articles.
type ArticleRepo interface {
	Create(ctx context.Context, article *Article) (*Article, error)
	Update(ctx context.Context, article *Article) (*Article, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Article, error)
	GetBySlug(ctx context.Context, slug string) (*Article, error)
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*Article, int64, error)
	UpdateState(ctx context.Context, id string, state string) error
}

// ArticleUseCase handles article business logic.
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
func (uc *ArticleUseCase) Create(ctx context.Context, article *Article) (*Article, error) {
	if article.State == "" {
		article.State = "draft"
	}
	return uc.repo.Create(ctx, article)
}

// Update updates an existing article.
func (uc *ArticleUseCase) Update(ctx context.Context, article *Article) (*Article, error) {
	return uc.repo.Update(ctx, article)
}

// Delete deletes an article by ID.
func (uc *ArticleUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// Get gets an article by ID.
func (uc *ArticleUseCase) Get(ctx context.Context, id string) (*Article, error) {
	return uc.repo.Get(ctx, id)
}

// GetBySlug gets an article by slug.
func (uc *ArticleUseCase) GetBySlug(ctx context.Context, slug string) (*Article, error) {
	return uc.repo.GetBySlug(ctx, slug)
}

// List lists articles with pagination and filters.
func (uc *ArticleUseCase) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*Article, int64, error) {
	return uc.repo.List(ctx, page, pageSize, filters)
}

// UpdateState updates the state of an article.
func (uc *ArticleUseCase) UpdateState(ctx context.Context, id string, state string) error {
	return uc.repo.UpdateState(ctx, id, state)
}
