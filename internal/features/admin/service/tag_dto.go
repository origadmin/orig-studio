/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Tag DTO for frontend-compatible API responses.
 *
 * B087-R2 Fix: entity.Tag uses "title"/"media_count"/"ACTIVE",
 * but frontend expects "name"/"count"/"active".
 * This DTO provides the mapping layer.
 */

package service

import (
	"fmt"
	"strings"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/tag"
)

// TagResponse is the frontend-compatible DTO for tag API responses.
// Field names match the frontend Tag interface in admin-tags.ts.
type TagResponse struct {
	ID                string `json:"id"`                  // int -> string
	Name              string `json:"name"`                // entity.Title -> name
	Slug              string `json:"slug,omitempty"`       // pass through
	Count             int    `json:"count"`                // entity.MediaCount -> count
	ListingsThumbnail string `json:"listings_thumbnail,omitempty"` // pass through
	Status            string `json:"status"`               // ACTIVE -> active
	Description       string `json:"description,omitempty"` // pass through
	Color             string `json:"color,omitempty"`       // pass through
	CreateTime        string `json:"create_time,omitempty"` // time.Time -> ISO8601 string
	UpdateTime        string `json:"update_time,omitempty"` // time.Time -> ISO8601 string
}

// ToTagResponse converts an entity.Tag to a frontend-compatible TagResponse.
// Returns nil if the input is nil.
func ToTagResponse(t *entity.Tag) *TagResponse {
	if t == nil {
		return nil
	}

	return &TagResponse{
		ID:                fmt.Sprintf("%d", t.ID),
		Name:              t.Title,
		Slug:              t.Slug,
		Count:             t.MediaCount,
		ListingsThumbnail: t.ListingsThumbnail,
		Status:            statusToString(t.Status),
		Description:       t.Description,
		Color:             t.Color,
		CreateTime:        t.CreateTime.Format("2006-01-02T15:04:05Z07:00"),
		UpdateTime:        t.UpdateTime.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToTagResponseList converts a slice of entity.Tag to TagResponse slice.
func ToTagResponseList(tags []*entity.Tag) []*TagResponse {
	if len(tags) == 0 {
		return []*TagResponse{}
	}
	result := make([]*TagResponse, len(tags))
	for i, t := range tags {
		result[i] = ToTagResponse(t)
	}
	return result
}

// statusToString converts a tag.Status enum value to a lowercase string.
func statusToString(s tag.Status) string {
	switch s {
	case tag.StatusACTIVE:
		return "active"
	case tag.StatusINACTIVE:
		return "inactive"
	default:
		return strings.ToLower(string(s))
	}
}

// ParseTagStatus converts a frontend status string to a tag.Status enum value.
// Accepts both lowercase ("active") and uppercase ("ACTIVE") formats.
// Defaults to ACTIVE for empty or unrecognized values.
func ParseTagStatus(s string) tag.Status {
	switch strings.ToLower(s) {
	case "active":
		return tag.StatusACTIVE
	case "inactive":
		return tag.StatusINACTIVE
	default:
		return tag.DefaultStatus // ACTIVE
	}
}
