package service

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/features/admin/biz"
)

// TagService handles tag service operations
type TagService struct {
	usecase *biz.TagUseCase
}

// NewTagService creates a new TagService
func NewTagService(usecase *biz.TagUseCase) *TagService {
	return &TagService{usecase: usecase}
}

// List returns a paginated list of tags
func (s *TagService) List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*entity.Tag, int64, error) {
	tags, total, err := s.usecase.List(ctx, page, pageSize, search, status, sortBy, sortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}

	return tags, total, nil
}

// Get returns a tag by ID
func (s *TagService) Get(ctx context.Context, id string) (*entity.Tag, error) {
	tag, err := s.usecase.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}

	return tag, nil
}

// Create creates a new tag
func (s *TagService) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	createdTag, err := s.usecase.Create(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	return createdTag, nil
}

// Update updates an existing tag
func (s *TagService) Update(ctx context.Context, id string, updates *entity.Tag) (*entity.Tag, error) {
	updatedTag, err := s.usecase.Update(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}

	return updatedTag, nil
}

// Delete deletes a tag by ID
func (s *TagService) Delete(ctx context.Context, id string) error {
	if err := s.usecase.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	return nil
}

// BulkDelete deletes multiple tags by IDs
func (s *TagService) BulkDelete(ctx context.Context, ids []string) (int, error) {
	count, err := s.usecase.BulkDelete(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("bulk delete tags: %w", err)
	}

	return count, nil
}

// Count returns tag counts
func (s *TagService) Count(ctx context.Context) (map[string]int64, error) {
	counts, err := s.usecase.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count tags: %w", err)
	}

	return counts, nil
}
