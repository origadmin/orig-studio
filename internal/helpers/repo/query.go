/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package repo provides common repository utilities.
package repo

import "google.golang.org/protobuf/types/known/emptypb"

// QueryOption provides common query options.
type QueryOption struct {
	Page     int32
	PageSize int32
	Keyword  string
	Status   *int32
}

// UpdateOption provides common update options.
type UpdateOption struct {
	// For future use
}

// Empty returns an empty proto message for delete responses.
func Empty() *emptypb.Empty {
	return &emptypb.Empty{}
}

// QueryOptionFromRequest extracts query options from a request.
func QueryOptionFromRequest(req interface{}) QueryOption {
	// Simplified version - in production, use reflection or generated code
	return QueryOption{
		Page:     1,
		PageSize: 10,
	}
}
