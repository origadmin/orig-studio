/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package repo provides common repository utilities.
package repo

import "google.golang.org/protobuf/types/known/emptypb"

// QueryOption provides common query options.
type QueryOption struct {
	Page          int32
	PageSize      int32
	Keyword       string
	Status        *int32
	PageToken     string
	PagingMode    string // "offset" | "cursor" | "none"
	OnlyCount     bool
	NoPaging      bool
	OrderBy       []string
	SortFromToken bool
}

// UpdateOption provides common update options.
type UpdateOption struct {
	// For future use
}

// Empty returns an empty proto message for delete responses.
func Empty() *emptypb.Empty {
	return &emptypb.Empty{}
}

// PaginatingRequest defines the contract for any request that supports offset-based pagination.
type PaginatingRequest interface {
	GetPage() int32
	GetPageSize() int32
}

// TokenPaginatingRequest defines the contract for any request that supports token-based pagination.
type TokenPaginatingRequest interface {
	GetPageToken() string
}

// KeywordRequest defines the contract for any request that supports keyword-based search.
type KeywordRequest interface {
	GetKeyword() string
}

// StatusRequest defines the contract for any request that supports status filtering.
type StatusRequest interface {
	GetStatus() *int32
}

// QueryOptionFromRequest extracts query options from a request using interface assertions.
// It automatically normalizes pagination parameters.
func QueryOptionFromRequest(req interface{}) QueryOption {
	opt := QueryOption{}

	if r, ok := req.(PaginatingRequest); ok {
		opt.Page = r.GetPage()
		opt.PageSize = r.GetPageSize()
	}

	if r, ok := req.(TokenPaginatingRequest); ok {
		opt.PageToken = r.GetPageToken()
	}

	if r, ok := req.(KeywordRequest); ok {
		opt.Keyword = r.GetKeyword()
	}

	if r, ok := req.(StatusRequest); ok {
		opt.Status = r.GetStatus()
	}

	// Normalize pagination parameters
	normalized := NormalizeQueryOption(&opt)
	return *normalized
}
