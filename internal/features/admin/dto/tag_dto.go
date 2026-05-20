/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Tag DTO for frontend-compatible API responses and internal data transfer.
 *
 * B087-R2 Fix: entity.Tag uses "title"/"media_count"/"ACTIVE",
 * but frontend expects "name"/"count"/"active".
 * This DTO provides the mapping layer.
 */

package dto

import (
	"fmt"
	"strings"
	"time"
)

// TagStatusType mirrors entity/tag.Status without importing the entity package.
type TagStatusType string

const (
	TagStatusActive   TagStatusType = "ACTIVE"
	TagStatusInactive TagStatusType = "INACTIVE"
)

// TagDTO is the data transfer object for tags, isolating biz/service from entity.
type TagDTO struct {
	ID                int               `json:"id,omitempty"`
	Title             string            `json:"title,omitempty"`
	Slug              string            `json:"slug,omitempty"`
	MediaCount        int               `json:"media_count,omitempty"`
	ChannelCount      int               `json:"channel_count,omitempty"`
	ListingsThumbnail string            `json:"listings_thumbnail,omitempty"`
	Status            TagStatusType     `json:"status,omitempty"`
	Description       string            `json:"description,omitempty"`
	TitleI18n         map[string]string `json:"title_i18n,omitempty"`
	DescriptionI18n   map[string]string `json:"description_i18n,omitempty"`
	Color             string            `json:"color,omitempty"`
	CreateTime        time.Time         `json:"create_time,omitempty"`
	UpdateTime        time.Time         `json:"update_time,omitempty"`
}

// TagResponse is the frontend-compatible DTO for tag API responses.
// Field names match the frontend Tag interface in admin-tags.ts.
type TagResponse struct {
	ID                string `json:"id"`                          // int -> string
	Name              string `json:"name"`                        // TagDTO.Title -> name
	Slug              string `json:"slug,omitempty"`              // pass through
	Count             int    `json:"count"`                       // TagDTO.MediaCount -> count
	ListingsThumbnail string `json:"listings_thumbnail,omitempty"` // pass through
	Status            string `json:"status"`                      // ACTIVE -> active
	Description       string `json:"description,omitempty"`       // pass through
	Color             string `json:"color,omitempty"`             // pass through
	CreateTime        string `json:"create_time,omitempty"`       // time.Time -> ISO8601 string
	UpdateTime        string `json:"update_time,omitempty"`       // time.Time -> ISO8601 string
}

// ToTagResponse converts a TagDTO to a frontend-compatible TagResponse.
// Returns nil if the input is nil.
func ToTagResponse(t *TagDTO) *TagResponse {
	if t == nil {
		return nil
	}

	return &TagResponse{
		ID:                fmt.Sprintf("%d", t.ID),
		Name:              t.Title,
		Slug:              t.Slug,
		Count:             t.MediaCount,
		ListingsThumbnail: t.ListingsThumbnail,
		Status:            StatusToString(t.Status),
		Description:       t.Description,
		Color:             t.Color,
		CreateTime:        t.CreateTime.Format("2006-01-02T15:04:05Z07:00"),
		UpdateTime:        t.UpdateTime.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToTagResponseList converts a slice of TagDTO to TagResponse slice.
func ToTagResponseList(tags []*TagDTO) []*TagResponse {
	if len(tags) == 0 {
		return []*TagResponse{}
	}
	result := make([]*TagResponse, len(tags))
	for i, t := range tags {
		result[i] = ToTagResponse(t)
	}
	return result
}

// StatusToString converts a TagStatusType to a lowercase string.
func StatusToString(s TagStatusType) string {
	switch s {
	case TagStatusActive:
		return "active"
	case TagStatusInactive:
		return "inactive"
	default:
		return strings.ToLower(string(s))
	}
}

// ParseTagStatus converts a frontend status string to a TagStatusType.
// Accepts both lowercase ("active") and uppercase ("ACTIVE") formats.
// Defaults to ACTIVE for empty or unrecognized values.
func ParseTagStatus(s string) TagStatusType {
	switch strings.ToLower(s) {
	case "active":
		return TagStatusActive
	case "inactive":
		return TagStatusInactive
	default:
		return TagStatusActive
	}
}
