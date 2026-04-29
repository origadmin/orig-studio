package dal

import (
	"context"
	"fmt"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/tag"
)

// TagStatus 定义tag状态类型
type TagStatus string

const (
	TagStatusActive   TagStatus = "active"
	TagStatusInactive TagStatus = "inactive"
)

// tagRepositoryImpl implements TagRepository using Ent ORM
type tagRepositoryImpl struct {
	client *entity.Client
}

// NewTagRepository creates a new TagRepository
func NewTagRepository(client *entity.Client) TagRepository {
	return &tagRepositoryImpl{client: client}
}

// List returns a paginated list of tags
func (r *tagRepositoryImpl) List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*entity.Tag, int64, error) {
	query := r.client.Tag.Query()

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count tags: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	tags, err := query.Offset(offset).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}

	return tags, int64(total), nil
}

// Get returns a tag by ID
func (r *tagRepositoryImpl) Get(ctx context.Context, id string) (*entity.Tag, error) {
	tagID := 0
	if _, err := fmt.Sscanf(id, "%d", &tagID); err != nil {
		return nil, fmt.Errorf("invalid tag ID: %w", err)
	}
	
	tag, err := r.client.Tag.Get(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}
	return tag, nil
}

// Create creates a new tag
func (r *tagRepositoryImpl) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	createdTag, err := r.client.Tag.Create().
		SetTitle(tag.Title).
		SetMediaCount(0).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return createdTag, nil
}

// Update updates an existing tag
func (r *tagRepositoryImpl) Update(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	updatedTag, err := r.client.Tag.UpdateOne(tag).
		SetTitle(tag.Title).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}
	return updatedTag, nil
}

// Delete deletes a tag by ID
func (r *tagRepositoryImpl) Delete(ctx context.Context, id string) error {
	tagID := 0
	if _, err := fmt.Sscanf(id, "%d", &tagID); err != nil {
		return fmt.Errorf("invalid tag ID: %w", err)
	}
	
	if err := r.client.Tag.DeleteOneID(tagID).Exec(ctx); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}

// BulkDelete deletes multiple tags by IDs
func (r *tagRepositoryImpl) BulkDelete(ctx context.Context, ids []string) (int, error) {
	count := 0
	for _, id := range ids {
		tagID := 0
		if _, err := fmt.Sscanf(id, "%d", &tagID); err != nil {
			continue
		}
		if err := r.client.Tag.DeleteOneID(tagID).Exec(ctx); err == nil {
			count++
		}
	}
	return count, nil
}

// GetBySlug returns a tag by slug
func (r *tagRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entity.Tag, error) {
	t, err := r.client.Tag.Query().Where(tag.SlugEQ(slug)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tag by slug: %w", err)
	}
	return t, nil
}

// Count returns the total number of tags
func (r *tagRepositoryImpl) Count(ctx context.Context, status string) (int64, error) {
	count, err := r.client.Tag.Query().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count tags: %w", err)
	}
	return int64(count), nil
}
