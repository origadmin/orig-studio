/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	CreateCategory(context.Context, *types.Category) (*types.Category, error)
	UpdateCategory(context.Context, *types.Category) (*types.Category, error)
	DeleteCategory(context.Context, string) error

	// Tag operations
	ListTags(context.Context, ...*TagQueryOption) ([]*types.Tag, int32, error)
	GetTag(context.Context, string) (*types.Tag, error)
	CreateTag(context.Context, *types.Tag) (*types.Tag, error)
	UpdateTag(context.Context, *types.Tag) (*types.Tag, error)
	DeleteTag(context.Context, string) error

	// Increment views
	IncrementViewCount(context.Context, string) (int64, error)
	UpdateCommentCount(context.Context, string, int) error
	UpdateLikeCount(context.Context, string, int) error
	UpdateDislikeCount(context.Context, string, int) error
	UpdateFavoriteCount(context.Context, string, int) error
	UpdateReportedTimes(context.Context, string, int) error
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

	// GetDefaultChannelID returns the ID of the user's first (default) channel.
	// Used by UploadMedia to auto-bind orphan videos to a channel when none is specified.
	GetDefaultChannelID(ctx context.Context, userID string) (string, error)

	// GetChannelOwnerID returns the owner user id of a channel ("" if not found).
	// Used by UpdateMedia (BUG-105) to validate that a caller may assign media to a channel.
	GetChannelOwnerID(ctx context.Context, channelID string) (string, error)

	// UpdateMediaChannel sets or clears a media's channel assignment.
	// channelID == "" clears the assignment; otherwise sets it (BUG-105).
	UpdateMediaChannel(ctx context.Context, mediaID, channelID string) error
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
	RandomSeed   *uint32 // BUG-226: deterministic seed for ORDER BY random(id,seed); nil = not random
	CreatedAfter string // RFC3339: filter create_time >= this instant (time-range)
	Tags         []string
	Listable     *bool
	ReviewStatus *string
	Privacy      *int32
	AdminMode    bool
	OwnerView    bool // BUG-141②: viewer is the owner of req.UserId → relax gate to own active (incl. unreviewed)
	// BUG-105: channel assignment filters. ChannelUnassigned takes precedence
	// (channel_id IS NULL); otherwise ChannelID filters by exact channel.
	ChannelID        *string
	ChannelUnassigned bool
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

// TagQueryOption specifies options for querying tags.
type TagQueryOption struct {
	repotypes.QueryOption
	Status string
}

// ListMediasRequestToQueryOption converts an API request to a query option object.
// Pagination parameters are automatically normalized.
//
// BUG-226 ("为您推荐" 换一批): when order_by=random a positive uint32 seed is
// mandatory. The seed is bound as a parameter into the ORDER BY expression in the
// DAL (never interpolated into WHERE/table/column names) so the same seed yields
// the same ordering across connections. A missing or non-positive seed is rejected
// with InvalidArgument (HTTP 400) at the gateway.
func ListMediasRequestToQueryOption(req *media.ListMediasRequest) (*MediaQueryOption, error) {
	if req == nil {
		return &MediaQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     1,
				PageSize: 20,
			},
		}, nil
	}
	page, pageSize := repotypes.NormalizePagination(int(req.Page), int(req.PageSize))
	opts := &MediaQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  req.Keyword,
		},
		OrderBy:    req.OrderBy,
		Descending: req.Descending,
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
	// BUG-131: wire the previously-dropped filter fields through to the query
	// option. The backend (media_repo.go) already applies all of these.
	if len(req.Tags) > 0 {
		opts.Tags = req.Tags
	}
	if len(req.CategoryIds) > 0 {
		opts.CategoryIDs = req.CategoryIds
	}
	if req.CreatedAfter != "" {
		opts.CreatedAfter = req.CreatedAfter
	}
	if req.State != nil {
		opts.State = *req.State
	}
	if req.Featured != nil {
		opts.Featured = req.Featured
	}
	// BUG-226: deterministic shuffle seed for order_by=random.
	if req.OrderBy == "random" {
		if req.Seed == nil {
			return nil, status.Error(codes.InvalidArgument, "seed is required when order_by=random")
		}
		if *req.Seed == 0 {
			return nil, status.Error(codes.InvalidArgument, "seed must be a positive integer (1..4294967295) when order_by=random")
		}
		s := *req.Seed
		opts.RandomSeed = &s
	}
	return opts, nil
}
