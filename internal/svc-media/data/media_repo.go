/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer for svc-media.
package data

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/category"
	"origadmin/application/origcms/internal/data/entity/encodingtask"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// mediaRepo implements the biz.MediaRepo interface using the shared entity.Client.
type mediaRepo struct {
	db *entity.Client
}

// NewMediaRepo creates a new Media repository.
func NewMediaRepo(db *entity.Client) biz.MediaRepo {
	return &mediaRepo{db: db}
}

func (r *mediaRepo) Get(
	ctx context.Context,
	id int64,
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

	// Apply filters
	if opt.UserID != nil {
		query = query.Where(media.UserIDEQ(int(*opt.UserID)))
	}
	if opt.CategoryID != nil {
		query = query.Where(media.HasCategoryWith(category.ID(int(*opt.CategoryID))))
	}
	if opt.Status != nil {
		// Map integer status to string state if needed, or use directly if it's already a string in your business logic.
		// Assuming status 1 = active for now, or just converting to string.
		state := fmt.Sprintf("%d", *opt.Status)
		query = query.Where(media.StateEQ(state))
	}
	if opt.Keyword != "" {
		query = query.Where(media.TitleContains(opt.Keyword))
	}

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
) (*types.Media, error) {
	create := r.db.Media.Create().
		SetTitle(in.Title).
		SetURL(in.Url).
		SetType(in.Type).
		SetMimeType(in.MimeType).
		SetSize(fmt.Sprintf("%d", in.Size))

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
) (*types.Media, error) {
	update := r.db.Media.UpdateOneID(int(in.Id)).
		SetTitle(in.Title).
		SetMimeType(in.MimeType).
		SetSize(fmt.Sprintf("%d", in.Size))

	if in.Description != "" {
		update = update.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		update = update.SetThumbnail(in.Thumbnail)
	}
	if in.HlsFile != "" {
		update = update.SetHlsFile(in.HlsFile)
	}
	if in.EncodingStatus != "" {
		update = update.SetEncodingStatus(in.EncodingStatus)
	}
	if in.Uuid != "" {
		update = update.SetUUID(in.Uuid)
	}
	if in.Duration > 0 {
		update = update.SetDuration(int(in.Duration))
	}
	if in.CategoryId > 0 {
		v := int(in.CategoryId)
		update = update.SetNillableCategoryID(&v)
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

// CountByEncodingStatus returns per-status media counts using a single GROUP BY query.
func (r *mediaRepo) CountByEncodingStatus(ctx context.Context) (*biz.StatusCounts, error) {
	type countRow struct {
		EncodingStatus string `json:"encoding_status"`
		Count          int    `json:"count"`
	}

	var rows []countRow
	err := r.db.Media.Query().
		GroupBy(media.FieldEncodingStatus).
		Aggregate(entity.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	counts := &biz.StatusCounts{}
	for _, row := range rows {
		switch row.EncodingStatus {
		case "processing":
			counts.Processing = row.Count
		case "pending":
			counts.Pending = row.Count
		case "partial":
			counts.Partial = row.Count
		case "failed":
			counts.Failed = row.Count
		case "success":
			counts.Success = row.Count
		}
	}
	return counts, nil
}

// ListFilteredByEncodingStatus returns a paginated list of media filtered by encoding status.
func (r *mediaRepo) ListFilteredByEncodingStatus(ctx context.Context, statuses []string, page, pageSize int) ([]*types.Media, int, error) {
	if len(statuses) == 0 {
		return nil, 0, nil
	}

	query := r.db.Media.Query().
		Where(media.EncodingStatusIn(statuses...)).
		Order(entity.Desc(media.FieldUpdatedAt))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	items, err := query.
		Limit(pageSize).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.Media, len(items))
	for i, item := range items {
		result[i] = convertMediaToProto(item)
	}
	return result, total, nil
}

// ResetStaleProcessing resets media stuck in "processing" back to "pending"
// and marks their associated encoding tasks still in "processing" as "failed".
// Returns the count of reset media items.
func (r *mediaRepo) ResetStaleProcessing(ctx context.Context) (int, error) {
	// 1. Find all media with encoding_status = "processing"
	staleMedia, err := r.db.Media.Query().
		Where(media.EncodingStatusEQ("processing")).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query stale processing media: %w", err)
	}

	if len(staleMedia) == 0 {
		return 0, nil
	}

	// 2. Delete orphaned encoding tasks still in "processing" — they were
	// interrupted by the restart and will be recreated when the media is re-processed.
	for _, m := range staleMedia {
		_, err := r.db.EncodingTask.Delete().
			Where(
				encodingtask.MediaIDEQ(m.ID),
				encodingtask.StatusEQ("processing"),
			).
			Exec(ctx)
		if err != nil {
			return 0, fmt.Errorf("delete orphaned tasks for media %d: %w", m.ID, err)
		}
	}

	// 3. Reset all stale media to "pending"
	count, err := r.db.Media.Update().
		Where(media.EncodingStatusEQ("processing")).
		SetEncodingStatus("pending").
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("reset stale media status: %w", err)
	}

	return count, nil
}

// convertMediaToProto converts entity.Media → proto types.Media.
func convertMediaToProto(m *entity.Media) *types.Media {
	var size int64
	_, _ = fmt.Sscanf(m.Size, "%d", &size)

	return &types.Media{
		Id:             int64(m.ID),
		Title:          m.Title,
		Description:    m.Description,
		Type:           m.Type,
		Url:            m.URL,
		Thumbnail:      m.Thumbnail,
		HlsFile:        m.HlsFile,
		EncodingStatus: m.EncodingStatus,
		Uuid:           m.UUID,
		Duration:       int32(m.Duration),
		Size:           size,
		MimeType:       m.MimeType,
		ViewCount:      m.ViewCount,
		LikeCount:      m.LikeCount,
		UserId:         int64(m.UserID),
		CreateTime:     timestamppb.New(m.CreatedAt),
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
