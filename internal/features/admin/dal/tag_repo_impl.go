package dal

import (
	"context"
	"fmt"
	"strings"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/tag"
	"origadmin/application/origcms/internal/helpers/hashtag"
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
// B087-R2 Fix: Now applies search and status filters.
func (r *tagRepositoryImpl) List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*entity.Tag, int64, error) {
	query := r.client.Tag.Query()

	// B087-R2 Fix: Apply search filter (case-insensitive title match)
	if search != "" {
		query = query.Where(tag.TitleContainsFold(search))
	}

	// B087-R2 Fix: Apply status filter (normalize lowercase to uppercase enum)
	if status != "" {
		upperStatus := strings.ToUpper(status)
		switch upperStatus {
		case "ACTIVE", "INACTIVE":
			query = query.Where(tag.StatusEQ(tag.Status(upperStatus)))
		}
	}

	// Get total count (after filters)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count tags: %w", err)
	}

	// Apply sorting
	orderFunc := entity.Desc
	if sortOrder == "asc" {
		orderFunc = entity.Asc
	}

	sortField := tag.FieldCreateTime
	switch sortBy {
	case "title", "name":
		sortField = tag.FieldTitle
	case "media_count", "count":
		sortField = tag.FieldMediaCount
	case "update_time":
		sortField = tag.FieldUpdateTime
	default:
		sortField = tag.FieldCreateTime
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	tags, err := query.Offset(offset).Limit(pageSize).Order(orderFunc(sortField)).All(ctx)
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
// B087-R2 Fix: Now also saves Description, Color, and Status fields.
func (r *tagRepositoryImpl) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	builder := r.client.Tag.Create().
		SetTitle(tag.Title).
		SetMediaCount(0)

	if tag.Slug != "" {
		builder = builder.SetSlug(tag.Slug)
	}
	// B087-R2 Fix: Pass through optional fields
	if tag.Description != "" {
		builder = builder.SetDescription(tag.Description)
	}
	if tag.Color != "" {
		builder = builder.SetColor(tag.Color)
	}
	if tag.Status != "" {
		builder = builder.SetStatus(tag.Status)
	}

	createdTag, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return createdTag, nil
}

// Update updates an existing tag
func (r *tagRepositoryImpl) Update(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	builder := r.client.Tag.UpdateOne(tag).
		SetTitle(tag.Title).
		SetStatus(tag.Status)

	if tag.Slug != "" {
		builder = builder.SetSlug(tag.Slug)
	}
	if tag.Color != "" {
		builder = builder.SetColor(tag.Color)
	} else {
		builder = builder.ClearColor()
	}
	if tag.Description != "" {
		builder = builder.SetDescription(tag.Description)
	} else {
		builder = builder.ClearDescription()
	}

	updatedTag, err := builder.Save(ctx)
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

// GetByName returns a tag by title. It performs a case-insensitive lookup
// by querying the exact title first, then falling back to a case-insensitive
// search if the exact match fails.
func (r *tagRepositoryImpl) GetByName(ctx context.Context, name string) (*entity.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	// Try exact match first (most common case, uses index)
	t, err := r.client.Tag.Query().Where(tag.TitleEQ(name)).Only(ctx)
	if err == nil {
		return t, nil
	}

	// Fall back to case-insensitive search
	t, err = r.client.Tag.Query().Where(tag.TitleContainsFold(name)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tag by name: %w", err)
	}
	return t, nil
}

// GetOrCreateTag returns an existing tag by name (case-insensitive) or creates
// one if not found. When creating, it auto-generates a slug from the name.
// This method handles concurrent creation races by relying on the database
// unique constraint on title and retrying the lookup on create failure.
func (r *tagRepositoryImpl) GetOrCreateTag(ctx context.Context, name string) (*entity.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	// 1. Try to find existing tag
	existing, err := r.GetByName(ctx, name)
	if err == nil {
		return existing, nil
	}

	// 2. Not found, create new tag with auto-generated slug
	slug := hashtag.GenerateTagSlug(name)
	newTag := &entity.Tag{
		Title: name,
		Slug:  slug,
	}

	created, err := r.Create(ctx, newTag)
	if err != nil {
		// 3. Concurrent creation race: another request may have created it
		// Rely on unique constraint failure, then retry lookup
		existing, err2 := r.GetByName(ctx, name)
		if err2 != nil {
			return nil, fmt.Errorf("getOrCreate tag: create failed: %w, retry get failed: %v", err, err2)
		}
		return existing, nil
	}

	return created, nil
}

// BatchGetOrCreateTags returns existing tags or creates missing ones for the
// given names. Names are deduplicated (case-insensitive). Max 20 names per call.
func (r *tagRepositoryImpl) BatchGetOrCreateTags(ctx context.Context, names []string) ([]*entity.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	// Deduplicate (case-insensitive) and limit to MaxHashtags
	seen := make(map[string]bool)
	var uniqueNames []string
	for _, name := range names {
		if name == "" {
			continue
		}
		lower := strings.ToLower(name)
		if !seen[lower] {
			seen[lower] = true
			uniqueNames = append(uniqueNames, name)
		}
		if len(uniqueNames) >= hashtag.MaxHashtags {
			break
		}
	}

	var result []*entity.Tag
	for _, name := range uniqueNames {
		t, err := r.GetOrCreateTag(ctx, name)
		if err != nil {
			// Log and skip failed individual tags rather than failing the batch
			continue
		}
		result = append(result, t)
	}

	return result, nil
}
