/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"

	"origadmin/application/origcms/api/v1/services/media"
	"origadmin/application/origcms/api/v1/services/types"

	"origadmin/application/origcms/internal/helpers/repo"
)

// MediaRepo is a Media repository interface.
type MediaRepo interface {
	Get(context.Context, int64, ...*MediaQueryOption) (*types.Media, error)
	List(context.Context, ...*MediaQueryOption) ([]*types.Media, int32, error)
	Create(context.Context, *types.Media, ...*MediaCreateOption) (*types.Media, error)
	Update(context.Context, *types.Media, ...*MediaUpdateOption) (*types.Media, error)
	Delete(context.Context, int64) error

	// Category operations
	ListCategories(context.Context, ...*CategoryQueryOption) ([]*types.Category, int32, error)
	GetCategory(context.Context, int64) (*types.Category, error)

	// Increment views
	IncrementViewCount(context.Context, int64) (int64, error)
}

// MediaQueryOption specifies options for querying media.
type MediaQueryOption struct {
	repo.QueryOption
	Type       *int32
	UserID     *int64
	CategoryID *int64
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
	ParentID *int64
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
