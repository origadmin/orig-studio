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
	"origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/dal/convpb"
	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/category"
	"origadmin/application/origstudio/internal/dal/entity/channel"
	"origadmin/application/origstudio/internal/dal/entity/encodingtask"
	"origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/dal/entity/tag"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
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
		WithChannel().
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
		WithChannel().
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
		WithChannel().
		Only(ctx)
	if err == nil {
		return convpb.ConvertMediaToMediaPBFull(m), nil
	}
	// 失败后按 ID 查询
	m, err = r.db.Media.Query().
		Where(media.IDEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		WithChannel().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

// EntityToMediaEntityDTO converts an entity.Media to a dto.MediaEntityDTO.
func EntityToMediaEntityDTO(m *entity.Media) *dto.MediaEntityDTO {
	if m == nil {
		return nil
	}
	result := &dto.MediaEntityDTO{
		ID:              m.ID,
		Title:           m.Title,
		Type:            dto.MediaEntityType(string(m.Type)),
		URL:             m.URL,
		ShortToken:      m.ShortToken,
		Status:          string(m.State),
		ReviewStatus:    string(m.ReviewStatus),
		SpriteStatus:    m.SpriteStatus,
		SpritePath:      m.SpritePath,
		VttPath:         m.VttPath,
		Thumbnail:       m.Thumbnail,
		ThumbnailTime:   m.ThumbnailTime,
		PreviewFilePath: m.PreviewFilePath,
		Width:           m.Width,
		Height:          m.Height,
		Duration:        float64(m.Duration),
		ViewCount:       m.ViewCount,
		LikeCount:       m.LikeCount,
		DislikeCount:    m.DislikeCount,
		FavoriteCount:   m.FavoriteCount,
		CommentCount:    m.CommentCount,
		CreateTime:      m.CreateTime,
		UpdateTime:      m.UpdateTime,
	}
	// Populate edge data if loaded
	if m.Edges.User != nil {
		result.UserID = m.Edges.User.ID
		result.UserName = m.Edges.User.Username
		result.UserNickname = m.Edges.User.Name
		result.UserAvatar = m.Edges.User.Avatar
		result.UserSlug = m.Edges.User.Slug
	}
	if m.Edges.Category != nil {
		result.CategoryID = int(m.Edges.Category.ID)
		result.CategoryName = m.Edges.Category.Name
	}
	return result
}

// GetWithEntity returns both MediaEntityDTO (with loaded edges) and types.Media.
func (r *mediaRepo) GetWithEntity(
	ctx context.Context,
	idOrShortToken string,
	_ ...*dto.MediaQueryOption,
) (*dto.MediaEntityDTO, *types.Media, error) {
	// Try short_token first
	m, err := r.db.Media.Query().
		Where(media.ShortTokenEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		WithChannel().
		Only(ctx)
	if err == nil {
		return EntityToMediaEntityDTO(m), convpb.ConvertMediaToMediaPBFull(m), nil
	}
	// Fall back to ID
	m, err = r.db.Media.Query().
		Where(media.IDEQ(idOrShortToken)).
		WithUser().
		WithCategory().
		WithChannel().
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}
	return EntityToMediaEntityDTO(m), convpb.ConvertMediaToMediaPBFull(m), nil
}

func (r *mediaRepo) List(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*types.Media, int32, error) {
	opt := &dto.MediaQueryOption{}
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	// BUG-164: server-side taxonomy tree expansion. The backend owns the
	// "category -> related media" semantics: given any category id (typically a
	// parent/ancestor from the caller), expand to the full subtree before
	// filtering. Idempotent when the input already includes the subtree (the
	// current portal pre-expands parent+children client-side), so this is a safe,
	// forward-compatible step toward the slug-driven, backend-owned expansion.
	if len(opt.CategoryIDs) > 0 {
		expanded, err := r.expandCategorySubtree(ctx, opt.CategoryIDs)
		if err != nil {
			return nil, 0, err
		}
		opt.CategoryIDs = expanded
	}

	// BUG-226: bound the random-order candidate pool to recent active media so the
	// seeded shuffle sorts a small, indexed window instead of the whole table.
	// Reuses opt.CreatedAfter -> ent parameterized CreateTimeGTE (injection-safe).
	const randomCandidateWindowDays = 90
	if opt.OrderBy == "random" && opt.CreatedAfter == "" {
		opt.CreatedAfter = time.Now().AddDate(0, 0, -randomCandidateWindowDays).Format(time.RFC3339)
	}

	// Build a fresh query for counting (to avoid ent ORM Count side effects on the original query)
	countQuery := r.db.Media.Query()
	// Build the main query for fetching data
	query := r.db.Media.Query()

	// Apply filters to both queries
	applyFilters := func(q *entity.MediaQuery) *entity.MediaQuery {
		if opt.UserID != nil {
			q = q.Where(media.UserIDEQ(*opt.UserID))
		}
		// BUG-105: channel assignment filter. "Unassigned" (channel_id IS NULL)
		// takes precedence over exact channel id.
		if opt.ChannelUnassigned {
			q = q.Where(media.ChannelIDIsNil())
		} else if opt.ChannelID != nil {
			q = q.Where(media.ChannelIDEQ(*opt.ChannelID))
		}
		if opt.CategoryID != nil {
			q = q.Where(media.HasCategoryWith(category.ID(*opt.CategoryID)))
		}
		if len(opt.CategoryIDs) > 0 {
			ids := make([]int64, len(opt.CategoryIDs))
			copy(ids, opt.CategoryIDs)
			q = q.Where(media.HasCategoryWith(category.IDIn(ids...)))
		}
		if opt.CreatedAfter != "" {
			if ts, err := time.Parse(time.RFC3339, opt.CreatedAfter); err == nil {
				q = q.Where(media.CreateTimeGTE(ts))
			}
		}
		if opt.State != "" {
			q = q.Where(media.StateEQ(opt.State))
		} else if opt.Status != nil {
			state := fmt.Sprintf("%d", *opt.Status)
			q = q.Where(media.StateEQ(state))
		} else if !opt.AdminMode {
			// BUG-141 ①② portal visibility gate.
			if opt.OwnerView {
				q = q.Where(media.StateEQ("active")) // owner sees own active (incl. unreviewed)
			} else if opt.Listable == nil {
				q = q.Where(media.ListableEQ(true)) // others: only reviewed+encoded+active
			}
		}

		if opt.MediaType != "" {
			q = q.Where(media.TypeEQ(opt.MediaType))
		}
		if opt.Keyword != "" {
			q = q.Where(media.TitleContains(opt.Keyword))
		}
		if len(opt.Tags) > 0 {
			q = q.Where(func(s *sql.Selector) {
				predicates := make([]*sql.Predicate, 0, len(opt.Tags))
				for _, tag := range opt.Tags {
					predicates = append(predicates, sqljson.ValueContains(media.FieldTags, tag))
				}
				s.Where(sql.Or(predicates...))
			})
		}
		if opt.Featured != nil {
			q = q.Where(media.FeaturedEQ(*opt.Featured))
		}
		if opt.Listable != nil {
			q = q.Where(media.ListableEQ(*opt.Listable))
		}
		if opt.ReviewStatus != nil {
			q = q.Where(media.ReviewStatusEQ(*opt.ReviewStatus))
		}
		if opt.Privacy != nil {
			q = q.Where(media.PrivacyEQ(convpb.ConvertPrivacyPBToMediaPrivacy(types.Privacy(*opt.Privacy))))
		} else if !opt.AdminMode && !opt.OwnerView {
			q = q.Where(media.PrivacyNEQ(media.PrivacyPRIVATE))
		}
		return q
	}

	countQuery = applyFilters(countQuery)
	query = applyFilters(query)

	if opt.Page < 1 {
		opt.Page = 1
	}
	if opt.PageSize < 1 {
		opt.PageSize = 20
	}

	// Count with all filters applied using independent query
	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting after count
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
	case "random":
		if opt.RandomSeed != nil {
			// PostgreSQL has no RAND(seed); use a parameterized seeded hash for a
			// deterministic, repeatable shuffle. seed is a bound numeric arg via
			// b.Arg (emits $N on PostgreSQL) — never concatenated into SQL text.
			// Same (seed, id) => same order across connections; different seed => different order.
			seed := *opt.RandomSeed
			query = query.Order(func(s *sql.Selector) {
				s.OrderExpr(sql.ExprFunc(func(b *sql.Builder) {
					b.WriteString("md5(concat(")
					b.Ident(media.FieldID).WriteString("::text")
					b.WriteString(", ")
					b.Arg(seed)
					b.WriteString("::text))")
				}))
			})
		} else {
			query = query.Order(entity.Desc(media.FieldCreateTime))
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

	offset := (opt.Page - 1) * opt.PageSize
	items, err := query.Offset(int(offset)).
		Limit(int(opt.PageSize)).
		WithUser().
		WithCategory().
		WithChannel().
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

// expandCategorySubtree returns the given category ids plus all of their
// descendant ids (BFS over the parent_id self-relation). It implements the
// BUG-164 "category -> related media" server-side taxonomy-tree expansion so
// callers may pass a parent (or any ancestor) and receive the full subtree.
// Idempotent when the input already contains the full subtree.
func (r *mediaRepo) expandCategorySubtree(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return ids, nil
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		seen[id] = struct{}{}
	}
	current := ids
	for len(current) > 0 {
		children, err := r.db.Category.Query().
			Where(category.ParentIDIn(current...)).
			IDs(ctx)
		if err != nil {
			return nil, err
		}
		next := make([]int64, 0, len(children))
		for _, child := range children {
			if _, ok := seen[child]; !ok {
				seen[child] = struct{}{}
				next = append(next, child)
			}
		}
		current = next
	}
	out := make([]int64, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}

// ListWithEntities returns both MediaEntityDTO (with loaded edges) and types.Media.
// This allows the server layer to access edges (e.g., User, Category) without N+1 queries.
func (r *mediaRepo) ListWithEntities(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*dto.MediaEntityDTO, []*types.Media, int32, error) {
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
	if opt.CreatedAfter != "" {
		if ts, err := time.Parse(time.RFC3339, opt.CreatedAfter); err == nil {
			query = query.Where(media.CreateTimeGTE(ts))
		}
	}
	if opt.State != "" {
		query = query.Where(media.StateEQ(opt.State))
	} else if opt.Status != nil {
		state := fmt.Sprintf("%d", *opt.Status)
		query = query.Where(media.StateEQ(state))
	} else if !opt.AdminMode {
		// BUG-141 ①② portal visibility gate (mirror of List).
		if opt.OwnerView {
			query = query.Where(media.StateEQ("active"))
		} else if opt.Listable == nil {
			query = query.Where(media.ListableEQ(true))
		}
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
	} else if !opt.AdminMode && !opt.OwnerView {
		query = query.Where(media.PrivacyNEQ(media.PrivacyPRIVATE))
	}

	if opt.Page < 1 {
		opt.Page = 1
	}
	if opt.PageSize < 1 {
		opt.PageSize = 20
	}

	// Count with all filters applied
	countQuery := query
	total, err := countQuery.Count(ctx)
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
	case "random":
		if opt.RandomSeed != nil {
			// PostgreSQL has no RAND(seed); use a parameterized seeded hash for a
			// deterministic, repeatable shuffle. seed is a bound numeric arg via
			// b.Arg (emits $N on PostgreSQL) — never concatenated into SQL text.
			// Same (seed, id) => same order across connections; different seed => different order.
			seed := *opt.RandomSeed
			query = query.Order(func(s *sql.Selector) {
				s.OrderExpr(sql.ExprFunc(func(b *sql.Builder) {
					b.WriteString("md5(concat(")
					b.Ident(media.FieldID).WriteString("::text")
					b.WriteString(", ")
					b.Arg(seed)
					b.WriteString("::text))")
				}))
			})
		} else {
			query = query.Order(entity.Desc(media.FieldCreateTime))
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

	offset := (opt.Page - 1) * opt.PageSize
	items, err := query.Offset(int(offset)).
		Limit(int(opt.PageSize)).
		WithUser().
		WithCategory().
		WithChannel().
		All(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	result := make([]*types.Media, len(items))
	entityDTOs := make([]*dto.MediaEntityDTO, len(items))
	for i, item := range items {
		result[i] = convpb.ConvertMediaToMediaPBFull(item)
		entityDTOs[i] = EntityToMediaEntityDTO(item)
	}
	return entityDTOs, result, int32(total), nil
}

func (r *mediaRepo) Create(
	ctx context.Context,
	in *types.Media,
	_ ...*dto.MediaCreateOption,
) (*types.Media, error) {
	// BUG-131/132: converge the three tag sources (manual / title / intro)
	// into one canonical set BEFORE persisting; jsonb projection and the
	// authoritative M2M pivot both get the merged set.
	mergedTags := biz.NormalizeAndMerge(in.Title, in.Description, in.Tags)
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
		SetFeatured(in.Featured).
		SetListable(in.Listable)

	if in.Description != "" {
		create = create.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		create = create.SetThumbnail(in.Thumbnail)
	}
	if in.Poster != "" {
		create = create.SetPoster(in.Poster)
	}
	if in.HlsFile != "" {
		create = create.SetHlsFile(in.HlsFile)
	}
	if in.PreviewFilePath != "" {
		create = create.SetPreviewFilePath(in.PreviewFilePath)
	}
	if in.Duration > 0 {
		create = create.SetDuration(int(in.Duration))
	}
	if in.Width > 0 {
		create = create.SetWidth(int(in.Width))
	}
	if in.Height > 0 {
		create = create.SetHeight(int(in.Height))
	}
	if in.UserId != "" {
		create = create.SetUserID(in.UserId)
	}
	if in.CategoryId != 0 {
		create = create.SetNillableCategoryID(&in.CategoryId)
	}
	if in.ChannelId != "" {
		create = create.SetChannelID(in.ChannelId)
	}
	// Projection: jsonb keeps the merged canonical set (was: raw manual tags).
	create = create.SetTags(mergedTags)
	if in.ReviewStatus != "" {
		create = create.SetReviewStatus(in.ReviewStatus)
	}
	if in.Extension != "" {
		create = create.SetExtension(in.Extension)
	}
	if in.Sha256 != "" {
		create = create.SetSha256(in.Sha256)
	}
	if in.ThumbnailTime > 0 {
		create = create.SetThumbnailTime(in.ThumbnailTime)
	}
	if in.SpriteStatus != "" {
		create = create.SetSpriteStatus(in.SpriteStatus)
	} else {
		if strings.Contains(in.MimeType, "video") {
			create = create.SetSpriteStatus("pending")
		} else {
			create = create.SetSpriteStatus("none")
		}
	}
	if in.SpritePath != "" {
		create = create.SetSpritePath(in.SpritePath)
	}
	if in.VttPath != "" {
		create = create.SetVttPath(in.VttPath)
	}
	if in.CreateAuthor != "" {
		create = create.SetCreateAuthor(in.CreateAuthor)
	}
	if in.UpdateAuthor != "" {
		create = create.SetUpdateAuthor(in.UpdateAuthor)
	}

	m, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	// BUG-132: persist the authoritative M2M rows (content_media_tags).
	if err := r.SyncMediaTags(ctx, m.ID, mergedTags); err != nil {
		return nil, err
	}
	return convpb.ConvertMediaToMediaPBFull(m), nil
}

// CreateWithEntity creates a new media and returns both MediaEntityDTO and types.Media.
func (r *mediaRepo) CreateWithEntity(
	ctx context.Context,
	in *types.Media,
) (*dto.MediaEntityDTO, *types.Media, error) {
	// BUG-131/132: same write-time merge as Create().
	mergedTags := biz.NormalizeAndMerge(in.Title, in.Description, in.Tags)
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
		SetFeatured(in.Featured).
		SetListable(in.Listable)

	if in.Description != "" {
		create = create.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		create = create.SetThumbnail(in.Thumbnail)
	}
	if in.Poster != "" {
		create = create.SetPoster(in.Poster)
	}
	if in.HlsFile != "" {
		create = create.SetHlsFile(in.HlsFile)
	}
	if in.PreviewFilePath != "" {
		create = create.SetPreviewFilePath(in.PreviewFilePath)
	}
	if in.Duration > 0 {
		create = create.SetDuration(int(in.Duration))
	}
	if in.Width > 0 {
		create = create.SetWidth(int(in.Width))
	}
	if in.Height > 0 {
		create = create.SetHeight(int(in.Height))
	}
	if in.UserId != "" {
		create = create.SetUserID(in.UserId)
	}
	if in.CategoryId != 0 {
		create = create.SetNillableCategoryID(&in.CategoryId)
	}
	if in.ChannelId != "" {
		create = create.SetChannelID(in.ChannelId)
	}
	// Projection: jsonb keeps the merged canonical set (was: raw manual tags).
	create = create.SetTags(mergedTags)
	if in.ReviewStatus != "" {
		create = create.SetReviewStatus(in.ReviewStatus)
	}
	if in.Extension != "" {
		create = create.SetExtension(in.Extension)
	}
	if in.Sha256 != "" {
		create = create.SetSha256(in.Sha256)
	}
	if in.ThumbnailTime > 0 {
		create = create.SetThumbnailTime(in.ThumbnailTime)
	}
	if in.SpriteStatus != "" {
		create = create.SetSpriteStatus(in.SpriteStatus)
	} else {
		if strings.Contains(in.MimeType, "video") || in.Type == "video" {
			create = create.SetSpriteStatus("pending")
		} else {
			create = create.SetSpriteStatus("none")
		}
	}
	if in.SpritePath != "" {
		create = create.SetSpritePath(in.SpritePath)
	}
	if in.VttPath != "" {
		create = create.SetVttPath(in.VttPath)
	}
	if in.CreateAuthor != "" {
		create = create.SetCreateAuthor(in.CreateAuthor)
	}
	if in.UpdateAuthor != "" {
		create = create.SetUpdateAuthor(in.UpdateAuthor)
	}

	m, err := create.Save(ctx)
	if err != nil {
		return nil, nil, err
	}
	// BUG-132: persist the authoritative M2M rows (content_media_tags).
	if err := r.SyncMediaTags(ctx, m.ID, mergedTags); err != nil {
		return nil, nil, err
	}
	return EntityToMediaEntityDTO(m), convpb.ConvertMediaToMediaPBFull(m), nil
}

func (r *mediaRepo) Update(
	ctx context.Context,
	in *types.Media,
	_ ...*dto.MediaUpdateOption,
) (*types.Media, error) {
	// BUG-131/132: converge the three tag sources (manual / title / intro)
	// into one canonical set BEFORE persisting; jsonb projection and the
	// authoritative M2M pivot both get the merged set.
	mergedTags := biz.NormalizeAndMerge(in.Title, in.Description, in.Tags)
	update := r.db.Media.UpdateOneID(in.Id).
		SetTitle(in.Title).
		SetMimeType(in.MimeType).
		SetSize(fmt.Sprintf("%d", in.Size)).
		SetListable(in.Listable).
		SetFeatured(in.Featured).
		SetAllowDownload(in.AllowDownload).
		SetEnableComments(in.EnableComments).
		SetPrivacy(convpb.ConvertPrivacyPBToMediaPrivacy(in.Privacy))

	if in.Description != "" {
		update = update.SetDescription(in.Description)
	}
	if in.Thumbnail != "" {
		update = update.SetThumbnail(in.Thumbnail)
	}
	if in.Poster != "" {
		update = update.SetPoster(in.Poster)
	}
	if in.Url != "" {
		update = update.SetURL(in.Url)
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
	if in.Duration > 0 {
		update = update.SetDuration(int(in.Duration))
	}
	if in.Width > 0 {
		update = update.SetWidth(int(in.Width))
	}
	if in.Height > 0 {
		update = update.SetHeight(int(in.Height))
	}
	if in.CategoryId != 0 {
		update = update.SetNillableCategoryID(&in.CategoryId)
	}
	if in.ChannelId != "" {
		update = update.SetChannelID(in.ChannelId)
	}
	if in.Extension != "" {
		update = update.SetExtension(in.Extension)
	}
	if in.Sha256 != "" {
		update = update.SetSha256(in.Sha256)
	}
	if in.ThumbnailTime > 0 {
		update = update.SetThumbnailTime(in.ThumbnailTime)
	}
	if in.SpriteStatus != "" {
		update = update.SetSpriteStatus(in.SpriteStatus)
	}
	if in.SpritePath != "" {
		update = update.SetSpritePath(in.SpritePath)
	}
	if in.VttPath != "" {
		update = update.SetVttPath(in.VttPath)
	}
	// Update tags: projection keeps the merged canonical set.
	update = update.SetTags(mergedTags)

	// Update review_status
	if in.ReviewStatus != "" {
		update = update.SetReviewStatus(in.ReviewStatus)
	}
	if in.State != "" {
		update = update.SetState(in.State)
	}
	if in.UpdateAuthor != "" {
		update = update.SetUpdateAuthor(in.UpdateAuthor)
	}

	m, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	// BUG-132: persist the authoritative M2M rows (content_media_tags).
	if err := r.SyncMediaTags(ctx, m.ID, mergedTags); err != nil {
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

func (r *mediaRepo) UpdateReportedTimes(ctx context.Context, idOrShortToken string, delta int) error {
	id, err := r.getMediaID(ctx, idOrShortToken)
	if err != nil {
		return err
	}
	return r.db.Media.UpdateOneID(id).
		AddReportedTimes(delta).
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

// GetEntityByID returns the MediaEntityDTO by ID for accessing internal fields.
func (r *mediaRepo) GetEntityByID(ctx context.Context, id string) (*dto.MediaEntityDTO, error) {
	m, err := r.db.Media.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return EntityToMediaEntityDTO(m), nil
}

// GetEntityByShortToken returns the MediaEntityDTO by short_token for accessing internal fields.
func (r *mediaRepo) GetEntityByShortToken(ctx context.Context, shortToken string) (*dto.MediaEntityDTO, error) {
	m, err := r.db.Media.Query().
		Where(media.ShortTokenEQ(shortToken)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return EntityToMediaEntityDTO(m), nil
}

// convertCategoryToProto converts entity.Category → proto types.Category.
func convertCategoryToProto(c *entity.Category) *types.Category {
	return &types.Category{
		Id:                c.ID,
		Name:              c.Name,
		Slug:              c.Slug,
		Description:       c.Description,
		Thumbnail:         c.Thumbnail,
		ListingsThumbnail: c.ListingsThumbnail,
		Icon:              c.Icon,
		Color:             c.Color,
		ParentId:          c.ParentID,
		Sequence:          int32(c.Sequence),
		Status:            convpb.ConvertCategoryStatusToInt32(c.Status),
		MediaCount:        int64(c.MediaCount),
	}
}

func (r *mediaRepo) CreateCategory(ctx context.Context, c *types.Category) (*types.Category, error) {
	builder := r.db.Category.Create().
		SetName(c.Name).
		SetSlug(c.Slug).
		SetDescription(c.Description)
	if c.ParentId > 0 {
		builder.SetParentID(c.ParentId)
	}
	if c.Thumbnail != "" {
		builder.SetThumbnail(c.Thumbnail)
	}
	if c.Icon != "" {
		builder.SetIcon(c.Icon)
	}
	if c.Color != "" {
		builder.SetColor(c.Color)
	}
	if c.Sequence != 0 {
		builder.SetSequence(int(c.Sequence))
	}
	if c.Status != 0 {
		builder.SetStatus(convpb.ConvertInt32ToCategoryStatus(c.Status))
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertCategoryToProto(created), nil
}

func (r *mediaRepo) UpdateCategory(ctx context.Context, c *types.Category) (*types.Category, error) {
	builder := r.db.Category.UpdateOneID(c.Id).
		SetName(c.Name).
		SetSlug(c.Slug).
		SetDescription(c.Description)
	if c.ParentId > 0 {
		builder.SetParentID(c.ParentId)
	}
	if c.Thumbnail != "" {
		builder.SetThumbnail(c.Thumbnail)
	}
	if c.Icon != "" {
		builder.SetIcon(c.Icon)
	}
	if c.Color != "" {
		builder.SetColor(c.Color)
	}
	if c.Sequence != 0 {
		builder.SetSequence(int(c.Sequence))
	}
	if c.Status != 0 {
		builder.SetStatus(convpb.ConvertInt32ToCategoryStatus(c.Status))
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertCategoryToProto(updated), nil
}

func (r *mediaRepo) DeleteCategory(ctx context.Context, id string) error {
	catID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid category id: %w", err)
	}
	return r.db.Category.DeleteOneID(catID).Exec(ctx)
}

func (r *mediaRepo) ListTags(ctx context.Context, opts ...*dto.TagQueryOption) ([]*types.Tag, int32, error) {
	var opt *dto.TagQueryOption
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}
	if opt == nil {
		opt = &dto.TagQueryOption{}
	}

	query := r.db.Tag.Query()
	if opt.Keyword != "" {
		query = query.Where(tag.TitleContainsFold(opt.Keyword))
	}
	if opt.Status != "" {
		upper := strings.ToUpper(opt.Status)
		switch upper {
		case "ACTIVE", "INACTIVE":
			query = query.Where(tag.StatusEQ(tag.Status(upper)))
		}
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// BUG-180: legacy seed tags carry a NULL create_time, and
	// `ORDER BY create_time DESC` sorts NULLs FIRST — freshly created tags
	// (which do carry a timestamp) sank to the last page of the admin list and
	// looked "invisible". NULLS LAST keeps the newest tags on top.
	query = query.Order(
		sql.OrderByField(tag.FieldCreateTime, sql.OrderDesc(), sql.OrderNullsLast()).ToFunc(),
		sql.OrderByField(tag.FieldID, sql.OrderDesc()).ToFunc(),
	)

	page, pageSize := 1, 20
	if opt.Page > 0 {
		page = int(opt.Page)
	}
	if opt.PageSize > 0 {
		pageSize = int(opt.PageSize)
	}

	items, err := query.Limit(pageSize).Offset((page - 1) * pageSize).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*types.Tag, len(items))
	for i, item := range items {
		pb := convertTagToProto(item)
		// BUG-180: the denormalized content_tags.media_count column is stale
		// for most tags. Derive from content_media.tags — the same source the
		// tag detail page filters on — so the admin count always matches the
		// number of videos the portal actually shows.
		pb.MediaCount = int64(r.countMediaByTag(ctx, item.Title))
		result[i] = pb
	}
	return result, int32(total), nil
}

// countMediaByTag mirrors the public media-list ?tags= filter
// (sqljson.ValueContains(media.FieldTags, title)) so the count shown in the
// admin tag list always equals the number of videos the tag detail page shows.
func (r *mediaRepo) countMediaByTag(ctx context.Context, title string) int {
	if title == "" {
		return 0
	}
	count, err := r.db.Media.Query().
		Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueContains(media.FieldTags, title))
		}).
		Count(ctx)
	if err != nil {
		return 0
	}
	return count
}

func (r *mediaRepo) GetTag(ctx context.Context, id string) (*types.Tag, error) {
	tagID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid tag id: %w", err)
	}
	t, err := r.db.Tag.Get(ctx, tagID)
	if err != nil {
		return nil, err
	}
	return convertTagToProto(t), nil
}

func (r *mediaRepo) CreateTag(ctx context.Context, t *types.Tag) (*types.Tag, error) {
	builder := r.db.Tag.Create().
		SetTitle(t.Title).
		SetSlug(t.Slug).
		SetDescription(t.Description)
	if t.Color != "" {
		builder.SetColor(t.Color)
	}
	if t.Status != 0 {
		builder.SetStatus(convpb.ConvertTagStatusPBToTagStatus(t.Status))
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertTagToProto(created), nil
}

func (r *mediaRepo) UpdateTag(ctx context.Context, t *types.Tag) (*types.Tag, error) {
	builder := r.db.Tag.UpdateOneID(int(t.Id)).
		SetTitle(t.Title).
		SetSlug(t.Slug).
		SetDescription(t.Description)
	if t.Color != "" {
		builder.SetColor(t.Color)
	}
	if t.Status != 0 {
		builder.SetStatus(convpb.ConvertTagStatusPBToTagStatus(t.Status))
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertTagToProto(updated), nil
}

func (r *mediaRepo) DeleteTag(ctx context.Context, id string) error {
	tagID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid tag id: %w", err)
	}
	return r.db.Tag.DeleteOneID(tagID).Exec(ctx)
}

func convertTagToProto(t *entity.Tag) *types.Tag {
	return &types.Tag{
		Id:          int64(t.ID),
		Title:       t.Title,
		Slug:        t.Slug,
		Description: t.Description,
		Color:       t.Color,
		Status:      convpb.ConvertTagStatusToTagStatusPB(t.Status),
		MediaCount:  int64(t.MediaCount),
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

// GetDefaultChannelID returns the ID of the user's first (default) channel.
// This is used by UploadMedia to auto-bind videos that don't have a channel specified.
func (r *mediaRepo) GetDefaultChannelID(ctx context.Context, userID string) (string, error) {
	ch, err := r.db.Channel.Query().
		Where(channel.UserIDEQ(userID)).
		Order(entity.Asc(channel.FieldID)).
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("query default channel for user %s: %w", userID, err)
	}
	return ch.ID, nil
}

// GetChannelOwnerID returns the owner user id of a channel, or "" when the
// channel does not exist. Used by UpdateMedia (BUG-105) to validate that a
// caller may assign media to a channel before moving/setting channel_id.
func (r *mediaRepo) GetChannelOwnerID(ctx context.Context, channelID string) (string, error) {
	ch, err := r.db.Channel.Get(ctx, channelID)
	if err != nil {
		if entity.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("query channel owner %s: %w", channelID, err)
	}
	return ch.UserID, nil
}

// UpdateMediaChannel sets or clears a media's channel assignment.
// channelID == "" clears the assignment (move to unassigned); otherwise sets it
// (assign / move A->B). BUG-105: the generic Update path only SetChannelID for
// non-empty values and therefore cannot express "clear", so channel changes go
// through this dedicated path.
func (r *mediaRepo) UpdateMediaChannel(ctx context.Context, mediaID, channelID string) error {
	u := r.db.Media.UpdateOneID(mediaID)
	if channelID == "" {
		u = u.ClearChannelID()
	} else {
		u = u.SetChannelID(channelID)
	}
	return u.Exec(ctx)
}
