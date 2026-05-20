package dal

import (
	"context"

	"origadmin/application/origstudio/internal/features/admin/dto"
)

// TagRepository defines the repository interface for tags
type TagRepository interface {
	// List returns a paginated list of tags
	List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*dto.TagDTO, int64, error)

	// Get returns a tag by ID
	Get(ctx context.Context, id string) (*dto.TagDTO, error)

	// Create creates a new tag
	Create(ctx context.Context, tag *dto.TagDTO) (*dto.TagDTO, error)

	// Update updates an existing tag
	Update(ctx context.Context, tag *dto.TagDTO) (*dto.TagDTO, error)

	// Delete deletes a tag by ID
	Delete(ctx context.Context, id string) error

	// BulkDelete deletes multiple tags by IDs
	BulkDelete(ctx context.Context, ids []string) (int, error)

	// GetBySlug returns a tag by slug
	GetBySlug(ctx context.Context, slug string) (*dto.TagDTO, error)

	// GetByName returns a tag by title (case-insensitive match)
	GetByName(ctx context.Context, name string) (*dto.TagDTO, error)

	// Count returns the total number of tags
	Count(ctx context.Context, status string) (int64, error)

	// GetOrCreateTag returns an existing tag by name (case-insensitive) or creates one if not found.
	// When creating, it auto-generates a slug from the name.
	GetOrCreateTag(ctx context.Context, name string) (*dto.TagDTO, error)

	// BatchGetOrCreateTags returns existing tags or creates missing ones for the given names.
	// Names are deduplicated (case-insensitive). Max 20 names per call.
	BatchGetOrCreateTags(ctx context.Context, names []string) ([]*dto.TagDTO, error)
}
