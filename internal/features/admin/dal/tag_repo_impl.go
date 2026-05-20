package dal

import (
	"context"
	"fmt"
	"strings"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/tag"
	"origadmin/application/origstudio/internal/features/admin/dto"
	"origadmin/application/origstudio/internal/pkg/hashtag"
)

// EntityToTagDTO converts an entity.Tag to a dto.TagDTO.
func EntityToTagDTO(t *entity.Tag) *dto.TagDTO {
	if t == nil {
		return nil
	}
	return &dto.TagDTO{
		ID:                t.ID,
		Title:             t.Title,
		Slug:              t.Slug,
		MediaCount:        t.MediaCount,
		ChannelCount:      t.ChannelCount,
		ListingsThumbnail: t.ListingsThumbnail,
		Status:            dto.TagStatusType(string(t.Status)),
		Description:       t.Description,
		TitleI18n:         t.TitleI18n,
		DescriptionI18n:   t.DescriptionI18n,
		Color:             t.Color,
		CreateTime:        t.CreateTime,
		UpdateTime:        t.UpdateTime,
	}
}

// EntityToTagDTOList converts a slice of entity.Tag to dto.TagDTO slice.
func EntityToTagDTOList(tags []*entity.Tag) []*dto.TagDTO {
	if len(tags) == 0 {
		return []*dto.TagDTO{}
	}
	result := make([]*dto.TagDTO, len(tags))
	for i, t := range tags {
		result[i] = EntityToTagDTO(t)
	}
	return result
}

// tagRepositoryImpl implements TagRepository using Ent ORM
type tagRepositoryImpl struct {
	client *entity.Client
}

// NewTagRepository creates a new TagRepository
func NewTagRepository(client *entity.Client) TagRepository {
	return &tagRepositoryImpl{client: client}
}

// List returns a paginated list of tags
func (r *tagRepositoryImpl) List(ctx context.Context, page, pageSize int, search, status, sortBy, sortOrder string) ([]*dto.TagDTO, int64, error) {
	query := r.client.Tag.Query()

	if search != "" {
		query = query.Where(tag.TitleContainsFold(search))
	}

	if status != "" {
		upperStatus := strings.ToUpper(status)
		switch upperStatus {
		case "ACTIVE", "INACTIVE":
			query = query.Where(tag.StatusEQ(tag.Status(upperStatus)))
		}
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count tags: %w", err)
	}

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

	offset := (page - 1) * pageSize
	tags, err := query.Offset(offset).Limit(pageSize).Order(orderFunc(sortField)).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}

	return EntityToTagDTOList(tags), int64(total), nil
}

// Get returns a tag by ID
func (r *tagRepositoryImpl) Get(ctx context.Context, id string) (*dto.TagDTO, error) {
	tagID := 0
	if _, err := fmt.Sscanf(id, "%d", &tagID); err != nil {
		return nil, fmt.Errorf("invalid tag ID: %w", err)
	}

	t, err := r.client.Tag.Get(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}
	return EntityToTagDTO(t), nil
}

// Create creates a new tag
func (r *tagRepositoryImpl) Create(ctx context.Context, tagDTO *dto.TagDTO) (*dto.TagDTO, error) {
	builder := r.client.Tag.Create().
		SetTitle(tagDTO.Title).
		SetMediaCount(0)

	if tagDTO.Slug != "" {
		builder = builder.SetSlug(tagDTO.Slug)
	}
	if tagDTO.Description != "" {
		builder = builder.SetDescription(tagDTO.Description)
	}
	if tagDTO.Color != "" {
		builder = builder.SetColor(tagDTO.Color)
	}
	if tagDTO.Status != "" {
		builder = builder.SetStatus(tag.Status(string(tagDTO.Status)))
	}

	createdTag, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return EntityToTagDTO(createdTag), nil
}

// Update updates an existing tag
func (r *tagRepositoryImpl) Update(ctx context.Context, tagDTO *dto.TagDTO) (*dto.TagDTO, error) {
	// Fetch the entity first to get the ent reference for update
	tagID := tagDTO.ID
	entityTag, err := r.client.Tag.Get(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("get tag for update: %w", err)
	}

	builder := r.client.Tag.UpdateOne(entityTag).
		SetTitle(tagDTO.Title).
		SetStatus(tag.Status(string(tagDTO.Status)))

	if tagDTO.Slug != "" {
		builder = builder.SetSlug(tagDTO.Slug)
	}
	if tagDTO.Color != "" {
		builder = builder.SetColor(tagDTO.Color)
	} else {
		builder = builder.ClearColor()
	}
	if tagDTO.Description != "" {
		builder = builder.SetDescription(tagDTO.Description)
	} else {
		builder = builder.ClearDescription()
	}

	updatedTag, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}
	return EntityToTagDTO(updatedTag), nil
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
func (r *tagRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*dto.TagDTO, error) {
	t, err := r.client.Tag.Query().Where(tag.SlugEQ(slug)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tag by slug: %w", err)
	}
	return EntityToTagDTO(t), nil
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
func (r *tagRepositoryImpl) GetByName(ctx context.Context, name string) (*dto.TagDTO, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	// Try exact match first (most common case, uses index)
	t, err := r.client.Tag.Query().Where(tag.TitleEQ(name)).Only(ctx)
	if err == nil {
		return EntityToTagDTO(t), nil
	}

	// Fall back to case-insensitive search
	t, err = r.client.Tag.Query().Where(tag.TitleContainsFold(name)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tag by name: %w", err)
	}
	return EntityToTagDTO(t), nil
}

// GetOrCreateTag returns an existing tag by name (case-insensitive) or creates
// one if not found. When creating, it auto-generates a slug from the name.
func (r *tagRepositoryImpl) GetOrCreateTag(ctx context.Context, name string) (*dto.TagDTO, error) {
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
	newTag := &dto.TagDTO{
		Title: name,
		Slug:  slug,
	}

	created, err := r.Create(ctx, newTag)
	if err != nil {
		// 3. Concurrent creation race: another request may have created it
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
func (r *tagRepositoryImpl) BatchGetOrCreateTags(ctx context.Context, names []string) ([]*dto.TagDTO, error) {
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

	var result []*dto.TagDTO
	for _, name := range uniqueNames {
		t, err := r.GetOrCreateTag(ctx, name)
		if err != nil {
			continue
		}
		result = append(result, t)
	}

	return result, nil
}
