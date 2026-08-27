/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/pkg/hashtag"
)

// maxSlugAttempts bounds the "-2", "-3", ... suffix probing used to resolve
// slug collisions before falling back to a timestamp suffix.
const maxSlugAttempts = 50

// Category represents a media category.
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	ParentID    int64     `json:"parent_id"`
	Sequence    int       `json:"order"`
	MediaCount  int       `json:"media_count"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// UpdateCategoryInput represents partial update input for a category.
type UpdateCategoryInput struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
	ParentID    *int64  `json:"parent_id"`
	Sequence    *int    `json:"order"`
}

// Tag represents a media tag.
type Tag struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	// Status holds the proto TagStatus enum value (0=UNSPECIFIED, 1=INACTIVE, 2=ACTIVE).
	Status      int       `json:"status"`
	MediaCount  int       `json:"media_count"`
	CreateTime  time.Time `json:"create_time"`
}

// CategoryRepo defines storage operations for categories.
type CategoryRepo interface {
	Create(ctx context.Context, c *Category) (*Category, error)
	Get(ctx context.Context, id int) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	Update(ctx context.Context, c *Category) (*Category, error)
	Delete(ctx context.Context, id int) error
	ListAll(ctx context.Context) ([]*Category, error)
	ListActive(ctx context.Context) ([]*Category, error)
}

// TagRepo defines storage operations for tags.
type TagRepo interface {
	Create(ctx context.Context, t *Tag) (*Tag, error)
	Get(ctx context.Context, id int) (*Tag, error)
	GetByName(ctx context.Context, name string) (*Tag, error)
	GetBySlug(ctx context.Context, slug string) (*Tag, error)
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

func NewCategoryTagUseCase(
	catRepo CategoryRepo,
	tagRepo TagRepo,
	logger log.Logger,
) *CategoryTagUseCase {
	return &CategoryTagUseCase{
		categoryRepo: catRepo,
		tagRepo:      tagRepo,
		log:          log.NewHelper(log.With(logger, "module", "category_tag.biz")),
	}
}

func (uc *CategoryTagUseCase) ListCategories(ctx context.Context) ([]*Category, error) {
	return uc.categoryRepo.ListAll(ctx)
}

func (uc *CategoryTagUseCase) ListActiveCategories(ctx context.Context) ([]*Category, error) {
	return uc.categoryRepo.ListActive(ctx)
}

func (uc *CategoryTagUseCase) CreateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.categoryRepo.Create(ctx, c)
}

func (uc *CategoryTagUseCase) UpdateCategory(ctx context.Context, c *Category) (*Category, error) {
	return uc.categoryRepo.Update(ctx, c)
}

func (uc *CategoryTagUseCase) UpdateCategoryPartial(ctx context.Context, id int, input *UpdateCategoryInput) (*Category, error) {
	cat, err := uc.categoryRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Apply partial updates: only update fields that are provided (non-nil)
	if input.Name != nil {
		cat.Name = *input.Name
	}
	if input.Slug != nil {
		cat.Slug = *input.Slug
	}
	if input.Description != nil {
		cat.Description = *input.Description
	}
	if input.Status != nil {
		cat.Status = *input.Status
	}
	if input.ParentID != nil {
		cat.ParentID = *input.ParentID
	}
	if input.Sequence != nil {
		cat.Sequence = *input.Sequence
	}
	return uc.categoryRepo.Update(ctx, cat)
}

func (uc *CategoryTagUseCase) DeleteCategory(ctx context.Context, id int) error {
	return uc.categoryRepo.Delete(ctx, id)
}

func (uc *CategoryTagUseCase) GetCategory(ctx context.Context, id int) (*Category, error) {
	return uc.categoryRepo.Get(ctx, id)
}

func (uc *CategoryTagUseCase) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	return uc.categoryRepo.GetBySlug(ctx, slug)
}

func (uc *CategoryTagUseCase) ListTags(
	ctx context.Context,
	page, pageSize int,
) ([]*Tag, int, error) {
	return uc.tagRepo.ListAll(ctx, page, pageSize)
}

// ensureTagSlug guarantees a tag carries a non-empty, collision-free slug.
//
// BUG-143 root cause: the portal create endpoint (POST /api/v1/tags) only sends
// a title, so the slug column stayed NULL. Those tags then 404'd on
// GET /api/v1/tags/{slug} and could not back a dedicated /tag/{slug} page.
// Slug generation is a business invariant, so it is enforced here rather than
// in each handler.
func (uc *CategoryTagUseCase) ensureTagSlug(ctx context.Context, t *Tag) {
	if t == nil || strings.TrimSpace(t.Slug) != "" {
		return
	}

	base := hashtag.GenerateTagSlug(t.Title)
	candidate := base
	for i := 2; i <= maxSlugAttempts; i++ {
		existing, err := uc.tagRepo.GetBySlug(ctx, candidate)
		// A lookup error (or no row) means the slug is free.
		if err != nil || existing == nil {
			t.Slug = candidate
			return
		}
		// Updating a tag that already owns this slug is not a collision.
		if t.ID != 0 && existing.ID == t.ID {
			t.Slug = candidate
			return
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}

	// Pathological collision run: fall back to a guaranteed-unique suffix.
	t.Slug = fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
}

func (uc *CategoryTagUseCase) CreateTag(ctx context.Context, t *Tag) (*Tag, error) {
	uc.ensureTagSlug(ctx, t)
	return uc.tagRepo.Create(ctx, t)
}

func (uc *CategoryTagUseCase) DeleteTag(ctx context.Context, id int) error {
	return uc.tagRepo.Delete(ctx, id)
}

func (uc *CategoryTagUseCase) GetTag(ctx context.Context, id int) (*Tag, error) {
	return uc.tagRepo.Get(ctx, id)
}

// GetTagBySlug resolves a tag by its URL slug.
//
// BUG-146 defense: a tag that still appears in media.tags jsonb can exist in
// content_tags with a NULL/empty slug — rows created before slug generation
// was enforced (BUG-143). Because the slug in the URL equals
// GenerateTagSlug(title), a plain slug lookup 404s and the frontend is left to
// render a false "no videos" empty state. When the slug lookup misses we fall
// back to a title match, and for legacy empty-slug rows to the generated slug
// of each title, so the detail page still resolves instead of 404-ing.
func (uc *CategoryTagUseCase) GetTagBySlug(ctx context.Context, slug string) (*Tag, error) {
	t, slugErr := uc.tagRepo.GetBySlug(ctx, slug)
	if slugErr == nil {
		return t, nil
	}
	if !entity.IsNotFound(slugErr) {
		return nil, slugErr
	}

	// Degenerate case: the slug literally equals a stored title.
	if byTitle, err := uc.tagRepo.GetByName(ctx, slug); err == nil {
		return byTitle, nil
	}

	// Legacy rows whose slug was never generated: match on the generated slug
	// of each title. The tags table is small and this runs only on the rare
	// slug-miss path.
	tags, _, err := uc.tagRepo.ListAll(ctx, 1, 1000)
	if err != nil {
		return nil, err
	}
	for _, candidate := range tags {
		if candidate.Slug == "" && hashtag.GenerateTagSlug(candidate.Title) == slug {
			return candidate, nil
		}
	}

	// Genuinely unknown slug — preserve the original not-found error so the
	// handler still maps it to 404.
	return nil, slugErr
}

func (uc *CategoryTagUseCase) UpdateTag(ctx context.Context, t *Tag) (*Tag, error) {
	// Backfill legacy rows created before slug generation was enforced.
	uc.ensureTagSlug(ctx, t)
	return uc.tagRepo.Update(ctx, t)
}
