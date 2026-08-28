/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"sync"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/category"
	"origadmin/application/origstudio/internal/data/entity/media"
	"origadmin/application/origstudio/internal/data/entity/tag"
	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
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

func (r *categoryRepo) GetBySlug(ctx context.Context, slug string) (*biz.Category, error) {
	ent, err := r.data.db.Category.Query().Where(category.SlugEQ(slug)).Only(ctx)
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

func (r *categoryRepo) ListActive(ctx context.Context) ([]*biz.Category, error) {
	ents, err := r.data.db.Category.Query().
		Where(category.StatusEQ(category.StatusACTIVE)).
		Order(entity.Asc(category.FieldSequence), entity.Desc(category.FieldCreateTime)).
		All(ctx)
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
	// BUG-180 (visibility): tags created through the public TagService
	// (POST /api/v1/tags, the only write path that survives the gateway) arrive
	// without an explicit status, so they defaulted to UNSPECIFIED and were
	// hidden by any ACTIVE filter / sorted to the bottom of the list. Default
	// new tags to ACTIVE so they show up immediately.
	builder := r.data.db.Tag.Create().
		SetTitle(t.Title).
		SetStatus(tag.StatusACTIVE)
	if t.Slug != "" {
		builder = builder.SetSlug(t.Slug)
	}
	if t.Description != "" {
		builder = builder.SetDescription(t.Description)
	}
	if t.Color != "" {
		builder = builder.SetColor(t.Color)
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

// withMediaCount returns a shallow copy of t carrying the live media count.
//
// BUG-180: the cached *biz.Tag holds the dead content_tags.media_count column.
// Copying keeps the shared cache entry immutable (no data race between
// concurrent readers) while still handing callers an accurate count derived
// from content_media.tags (the same source the tag detail page reads).
func (r *tagRepo) withMediaCount(ctx context.Context, t *biz.Tag) *biz.Tag {
	if t == nil {
		return nil
	}
	out := *t
	out.MediaCount = r.countMediaByTag(ctx, t.Title)
	return &out
}

func (r *tagRepo) Get(ctx context.Context, id int) (*biz.Tag, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if tag, ok := r.tagCache[id]; ok {
		r.cacheMutex.RUnlock()
		return r.withMediaCount(ctx, tag), nil
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
	
	return r.withMediaCount(ctx, tag), nil
}

func (r *tagRepo) GetByName(ctx context.Context, name string) (*biz.Tag, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if tag, ok := r.tagByName[name]; ok {
		r.cacheMutex.RUnlock()
		return r.withMediaCount(ctx, tag), nil
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
	
	return r.withMediaCount(ctx, tag), nil
}

func (r *tagRepo) GetBySlug(ctx context.Context, slug string) (*biz.Tag, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if tag, ok := r.tagBySlug[slug]; ok {
		r.cacheMutex.RUnlock()
		return r.withMediaCount(ctx, tag), nil
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

	return r.withMediaCount(ctx, tag), nil
}

func (r *tagRepo) Update(ctx context.Context, t *biz.Tag) (*biz.Tag, error) {
	builder := r.data.db.Tag.UpdateOneID(t.ID).
		SetTitle(t.Title)
	if t.Slug != "" {
		builder = builder.SetSlug(t.Slug)
	}
	if t.Description != "" {
		builder = builder.SetDescription(t.Description)
	}
	if t.Color != "" {
		builder = builder.SetColor(t.Color)
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
		// BUG-180: seed tags were inserted with a NULL create_time, and
		// `ORDER BY create_time DESC` puts NULLs FIRST — so freshly created
		// tags (which DO carry a timestamp) sank to the last page and looked
		// "invisible". NULLS LAST keeps newest tags at the top of "latest".
		// The secondary id DESC tie-break keeps cross-page ordering stable
		// when many tags share the same create_time.
		Order(
			sql.OrderByField(tag.FieldCreateTime, sql.OrderDesc(), sql.OrderNullsLast()).ToFunc(),
			sql.OrderByField(tag.FieldID, sql.OrderDesc()).ToFunc(),
		).
		Limit(pageSize).
		Offset(types.CalcOffset(page, pageSize)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*biz.Tag, len(ents))
	for i, ent := range ents {
		res[i] = mapTag(ent)
	}
	r.fillMediaCounts(ctx, res)
	return res, total, nil
}

// fillMediaCounts replaces the denormalized content_tags.media_count value with
// the live count derived from content_media.tags.
//
// BUG-180 (same class as BUG-128 on playlists): the previous implementation
// counted rows in the content_media_tags pivot, which no write path kept in
// sync, so the portal reported "0 videos" for almost every tag while the tag
// detail page (which filters content_media.tags directly) showed the real
// videos. Both now read the SAME source — content_media.tags — so a tag's
// listed count always matches what its detail page actually displays.
func (r *tagRepo) fillMediaCounts(ctx context.Context, tags []*biz.Tag) {
	for _, t := range tags {
		if t != nil {
			t.MediaCount = r.countMediaByTag(ctx, t.Title)
		}
	}
}

// countMediaByTag returns the number of media whose `tags` jsonb array contains
// the given tag title. This mirrors exactly the predicate the public media list
// uses for the ?tags= filter (sqljson.ValueContains(media.FieldTags, tag)), so
// the count shown in tag lists is always identical to the tag detail page.
//
// The previous pivot-based count (content_media_tags) was stale for most tags;
// content_media.tags is the authoritative, live relationship.
func (r *tagRepo) countMediaByTag(ctx context.Context, title string) int {
	if title == "" {
		return 0
	}
	count, err := r.data.db.Media.Query().
		Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueContains(media.FieldTags, title))
		}).
		Count(ctx)
	if err != nil {
		r.log.Warnf("tag %q: count media by tags jsonb failed: %v", title, err)
		return 0
	}
	return count
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
		ID:          ent.ID,
		Title:       ent.Title,
		Slug:        ent.Slug,
		Description: ent.Description,
		Color:       ent.Color,
		Status:      tagStatusToInt(ent.Status),
		MediaCount:  ent.MediaCount,
		CreateTime:  ent.CreateTime,
	}
}

// tagStatusToInt maps the ent string enum ("ACTIVE"/"INACTIVE") to the proto
// TagStatus int values (0=UNSPECIFIED, 1=INACTIVE, 2=ACTIVE).
func tagStatusToInt(status tag.Status) int {
	switch status {
	case tag.StatusACTIVE:
		return 2
	case tag.StatusINACTIVE:
		return 1
	default:
		return 0
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
