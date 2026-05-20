/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Tag DTO helpers for the service layer.
 * Re-exports dto package functions to avoid entity imports in service.
 */

package service

import (
	"origadmin/application/origstudio/internal/features/admin/dto"
)

// ToTagResponse converts a TagDTO to a frontend-compatible TagResponse.
func ToTagResponse(t *dto.TagDTO) *dto.TagResponse {
	return dto.ToTagResponse(t)
}

// ToTagResponseList converts a slice of TagDTO to TagResponse slice.
func ToTagResponseList(tags []*dto.TagDTO) []*dto.TagResponse {
	return dto.ToTagResponseList(tags)
}

// ParseTagStatus converts a frontend status string to a TagStatusType.
func ParseTagStatus(s string) dto.TagStatusType {
	return dto.ParseTagStatus(s)
}
