/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"

	"origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/helpers/repo"
)

// MediaRepo is a Media repository interface.
type MediaRepo interface {
	Get(context.Context, string, ...*MediaQueryOption) (*types.Media, error)
	List(context.Context, ...*MediaQueryOption) ([]*types.Media, int32, error)
	Create(context.Context, *types.Media, ...*MediaCreateOption) (*types.Media, error)
	Update(context.Context, *types.Media, ...*MediaUpdateOption) (*types.Media, error)
	Delete(context.Context, string) error

	// Category operations
	ListCategories(context.Context, ...*CategoryQueryOption) ([]*types.Category, int32, error)
	GetCategory(context.Context, string) (*types.Category, error)

	// Increment views
	IncrementViewCount(context.Context, string) (int64, error)
}

// MediaQueryOption specifies options for querying media.
type MediaQueryOption struct {
	repo.QueryOption
	Type       *int32
	UserID     *string
	CategoryID *string
	Status     *int32
	// Added for Gin handler parity
	State      string
	MediaType  string
	Featured   *bool
	OrderBy    string
	Descending bool
	Tags       []string
}

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
