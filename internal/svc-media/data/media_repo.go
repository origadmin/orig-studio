/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer for svc-media.
package data

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// mediaRepo implements the dto.MediaRepo interface using the shared entity.Client.
type mediaRepo struct {
	db *entity.Client
}

// NewMediaRepo creates a new Media repository.
func NewMediaRepo(db *entity.Client) dto.MediaRepo {
	return &mediaRepo{db: db}
}

func (r *mediaRepo) Get(
	ctx context.Context,
	id int64,
	opts ...*dto.MediaQueryOption,
) (*types.Media, error) {
	m, err := r.db.Media.Get(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return convertMediaToProto(m), nil
}

func (r *mediaRepo) List(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*types.Media, int32, error) {
	opt := &dto.MediaQueryOption{}
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	query := r.db.Media.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if opt.Page < 1 {
		opt.Page = 1
	}
	if opt.PageSize < 1 {
		opt.PageSize = 20
	}

	offset := (opt.Page - 1) * opt.PageSize
	items, err := query.Offset(int(offset)).
		Limit(int(opt.PageSize)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.Media, len(items))
	for i, item := range items {
		result[i] = convertMediaToProto(item)
	}
	return result, int32(total), nil
}

func (r *mediaRepo) Create(
	ctx context.Context,
	in *types.Media,
	opts ...*dto.MediaCreateOption,
) (*types.Media, error) {
	create := r.db.Media.Create().
		SetTitle(in.Title).
		SetURL(in.Url).
		SetType(in.Type)

	if in.Description != "" {
		create = create.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		create = create.SetThumbnail(in.Thumbnail)
	}
	if in.Duration > 0 {
		create = create.SetDuration(int(in.Duration))
	}
	if in.UserId > 0 {
		create = create.SetUserID(int(in.UserId))
	}
	if in.CategoryId > 0 {
		v := int(in.CategoryId)
		create = create.SetNillableCategoryID(&v)
	}

	m, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertMediaToProto(m), nil
}

func (r *mediaRepo) Update(
	ctx context.Context,
	in *types.Media,
	opts ...*dto.MediaUpdateOption,
) (*types.Media, error) {
	update := r.db.Media.UpdateOneID(int(in.Id)).
		SetTitle(in.Title)

	if in.Description != "" {
		update = update.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		update = update.SetThumbnail(in.Thumbnail)
	}

	m, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertMediaToProto(m), nil
}

func (r *mediaRepo) Delete(ctx context.Context, id int64) error {
	return r.db.Media.DeleteOneID(int(id)).Exec(ctx)
}

func (r *mediaRepo) ListCategories(
	ctx context.Context,
	opts ...*dto.CategoryQueryOption,
) ([]*types.Category, int32, error) {
	query := r.db.Category.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	items, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.Category, len(items))
	for i, item := range items {
		result[i] = convertCategoryToProto(item)
	}
	return result, int32(total), nil
}

func (r *mediaRepo) GetCategory(ctx context.Context, id int64) (*types.Category, error) {
	c, err := r.db.Category.Get(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return convertCategoryToProto(c), nil
}

func (r *mediaRepo) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	m, err := r.db.Media.UpdateOneID(int(id)).AddViewCount(1).Save(ctx)
	if err != nil {
		return 0, err
	}
	return m.ViewCount, nil
}

// convertMediaToProto converts entity.Media → proto types.Media.
func convertMediaToProto(m *entity.Media) *types.Media {
	return &types.Media{
		Id:          int64(m.ID),
		Title:       m.Title,
		Description: m.Description,
		Type:        m.Type,
		Url:         m.URL,
		Thumbnail:   m.Thumbnail,
		Duration:    int32(m.Duration),
		ViewCount:   m.ViewCount,
		LikeCount:   m.LikeCount,
		UserId:      int64(m.UserID),
		CreateTime:  timestamppb.New(m.CreatedAt),
	}
}

// convertCategoryToProto converts entity.Category → proto types.Category.
func convertCategoryToProto(c *entity.Category) *types.Category {
	return &types.Category{
		Id:          int64(c.ID),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		MediaCount:  int64(c.MediaCount),
	}
}
