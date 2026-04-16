package data

import (
	"context"

	"origadmin/application/origcms/internal/data/entity"
)

// TagRepository defines the repository interface for tags
type TagRepository interface {
	// List returns a paginated list of tags
	List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*entity.Tag, int64, error)

	// Get returns a tag by ID
	Get(ctx context.Context, id string) (*entity.Tag, error)

	// Create creates a new tag
	Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)

	// Update updates an existing tag
	Update(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)

	// Delete deletes a tag by ID
	Delete(ctx context.Context, id string) error

	// BulkDelete deletes multiple tags by IDs
	BulkDelete(ctx context.Context, ids []string) (int, error)

	// GetBySlug returns a tag by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Tag, error)

	// Count returns the total number of tags
	Count(ctx context.Context, status string) (int64, error)
}
