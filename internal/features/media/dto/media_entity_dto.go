/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * MediaEntityDTO provides entity-level fields not available in proto types.
 * This isolates biz/ and service/ layers from internal/dal/entity imports.
 */

package dto

import "time"

// MediaEntityType mirrors the media type field without importing the entity package.
type MediaEntityType string

const (
	MediaEntityTypeVideo MediaEntityType = "video"
	MediaEntityTypeImage MediaEntityType = "image"
	MediaEntityTypeAudio MediaEntityType = "audio"
)

// MediaEntityDTO carries entity-level media fields (e.g., Edges, SpriteStatus)
// that are not present in the proto-generated types.Media.
// The biz/ and service/ layers use this instead of importing entity.Media directly.
type MediaEntityDTO struct {
	ID                  string          `json:"id,omitempty"`
	Title               string          `json:"title,omitempty"`
	Type                MediaEntityType `json:"type,omitempty"`
	URL                 string          `json:"url,omitempty"`
	ShortToken          string          `json:"short_token,omitempty"`
	Status              string          `json:"status,omitempty"`
	ReviewStatus        string          `json:"review_status,omitempty"`
	SpriteStatus        string          `json:"sprite_status,omitempty"`
	SpritePath          string          `json:"sprite_path,omitempty"`
	VttPath             string          `json:"vtt_path,omitempty"`
	Thumbnail           string          `json:"thumbnail,omitempty"`
	ThumbnailTime       float64         `json:"thumbnail_time,omitempty"`
	PreviewFilePath     string          `json:"preview_file_path,omitempty"`
	Width               int             `json:"width,omitempty"`
	Height              int             `json:"height,omitempty"`
	Duration            float64         `json:"duration,omitempty"`
	ViewCount           int64           `json:"view_count,omitempty"`
	LikeCount           int64           `json:"like_count,omitempty"`
	DislikeCount        int64           `json:"dislike_count,omitempty"`
	FavoriteCount       int64           `json:"favorite_count,omitempty"`
	CommentCount        int64           `json:"comment_count,omitempty"`
	CreateTime          time.Time       `json:"create_time,omitempty"`
	UpdateTime          time.Time       `json:"update_time,omitempty"`

	// Edge data - populated by dal/ layer when edges are loaded
	UserID       string `json:"user_id,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserNickname string `json:"user_nickname,omitempty"`
	UserAvatar   string `json:"user_avatar,omitempty"`
	UserSlug     string `json:"user_slug,omitempty"`
	CategoryID   int    `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}
