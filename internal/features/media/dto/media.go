/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"

	"origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/helpers/repo"
)

// MediaRepo is a Media repository interface.
type MediaRepo interface {
	// ========== 公开 API 使用 (short_token based) ==========

	// GetByShortToken 通过 short_token 获取媒体
	// 用于: GET /api/v1/medias/{short_token}
	GetByShortToken(context.Context, string) (*types.Media, error)

	// ResolveToID 将 short_token 解析为内部 ID
	// 用于: 后续操作需要 ID 时（如点赞计数）
	ResolveToID(context.Context, string) (string, error)

	// ========== 内部/Admin API 使用 (ID based) ==========

	// GetByID 通过 UUID 获取媒体完整信息
	// 用于: GET /api/v1/admin/medias/:id
	GetByID(context.Context, string) (*types.Media, error)

	// ========== 原有方法（保持兼容）==========
	Get(context.Context, string, ...*MediaQueryOption) (*types.Media, error)
	// GetWithEntity returns a single media with its loaded ent entity (including Edges).
	// This avoids extra queries when edges (user, category) are needed by the server layer.
	GetWithEntity(context.Context, string, ...*MediaQueryOption) (*entity.Media, *types.Media, error)
	List(context.Context, ...*MediaQueryOption) ([]*types.Media, int32, error)
	// ListWithEntities returns media list with loaded ent entities (including Edges)
	// This avoids N+1 queries when edges (user, category) are needed by the server layer.
	ListWithEntities(context.Context, ...*MediaQueryOption) ([]*entity.Media, []*types.Media, int32, error)
	Create(context.Context, *types.Media, ...*MediaCreateOption) (*types.Media, error)
	CreateWithEntity(context.Context, *types.Media) (*entity.Media, *types.Media, error)
	Update(context.Context, *types.Media, ...*MediaUpdateOption) (*types.Media, error)
	Delete(context.Context, string) error

	// Category operations
	ListCategories(context.Context, ...*CategoryQueryOption) ([]*types.Category, int32, error)
	GetCategory(context.Context, string) (*types.Category, error)

	// Increment views
	IncrementViewCount(context.Context, string) (int64, error)
	UpdateCommentCount(context.Context, string, int) error
	UpdateLikeCount(context.Context, string, int) error
	UpdateDislikeCount(context.Context, string, int) error
	UpdateFavoriteCount(context.Context, string, int) error
	ResetStaleProcessing(context.Context) (int, error)
	CountByEncodingStatus(context.Context) (*StatusCounts, error)
	ListFilteredByEncodingStatus(context.Context, []string, int, int) ([]*types.Media, int, error)

	UpdateSpriteFields(ctx context.Context, mediaID string, spriteStatus string, spritePath string, vttPath string) error
	UpdateThumbnailFields(ctx context.Context, mediaID string, thumbnail string, thumbnailTime float64) error
	UpdatePreviewFilePath(ctx context.Context, mediaID string, previewFilePath string) error
	UpdateDimensions(ctx context.Context, mediaID string, width, height int) error

	// ========== Entity-level data access (for handler -> biz migration) ==========

	// GetEntityByID returns the raw entity.Media by ID for accessing internal fields
	// (SpriteStatus, VttPath, SpritePath, ThumbnailTime, etc.) not exposed in types.Media.
	GetEntityByID(ctx context.Context, id string) (*entity.Media, error)

	// GetEntityByShortToken returns the raw entity.Media by short_token for accessing
	// internal fields not exposed in types.Media.
	GetEntityByShortToken(ctx context.Context, shortToken string) (*entity.Media, error)
}

// MediaQueryOption specifies options for querying media.
type MediaQueryOption struct {
	repo.QueryOption
	Type         *int32
	UserID       *string
	CategoryID   *int64
	Status       *int32
	State        string
	MediaType    string
	Featured     *bool
	OrderBy      string
	Descending   bool
	Tags         []string
	Listable     *bool
	ReviewStatus *string
	Privacy      *int32
	AdminMode    bool
}

func ptrBool(v bool) *bool       { return &v }
func ptrString(v string) *string { return &v }

// MediaCreateOption specifies options for creating media.
type MediaCreateOption struct{}

// MediaUpdateOption specifies options for updating media.
type MediaUpdateOption struct {
	repo.UpdateOption
}

// CategoryQueryOption specifies options for querying categories.
type CategoryQueryOption struct {
	repo.QueryOption
	ParentID *string
}

// ListMediasRequestToQueryOption converts an API request to a query option object.
func ListMediasRequestToQueryOption(req *media.ListMediasRequest) *MediaQueryOption {
	if req == nil {
		return &MediaQueryOption{}
	}
	opts := &MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     req.Page,
			PageSize: req.PageSize,
			Keyword:  req.Keyword,
		},
	}
	if req.Type != nil {
		opts.Type = req.Type
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	if req.UserId != nil {
		opts.UserID = req.UserId
	}
	if req.CategoryId != nil {
		opts.CategoryID = req.CategoryId
	}
	return opts
}
