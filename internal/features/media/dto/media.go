/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"
	"time"

	"origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/api/gen/v1/types"
	repotypes "origadmin/application/origstudio/internal/domain/types"
)

// MediaRepo is a Media repository interface.
type MediaRepo interface {
	// ========== Public API (short_token based) ==========

	// GetByShortToken gets media by short_token
	GetByShortToken(context.Context, string) (*types.Media, error)

	// ResolveToID resolves short_token to internal ID
	ResolveToID(context.Context, string) (string, error)

	// ========== Internal/Admin API (ID based) ==========

	// GetByID gets full media info by UUID
	GetByID(context.Context, string) (*types.Media, error)

	// ========== Existing methods (backwards compatible) ==========
	Get(context.Context, string, ...*MediaQueryOption) (*types.Media, error)
	// GetWithEntity returns a single media with its entity-level data (including edges).
	GetWithEntity(context.Context, string, ...*MediaQueryOption) (*MediaEntityDTO, *types.Media, error)
	List(context.Context, ...*MediaQueryOption) ([]*types.Media, int32, error)
	// ListWithEntities returns media list with entity-level data (including edges).
	ListWithEntities(context.Context, ...*MediaQueryOption) ([]*MediaEntityDTO, []*types.Media, int32, error)
	Create(context.Context, *types.Media, ...*MediaCreateOption) (*types.Media, error)
	// CreateWithEntity creates media and returns entity-level data.
	CreateWithEntity(context.Context, *types.Media) (*MediaEntityDTO, *types.Media, error)
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

	// ========== Entity-level data access ==========

	// GetEntityByID returns the MediaEntityDTO by ID for accessing internal fields
	// (SpriteStatus, VttPath, SpritePath, ThumbnailTime, etc.) not exposed in types.Media.
	GetEntityByID(ctx context.Context, id string) (*MediaEntityDTO, error)

	// GetEntityByShortToken returns the MediaEntityDTO by short_token for accessing
	// internal fields not exposed in types.Media.
	GetEntityByShortToken(ctx context.Context, shortToken string) (*MediaEntityDTO, error)

	// ListTempMediaBefore returns media records whose URL starts with "temp/" and
	// whose create_time is before the given cutoff.
	ListTempMediaBefore(ctx context.Context, cutoff time.Time) ([]*types.Media, error)
}

// MediaQueryOption specifies options for querying media.
type MediaQueryOption struct {
	repotypes.QueryOption
	Type         *int32
	UserID       *string
	CategoryID   *int64
	CategoryIDs  []int64
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
	repotypes.UpdateOption
}

// CategoryQueryOption specifies options for querying categories.
type CategoryQueryOption struct {
	repotypes.QueryOption
	ParentID *string
}

// ListMediasRequestToQueryOption converts an API request to a query option object.
// Pagination parameters are automatically normalized.
func ListMediasRequestToQueryOption(req *media.ListMediasRequest) *MediaQueryOption {
	if req == nil {
		return &MediaQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     1,
				PageSize: 20,
			},
		}
	}
	page, pageSize := repotypes.NormalizePagination(int(req.Page), int(req.PageSize))
	opts := &MediaQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
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
