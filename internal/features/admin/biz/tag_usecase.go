package biz

import (
	"context"
	"errors"
	"fmt"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/features/admin/dal"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/helpers/tagcolor"
)

// TagUseCase handles tag business logic
type TagUseCase struct {
	repo dal.TagRepository
}

// NewTagUseCase creates a new TagUseCase
func NewTagUseCase(repo dal.TagRepository) *TagUseCase {
	return &TagUseCase{repo: repo}
}

// List returns a paginated list of tags
func (uc *TagUseCase) List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*entity.Tag, int64, error) {
	// Normalize pagination parameters using centralized validation
	page, pageSize = repo.NormalizePagination(page, pageSize)

	tags, total, err := uc.repo.List(ctx, page, pageSize, search, status, sortBy, sortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}

	return tags, total, nil
}

// Get returns a tag by ID
func (uc *TagUseCase) Get(ctx context.Context, id string) (*entity.Tag, error) {
	if id == "" {
		return nil, errors.New("tag id is required")
	}

	tag, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}

	return tag, nil
}

// Create creates a new tag
func (uc *TagUseCase) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	if tag.Title == "" {
		return nil, errors.New("tag title is required")
	}

	// B087-R2 Fix: Check if title is unique using GetByName (not GetBySlug).
	// GetBySlug checks the slug field, not the title field.
	existingTag, err := uc.repo.GetByName(ctx, tag.Title)
	if err == nil && existingTag != nil {
		return nil, errors.New("tag title already exists")
	}

	if tag.Color == "" {
		tag.Color = tagcolor.ColorFromName(tag.Title)
	}

	createdTag, err := uc.repo.Create(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	return createdTag, nil
}

// Update updates an existing tag
func (uc *TagUseCase) Update(ctx context.Context, id string, updates *entity.Tag) (*entity.Tag, error) {
	if id == "" {
		return nil, errors.New("tag id is required")
	}

	// Get existing tag
	tag, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}

	// Update fields
	if updates.Title != "" {
		// Check if title is unique
		existingTag, err := uc.repo.GetBySlug(ctx, updates.Title)
		if err == nil && existingTag != nil {
			tagID := fmt.Sprintf("%d", existingTag.ID)
			if tagID != id {
				return nil, errors.New("tag title already exists")
			}
		}
		tag.Title = updates.Title
	}
	if updates.Slug != "" {
		tag.Slug = updates.Slug
	}
	if updates.Color != "" {
		tag.Color = updates.Color
	} else {
		tag.Color = ""
	}
	if updates.Description != "" {
		tag.Description = updates.Description
	} else {
		tag.Description = ""
	}
	if updates.Status != "" {
		tag.Status = updates.Status
	}

	updatedTag, err := uc.repo.Update(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}

	return updatedTag, nil
}

// Delete deletes a tag by ID
func (uc *TagUseCase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("tag id is required")
	}

	// Check if tag exists
	_, err := uc.repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get tag: %w", err)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	return nil
}

// BulkDelete deletes multiple tags by IDs
func (uc *TagUseCase) BulkDelete(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("tag ids are required")
	}

	count, err := uc.repo.BulkDelete(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("bulk delete tags: %w", err)
	}

	return count, nil
}

// Count returns tag counts
func (uc *TagUseCase) Count(ctx context.Context) (map[string]int64, error) {
	total, err := uc.repo.Count(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("count total tags: %w", err)
	}

	active, err := uc.repo.Count(ctx, "active")
	if err != nil {
		return nil, fmt.Errorf("count active tags: %w", err)
	}

	inactive, err := uc.repo.Count(ctx, "inactive")
	if err != nil {
		return nil, fmt.Errorf("count inactive tags: %w", err)
	}

	return map[string]int64{
		"total":    total,
		"active":   active,
		"inactive": inactive,
	}, nil
}
