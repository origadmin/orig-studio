/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Category represents a media category.
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Tag represents a media tag.
type Tag struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	MediaCount int    `json:"media_count"`
}

// CategoryRepo defines storage operations for categories.
type CategoryRepo interface {
	Create(ctx context.Context, c *Category) (*Category, error)
	Get(ctx context.Context, id int) (*Category, error)
	Update(ctx context.Context, c *Category) (*Category, error)
	Delete(ctx context.Context, id int) error
	ListAll(ctx context.Context) ([]*Category, error)
}

// TagRepo defines storage operations for tags.
type TagRepo interface {
	Create(ctx context.Context, t *Tag) (*Tag, error)
	Get(ctx context.Context, id int) (*Tag, error)
	GetByName(ctx context.Context, name string) (*Tag, error)
	Update(ctx context.Context, t *Tag) (*Tag, error)
	Delete(ctx context.Context, id int) error
	ListAll(ctx context.Context, page, pageSize int) ([]*Tag, int, error)
}

// CategoryTagUseCase handles category and tag business logic.
type CategoryTagUseCase struct {
	categoryRepo CategoryRepo
	tagRepo      TagRepo
	log          *log.Helper
}

func NewCategoryTagUseCase(catRepo CategoryRepo, tagRepo TagRepo, logger log.Logger) *CategoryTagUseCase {
	return &CategoryTagUseCase{
		categoryRepo: catRepo,
		tagRepo:      tagRepo,
		log:          log.NewHelper(log.With(logger, "module", "category_tag.biz")),
	}
}

func (uc *CategoryTagUseCase) ListCategories(ctx context.Context) ([]*Category, error) {
	return uc.categoryRepo.ListAll(ctx)
}

func (uc *CategoryTagUseCase) CreateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.categoryRepo.Create(ctx, c)
}

func (uc *CategoryTagUseCase) UpdateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.categoryRepo.Update(ctx, c)
}

func (uc *CategoryTagUseCase) DeleteCategory(ctx context.Context, id int) error {
	return uc.categoryRepo.Delete(ctx, id)
}

func (uc *CategoryTagUseCase) ListTags(ctx context.Context, page, pageSize int) ([]*Tag, int, error) {
	return uc.tagRepo.ListAll(ctx, page, pageSize)
}

func (uc *CategoryTagUseCase) CreateTag(ctx context.Context, t *Tag) (*Tag, error) {
	return uc.tagRepo.Create(ctx, t)
}

func (uc *CategoryTagUseCase) DeleteTag(ctx context.Context, id int) error {
	return uc.tagRepo.Delete(ctx, id)
}
