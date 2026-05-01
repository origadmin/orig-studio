/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/category"
	"origadmin/application/origcms/internal/data/entity/tag"
	"origadmin/application/origcms/internal/features/content/biz"
)

type categoryRepo struct {
	data *Data
	log  *log.Helper
}

type tagRepo struct {
	data *Data
	log  *log.Helper
	cacheMutex sync.RWMutex
	tagCache   map[int]*biz.Tag
	tagByName  map[string]*biz.Tag
	tagBySlug  map[string]*biz.Tag
}

func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{data: data, log: log.NewHelper(log.With(logger, "module", "category.data"))}
}

func NewTagRepo(data *Data, logger log.Logger) biz.TagRepo {
	return &tagRepo{
		data:       data,
		log:        log.NewHelper(log.With(logger, "module", "tag.data")),
		tagCache:   make(map[int]*biz.Tag),
		tagByName:  make(map[string]*biz.Tag),
		tagBySlug:  make(map[string]*biz.Tag),
	}
}

func (r *categoryRepo) Create(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	ent, err := r.data.db.Category.Create().
		SetName(c.Name).
		SetSlug(c.Slug).
		SetDescription(c.Description).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapCategory(ent), nil
}

func (r *categoryRepo) Get(ctx context.Context, id int) (*biz.Category, error) {
	ent, err := r.data.db.Category.Query().Where(category.ID(int64(id))).First(ctx)
	if err != nil {
		return nil, err
	}
	return mapCategory(ent), nil
}

func (r *categoryRepo) Update(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	ent, err := r.data.db.Category.UpdateOneID(int64(c.ID)).
		SetName(c.Name).
		SetSlug(c.Slug).
		SetDescription(c.Description).
		SetStatus(categoryStatusFromInt(c.Status)).
		SetParentID(c.ParentID).
		SetSequence(c.Sequence).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapCategory(ent), nil
}

func (r *categoryRepo) Delete(ctx context.Context, id int) error {
	return r.data.db.Category.DeleteOneID(int64(id)).Exec(ctx)
}

func (r *categoryRepo) ListAll(ctx context.Context) ([]*biz.Category, error) {
	ents, err := r.data.db.Category.Query().Order(entity.Desc(category.FieldCreateTime)).All(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*biz.Category, len(ents))
	for i, ent := range ents {
		res[i] = mapCategory(ent)
	}
	return res, nil
}

func (r *tagRepo) Create(ctx context.Context, t *biz.Tag) (*biz.Tag, error) {
	builder := r.data.db.Tag.Create().
		SetTitle(t.Title)
	if t.Slug != "" {
		builder = builder.SetSlug(t.Slug)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)
	
	// Add to cache
	r.cacheMutex.Lock()
	r.tagCache[tag.ID] = tag
	r.tagByName[tag.Title] = tag
	if tag.Slug != "" {
		r.tagBySlug[tag.Slug] = tag
	}
	r.cacheMutex.Unlock()
	
	return tag, nil
}

func (r *tagRepo) Get(ctx context.Context, id int) (*biz.Tag, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if tag, ok := r.tagCache[id]; ok {
		r.cacheMutex.RUnlock()
		return tag, nil
	}
	r.cacheMutex.RUnlock()
	
	// Cache miss, query database
	ent, err := r.data.db.Tag.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)
	
	// Add to cache
	r.cacheMutex.Lock()
	r.tagCache[id] = tag
	r.tagByName[tag.Title] = tag
	r.cacheMutex.Unlock()
	
	return tag, nil
}

func (r *tagRepo) GetByName(ctx context.Context, name string) (*biz.Tag, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if tag, ok := r.tagByName[name]; ok {
		r.cacheMutex.RUnlock()
		return tag, nil
	}
	r.cacheMutex.RUnlock()
	
	// Cache miss, query database
	ent, err := r.data.db.Tag.Query().Where(tag.TitleEQ(name)).Only(ctx)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)
	
	// Add to cache
	r.cacheMutex.Lock()
	r.tagCache[tag.ID] = tag
	r.tagByName[name] = tag
	if tag.Slug != "" {
		r.tagBySlug[tag.Slug] = tag
	}
	r.cacheMutex.Unlock()
	
	return tag, nil
}

func (r *tagRepo) GetBySlug(ctx context.Context, slug string) (*biz.Tag, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if tag, ok := r.tagBySlug[slug]; ok {
		r.cacheMutex.RUnlock()
		return tag, nil
	}
	r.cacheMutex.RUnlock()

	// Cache miss, query database
	ent, err := r.data.db.Tag.Query().Where(tag.SlugEQ(slug)).Only(ctx)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)

	// Add to cache
	r.cacheMutex.Lock()
	r.tagCache[tag.ID] = tag
	r.tagByName[tag.Title] = tag
	if tag.Slug != "" {
		r.tagBySlug[tag.Slug] = tag
	}
	r.cacheMutex.Unlock()

	return tag, nil
}

func (r *tagRepo) Update(ctx context.Context, t *biz.Tag) (*biz.Tag, error) {
	builder := r.data.db.Tag.UpdateOneID(t.ID).
		SetTitle(t.Title)
	if t.Slug != "" {
		builder = builder.SetSlug(t.Slug)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)
	
	// Update cache
	r.cacheMutex.Lock()
	// Remove old entry from tagByName and tagBySlug
	if oldTag, ok := r.tagCache[tag.ID]; ok {
		delete(r.tagByName, oldTag.Title)
		delete(r.tagBySlug, oldTag.Slug)
	}
	// Add updated tag to cache
	r.tagCache[tag.ID] = tag
	r.tagByName[tag.Title] = tag
	if tag.Slug != "" {
		r.tagBySlug[tag.Slug] = tag
	}
	r.cacheMutex.Unlock()
	
	return tag, nil
}

func (r *tagRepo) Delete(ctx context.Context, id int) error {
	// Get tag before deleting to remove from cache
	tag, err := r.Get(ctx, id)
	if err != nil {
		// Tag not found, just proceed with deletion
	}
	
	// Delete from database
	if err := r.data.db.Tag.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}
	
	// Remove from cache
	if tag != nil {
		r.cacheMutex.Lock()
		delete(r.tagCache, id)
		delete(r.tagByName, tag.Title)
		delete(r.tagBySlug, tag.Slug)
		r.cacheMutex.Unlock()
	}
	
	return nil
}

func (r *tagRepo) ListAll(ctx context.Context, page, pageSize int) ([]*biz.Tag, int, error) {
	query := r.data.db.Tag.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	ents, err := query.
		Order(entity.Desc(tag.FieldCreateTime)).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*biz.Tag, len(ents))
	for i, ent := range ents {
		res[i] = mapTag(ent)
	}
	return res, total, nil
}

func mapCategory(ent *entity.Category) *biz.Category {
	return &biz.Category{
		ID:          int(ent.ID),
		Name:        ent.Name,
		Slug:        ent.Slug,
		Description: ent.Description,
		Status:      categoryStatusToInt(ent.Status),
		ParentID:    ent.ParentID,
		Sequence:    ent.Sequence,
		MediaCount:  ent.MediaCount,
		CreateTime:   ent.CreateTime,
		UpdateTime:   ent.UpdateTime,
	}
}

func mapTag(ent *entity.Tag) *biz.Tag {
	return &biz.Tag{
		ID:         ent.ID,
		Title:      ent.Title,
		Slug:       ent.Slug,
		MediaCount: ent.MediaCount,
	}
}

// categoryStatusFromInt converts int (biz layer) to category.Status (entity enum).
// Frontend convention: 1 = active/enabled, 0 = inactive/disabled.
// Ent enum convention: ACTIVE = 1, INACTIVE = 2.
// Both 0 and 2 map to INACTIVE; default falls to INACTIVE for safety.
func categoryStatusFromInt(status int) category.Status {
	switch status {
	case 1:
		return category.StatusACTIVE
	case 2, 0: // 2 = INACTIVE (Ent enum), 0 = inactive (frontend convention)
		return category.StatusINACTIVE
	default:
		return category.StatusINACTIVE // safe default: unknown status treated as inactive
	}
}

// categoryStatusToInt converts category.Status (entity enum) to int (biz layer).
func categoryStatusToInt(status category.Status) int {
	switch status {
	case category.StatusACTIVE:
		return 1
	case category.StatusINACTIVE:
		return 2
	default:
		return 1
	}
}
