/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"
	"fmt"
	"sync"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/category"
	"origadmin/application/origcms/internal/data/entity/tag"
	"origadmin/application/origcms/internal/svc-content/biz"
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
	// 由于实体的ID是string类型，我们使用字符串格式查询
	ent, err := r.data.db.Category.Query().Where(category.ID(fmt.Sprintf("%d", id))).First(ctx)
	if err != nil {
		return nil, err
	}
	return mapCategory(ent), nil
}

func (r *categoryRepo) Update(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	ent, err := r.data.db.Category.UpdateOneID(fmt.Sprintf("%d", c.ID)).
		SetName(c.Name).
		SetSlug(c.Slug).
		SetDescription(c.Description).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapCategory(ent), nil
}

func (r *categoryRepo) Delete(ctx context.Context, id int) error {
	return r.data.db.Category.DeleteOneID(fmt.Sprintf("%d", id)).Exec(ctx)
}

func (r *categoryRepo) ListAll(ctx context.Context) ([]*biz.Category, error) {
	ents, err := r.data.db.Category.Query().All(ctx)
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
	ent, err := r.data.db.Tag.Create().
		SetTitle(t.Title).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)
	
	// Add to cache
	r.cacheMutex.Lock()
	r.tagCache[tag.ID] = tag
	r.tagByName[tag.Title] = tag
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
	r.cacheMutex.Unlock()
	
	return tag, nil
}

func (r *tagRepo) Update(ctx context.Context, t *biz.Tag) (*biz.Tag, error) {
	ent, err := r.data.db.Tag.UpdateOneID(t.ID).
		SetTitle(t.Title).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	tag := mapTag(ent)
	
	// Update cache
	r.cacheMutex.Lock()
	// Remove old entry from tagByName
	if oldTag, ok := r.tagCache[tag.ID]; ok {
		delete(r.tagByName, oldTag.Title)
	}
	// Add updated tag to cache
	r.tagCache[tag.ID] = tag
	r.tagByName[tag.Title] = tag
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
	// 由于实体的ID是string类型，而业务模型使用int类型
	// 尝试将字符串ID转换为整数
	id := 0
	if ent.ID != "" {
		// 尝试将字符串ID解析为整数
		// 注意：在实际生产环境中，可能需要根据ID生成策略调整
		if parsedID, err := strconv.Atoi(ent.ID); err == nil {
			id = parsedID
		} else {
			// 如果转换失败，使用一个默认值或基于字符串的哈希值
			id = len(ent.ID)
		}
	}
	return &biz.Category{
		ID:          id,
		Name:        ent.Name,
		Slug:        ent.Slug,
		Description: ent.Description,
		CreatedAt:   ent.CreatedAt,
		UpdatedAt:   ent.UpdatedAt,
	}
}

func mapTag(ent *entity.Tag) *biz.Tag {
	return &biz.Tag{
		ID:         ent.ID,
		Title:      ent.Title,
		MediaCount: ent.MediaCount,
	}
}
