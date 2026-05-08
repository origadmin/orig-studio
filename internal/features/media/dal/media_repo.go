/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer for svc-media.
package dal

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/convpb"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/category"
	"origadmin/application/origcms/internal/data/entity/encodingtask"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dto"
)

// mediaRepo implements the biz.MediaRepo interface using the shared entity.Client.
type mediaRepo struct {
	db *entity.Client
}

// NewMediaRepo creates a new Media repository.
func NewMediaRepo(db *entity.Client) biz.MediaRepo {
	return &mediaRepo{db: db}
}

// GetByShortToken 通过 short_token 获取媒体（仅用于公开 API）
func (r *mediaRepo) GetByShortToken(ctx context.Context, shortToken string) (*types.Media, error) {
	if strings.TrimSpace(shortToken) == "" {
		return nil, fmt.Errorf("short_token cannot be empty")
	}
	m, err := r.db.Media.Query().
		Where(media.ShortTokenEQ(shortToken)).
		WithUser().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("media not found by short_token %s: %w", shortToken, err)
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

// GetByID 通过 UUID 获取媒体（仅用于 Admin/Authenticated API）
func (r *mediaRepo) GetByID(ctx context.Context, id string) (*types.Media, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	m, err := r.db.Media.Query().
		Where(media.IDEQ(id)).
		WithUser().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("media not found by id %s: %w", id, err)
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

// ResolveToID 将 short_token 解析为内部 ID
func (r *mediaRepo) ResolveToID(ctx context.Context, shortToken string) (string, error) {
	m, err := r.GetByShortToken(ctx, shortToken)
	if err != nil {
		return "", err
	}
	return m.Id, nil
}

func (r *mediaRepo) Get(
	ctx context.Context,
	idOrShortToken string,
	_ ...*dto.MediaQueryOption,
) (*types.Media, error) {
	// 优先按 short_token 查询
	m, err := r.db.Media.Query().
		Where(media.ShortTokenEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		Only(ctx)
	if err == nil {
		return convpb.ConvertMediaToMediaPBFull(m), nil
	}
	// 失败后按 ID 查询
	m, err = r.db.Media.Query().
		Where(media.IDEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

// GetWithEntity returns both entity.Media (with loaded Edges) and types.Media.
func (r *mediaRepo) GetWithEntity(
	ctx context.Context,
	idOrShortToken string,
	_ ...*dto.MediaQueryOption,
) (*entity.Media, *types.Media, error) {
	// 优先按 short_token 查询
	m, err := r.db.Media.Query().
		Where(media.ShortTokenEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		Only(ctx)
	if err == nil {
		return m, convpb.ConvertMediaToMediaPBFull(m), nil
	}
	// 失败后按 ID 查询
	m, err = r.db.Media.Query().
		Where(media.IDEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}
	return m, convpb.ConvertMediaToMediaPBFull(m), nil
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
		query = query.Where(media.UserIDEQ(*opt.UserID))
	}
	if opt.CategoryID != nil {
		query = query.Where(media.HasCategoryWith(category.ID(*opt.CategoryID)))
	}
	if len(opt.CategoryIDs) > 0 {
		ids := make([]int64, len(opt.CategoryIDs))
		copy(ids, opt.CategoryIDs)
		query = query.Where(media.HasCategoryWith(category.IDIn(ids...)))
	}
	if opt.State != "" {
		query = query.Where(media.StateEQ(opt.State))
	} else if opt.Status != nil {
		state := fmt.Sprintf("%d", *opt.Status)
		query = query.Where(media.StateEQ(state))
	} else if !opt.AdminMode {
		query = query.Where(media.StateEQ("active"))
	}

	if opt.MediaType != "" {
		if opt.MediaType == "video" {
			query = query.Where(media.TypeIn("video", "short_video"))
		} else {
			query = query.Where(media.TypeEQ(opt.MediaType))
		}
	}
	if opt.Keyword != "" {
		query = query.Where(media.TitleContains(opt.Keyword))
	}
	if len(opt.Tags) > 0 {
		query = query.Where(func(s *sql.Selector) {
			predicates := make([]*sql.Predicate, 0, len(opt.Tags))
			for _, tag := range opt.Tags {
				predicates = append(predicates, sqljson.ValueContains(media.FieldTags, tag))
			}
			s.Where(sql.Or(predicates...))
		})
	}
	if opt.Featured != nil {
		query = query.Where(media.FeaturedEQ(*opt.Featured))
	}
	if opt.Listable != nil {
		query = query.Where(media.ListableEQ(*opt.Listable))
	}
	if opt.ReviewStatus != nil {
		query = query.Where(media.ReviewStatusEQ(*opt.ReviewStatus))
	}
	if opt.Privacy != nil {
		query = query.Where(media.PrivacyEQ(convpb.ConvertPrivacyPBToMediaPrivacy(types.Privacy(*opt.Privacy))))
	} else if !opt.AdminMode {
		// Non-admin mode: exclude private media by default
		query = query.Where(media.PrivacyNEQ(media.PrivacyPRIVATE))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := opt.OrderBy
	if orderBy == "" {
		orderBy = "create_time"
	}

	desc := opt.Descending
	// Default to DESC for create_time (newest first) when not explicitly specified
	if !opt.Descending && opt.OrderBy == "" {
		desc = true
	}

	switch orderBy {
	case "title":
		if desc {
			query = query.Order(entity.Desc(media.FieldTitle))
		} else {
			query = query.Order(entity.Asc(media.FieldTitle))
		}
	case "view_count":
		if desc {
			query = query.Order(entity.Desc(media.FieldViewCount))
		} else {
			query = query.Order(entity.Asc(media.FieldViewCount))
		}
	case "create_time":
		fallthrough
	default:
		if desc {
			query = query.Order(entity.Desc(media.FieldCreateTime))
		} else {
			query = query.Order(entity.Asc(media.FieldCreateTime))
		}
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
		WithUser().
		WithCategory().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.Media, len(items))
	for i, item := range items {
		result[i] = convpb.ConvertMediaToMediaPBFull(item)
	}
	return result, int32(total), nil
}

// ListWithEntities returns both entity.Media (with loaded Edges) and types.Media.
// This allows the server layer to access Edges (e.g., User, Category) without N+1 queries.
func (r *mediaRepo) ListWithEntities(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*entity.Media, []*types.Media, int32, error) {
	opt := &dto.MediaQueryOption{}
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	query := r.db.Media.Query()

	// Apply filters (same as List)
	if opt.UserID != nil {
		query = query.Where(media.UserIDEQ(*opt.UserID))
	}
	if opt.CategoryID != nil {
		query = query.Where(media.HasCategoryWith(category.ID(*opt.CategoryID)))
	}
	if opt.State != "" {
		query = query.Where(media.StateEQ(opt.State))
	} else if opt.Status != nil {
		state := fmt.Sprintf("%d", *opt.Status)
		query = query.Where(media.StateEQ(state))
	} else if !opt.AdminMode {
		query = query.Where(media.StateEQ("active"))
	}

	if opt.MediaType != "" {
		if opt.MediaType == "video" {
			query = query.Where(media.TypeIn("video", "short_video"))
		} else {
			query = query.Where(media.TypeEQ(opt.MediaType))
		}
	}
	if opt.Keyword != "" {
		query = query.Where(media.TitleContains(opt.Keyword))
	}
	if len(opt.Tags) > 0 {
		query = query.Where(func(s *sql.Selector) {
			predicates := make([]*sql.Predicate, 0, len(opt.Tags))
			for _, tag := range opt.Tags {
				predicates = append(predicates, sqljson.ValueContains(media.FieldTags, tag))
			}
			s.Where(sql.Or(predicates...))
		})
	}
	if opt.Featured != nil {
		query = query.Where(media.FeaturedEQ(*opt.Featured))
	}
	if opt.Listable != nil {
		query = query.Where(media.ListableEQ(*opt.Listable))
	}
	if opt.ReviewStatus != nil {
		query = query.Where(media.ReviewStatusEQ(*opt.ReviewStatus))
	}
	if opt.Privacy != nil {
		query = query.Where(media.PrivacyEQ(convpb.ConvertPrivacyPBToMediaPrivacy(types.Privacy(*opt.Privacy))))
	} else if !opt.AdminMode {
		query = query.Where(media.PrivacyNEQ(media.PrivacyPRIVATE))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	// Apply sorting (same as List)
	orderBy := opt.OrderBy
	if orderBy == "" {
		orderBy = "create_time"
	}
	desc := opt.Descending
	// Default to DESC for create_time (newest first) when not explicitly specified
	if !opt.Descending && opt.OrderBy == "" {
		desc = true
	}

	switch orderBy {
	case "title":
		if desc {
			query = query.Order(entity.Desc(media.FieldTitle))
		} else {
			query = query.Order(entity.Asc(media.FieldTitle))
		}
	case "view_count":
		if desc {
			query = query.Order(entity.Desc(media.FieldViewCount))
		} else {
			query = query.Order(entity.Asc(media.FieldViewCount))
		}
	case "create_time":
		fallthrough
	default:
		if desc {
			query = query.Order(entity.Desc(media.FieldCreateTime))
		} else {
			query = query.Order(entity.Asc(media.FieldCreateTime))
		}
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
		WithUser().
		WithCategory().
		All(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	result := make([]*types.Media, len(items))
	for i, item := range items {
		result[i] = convpb.ConvertMediaToMediaPBFull(item)
	}
	return items, result, int32(total), nil
}

func (r *mediaRepo) Create(
	ctx context.Context,
	in *types.Media,
	_ ...*dto.MediaCreateOption,
) (*types.Media, error) {
	create := r.db.Media.Create().
		SetTitle(in.Title).
		SetURL(in.Url).
		SetType(in.Type).
		SetMimeType(in.MimeType).
		SetSize(fmt.Sprintf("%d", in.Size)).
		SetState(in.State).
		SetPrivacy(convpb.ConvertPrivacyPBToMediaPrivacy(in.Privacy)).
		SetEncodingStatus(in.EncodingStatus).
		SetAllowDownload(in.AllowDownload).
		SetEnableComments(in.EnableComments).
		SetFeatured(in.Featured)

	if in.Description != "" {
		create = create.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		create = create.SetThumbnail(in.Thumbnail)
	}
	if in.Duration > 0 {
		create = create.SetDuration(int(in.Duration))
	}
	if in.UserId != "" {
		create = create.SetUserID(in.UserId)
	}
	if in.CategoryId != 0 {
		create = create.SetNillableCategoryID(&in.CategoryId)
	}
	if len(in.Tags) > 0 {
		create = create.SetTags(in.Tags)
	}
	if in.ReviewStatus != "" {
		create = create.SetReviewStatus(in.ReviewStatus)
	}
	create = create.SetListable(in.Listable)
	if in.Extension != "" {
		create = create.SetExtension(in.Extension)
	}
	if in.Md5Sum != "" {
		create = create.SetMd5sum(in.Md5Sum)
	}
	if in.Poster != "" {
		create = create.SetPoster(in.Poster)
	}

	m, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

// CreateWithEntity creates a new media and returns both entity.Media and types.Media.
func (r *mediaRepo) CreateWithEntity(
	ctx context.Context,
	in *types.Media,
) (*entity.Media, *types.Media, error) {
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
	if in.UserId != "" {
		create = create.SetUserID(in.UserId)
	}
	if in.CategoryId != 0 {
		create = create.SetNillableCategoryID(&in.CategoryId)
	}
	if len(in.Tags) > 0 {
		create = create.SetTags(in.Tags)
	}
	if in.ReviewStatus != "" {
		create = create.SetReviewStatus(in.ReviewStatus)
	}
	create = create.SetListable(in.Listable)
	if in.Extension != "" {
		create = create.SetExtension(in.Extension)
	}
	if in.Md5Sum != "" {
		create = create.SetMd5sum(in.Md5Sum)
	}
	if in.Poster != "" {
		create = create.SetPoster(in.Poster)
	}

	m, err := create.Save(ctx)
	if err != nil {
		return nil, nil, err
	}
	return m, convpb.ConvertMediaToMediaPBFull(m), nil
}

func (r *mediaRepo) Update(
	ctx context.Context,
	in *types.Media,
	_ ...*dto.MediaUpdateOption,
) (*types.Media, error) {
	update := r.db.Media.UpdateOneID(in.Id).
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
	if in.PreviewFilePath != "" {
		update = update.SetPreviewFilePath(in.PreviewFilePath)
	}
	if in.EncodingStatus != "" {
		update = update.SetEncodingStatus(in.EncodingStatus)
	}
	// Uuid field is deprecated, use Id instead
	// if in.Uuid != "" {
	// 	update = update.SetUUID(in.Uuid)
	// }
	if in.Duration > 0 {
		update = update.SetDuration(int(in.Duration))
	}
	if in.CategoryId != 0 {
		update = update.SetNillableCategoryID(&in.CategoryId)
	}
	// Update tags
	update = update.SetTags(in.Tags)

	// Update review_status and listable
	if in.ReviewStatus != "" {
		update = update.SetReviewStatus(in.ReviewStatus)
	}
	update = update.SetListable(in.Listable)
	if in.State != "" {
		update = update.SetState(in.State)
	}
	update = update.SetPrivacy(convpb.ConvertPrivacyPBToMediaPrivacy(in.Privacy))
	update = update.SetFeatured(in.Featured)
	update = update.SetAllowDownload(in.AllowDownload)
	update = update.SetEnableComments(in.EnableComments)

	m, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

func (r *mediaRepo) Delete(ctx context.Context, id string) error {
	return r.db.Media.DeleteOneID(id).Exec(ctx)
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

func (r *mediaRepo) GetCategory(ctx context.Context, id string) (*types.Category, error) {
	catID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid category id: %w", err)
	}
	c, err := r.db.Category.Get(ctx, catID)
	if err != nil {
		return nil, err
	}
	return convertCategoryToProto(c), nil
}

func (r *mediaRepo) IncrementViewCount(ctx context.Context, idOrShortToken string) (int64, error) {
	id, err := r.getMediaID(ctx, idOrShortToken)
	if err != nil {
		return 0, err
	}
	m, err := r.db.Media.UpdateOneID(id).
		AddViewCount(1).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return m.ViewCount, nil
}

func (r *mediaRepo) UpdateCommentCount(ctx context.Context, idOrShortToken string, delta int) error {
	id, err := r.getMediaID(ctx, idOrShortToken)
	if err != nil {
		return err
	}
	return r.db.Media.UpdateOneID(id).
		AddCommentCount(int64(delta)).
		Exec(ctx)
}

func (r *mediaRepo) UpdateLikeCount(ctx context.Context, idOrShortToken string, delta int) error {
	id, err := r.getMediaID(ctx, idOrShortToken)
	if err != nil {
		return err
	}
	return r.db.Media.UpdateOneID(id).
		AddLikeCount(int64(delta)).
		Exec(ctx)
}

func (r *mediaRepo) UpdateDislikeCount(ctx context.Context, idOrShortToken string, delta int) error {
	id, err := r.getMediaID(ctx, idOrShortToken)
	if err != nil {
		return err
	}
	return r.db.Media.UpdateOneID(id).
		AddDislikeCount(int64(delta)).
		Exec(ctx)
}

func (r *mediaRepo) UpdateFavoriteCount(ctx context.Context, idOrShortToken string, delta int) error {
	id, err := r.getMediaID(ctx, idOrShortToken)
	if err != nil {
		return err
	}
	return r.db.Media.UpdateOneID(id).
		AddFavoriteCount(int64(delta)).
		Exec(ctx)
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
func (r *mediaRepo) ListFilteredByEncodingStatus(
	ctx context.Context,
	statuses []string,
	page, pageSize int,
) ([]*types.Media, int, error) {
	if len(statuses) == 0 {
		return nil, 0, nil
	}

	query := r.db.Media.Query().
		Where(media.EncodingStatusIn(statuses...)).
		Order(entity.Desc(media.FieldUpdateTime))

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
		result[i] = convpb.ConvertMediaToMediaPBFull(item)
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
			return 0, fmt.Errorf("delete orphaned tasks for media %s: %w", m.ID, err)
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

func (r *mediaRepo) getMediaID(ctx context.Context, idOrShortToken string) (string, error) {
	// 优先按 short_token 查询
	m, err := r.db.Media.Query().
		Where(media.ShortTokenEQ(idOrShortToken)).
		Only(ctx)
	if err == nil {
		return m.ID, nil
	}
	// 失败后按 ID 查询
	m, err = r.db.Media.Get(ctx, idOrShortToken)
	if err != nil {
		return "", err
	}
	return m.ID, nil
}

func (r *mediaRepo) UpdateSpriteFields(ctx context.Context, mediaID string, spriteStatus string, spritePath string, vttPath string) error {
	update := r.db.Media.UpdateOneID(mediaID).
		SetSpriteStatus(spriteStatus)
	if spritePath != "" {
		update = update.SetSpritePath(spritePath)
	}
	if vttPath != "" {
		update = update.SetVttPath(vttPath)
	}
	return update.Exec(ctx)
}

func (r *mediaRepo) UpdateThumbnailFields(ctx context.Context, mediaID string, thumbnail string, thumbnailTime float64) error {
	return r.db.Media.UpdateOneID(mediaID).
		SetThumbnail(thumbnail).
		SetThumbnailTime(thumbnailTime).
		Exec(ctx)
}

func (r *mediaRepo) UpdatePreviewFilePath(ctx context.Context, mediaID string, previewFilePath string) error {
	return r.db.Media.UpdateOneID(mediaID).
		SetPreviewFilePath(previewFilePath).
		Exec(ctx)
}

func (r *mediaRepo) UpdateDimensions(ctx context.Context, mediaID string, width, height int) error {
	return r.db.Media.UpdateOneID(mediaID).
		SetWidth(width).
		SetHeight(height).
		Exec(ctx)
}

// GetEntityByID returns the raw entity.Media by ID for accessing internal fields.
func (r *mediaRepo) GetEntityByID(ctx context.Context, id string) (*entity.Media, error) {
	return r.db.Media.Get(ctx, id)
}

// GetEntityByShortToken returns the raw entity.Media by short_token for accessing internal fields.
func (r *mediaRepo) GetEntityByShortToken(ctx context.Context, shortToken string) (*entity.Media, error) {
	return r.db.Media.Query().
		Where(media.ShortTokenEQ(shortToken)).
		Only(ctx)
}

// convertCategoryToProto converts entity.Category → proto types.Category.
func convertCategoryToProto(c *entity.Category) *types.Category {
	return &types.Category{
		Id:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		MediaCount:  int64(c.MediaCount),
	}
}

// ListTempMediaBefore returns media records whose URL starts with "temp/" and
// whose create_time is before the given cutoff. Used by CleanupExpiredTemp to
// find stale temp files that were never promoted (failed/expired transcodes).
func (r *mediaRepo) ListTempMediaBefore(ctx context.Context, cutoff time.Time) ([]*types.Media, error) {
	items, err := r.db.Media.Query().
		Where(
			media.URLHasPrefix("temp/"),
			media.CreateTimeLT(cutoff),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query temp media before cutoff: %w", err)
	}

	result := make([]*types.Media, len(items))
	for i, item := range items {
		result[i] = convpb.ConvertMediaToMediaPBFull(item)
	}
	return result, nil
}
