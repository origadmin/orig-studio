/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Content entity DTOs isolate biz/ and service/ layers from internal/dal/entity imports.
 */

package dto

import "time"

// ==================== Portal DTOs ====================

// PortalNavItemDTO mirrors entity.PortalNavItem without importing the entity package.
type PortalNavItemDTO struct {
	ID          string            `json:"id,omitempty"`
	Type        string            `json:"type,omitempty"`
	Label       string            `json:"label,omitempty"`
	LabelI18n   map[string]string `json:"label_i18n,omitempty"`
	URL         string            `json:"url,omitempty"`
	TargetType  string            `json:"target_type,omitempty"`
	TargetID    string            `json:"target_id,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Color       string            `json:"color,omitempty"`
	Sequence    int               `json:"sequence,omitempty"`
	ParentID    string            `json:"parent_id,omitempty"`
	IsVisible   bool              `json:"is_visible,omitempty"`
	OpenNewTab  bool              `json:"open_new_tab,omitempty"`
	CSSClass    string            `json:"css_class,omitempty"`
	// Legacy compatibility fields
	Target    string    `json:"target,omitempty"`
	SortOrder int       `json:"sort_order,omitempty"`
	IsActive  bool      `json:"is_active,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
	UpdateTime time.Time `json:"update_time,omitempty"`
}

// PortalBannerDTO mirrors entity.PortalBanner without importing the entity package.
type PortalBannerDTO struct {
	ID                string            `json:"id,omitempty"`
	Title             string            `json:"title,omitempty"`
	TitleI18n         map[string]string `json:"title_i18n,omitempty"`
	Subtitle          string            `json:"subtitle,omitempty"`
	SubtitleI18n      map[string]string `json:"subtitle_i18n,omitempty"`
	BadgeText         string            `json:"badge_text,omitempty"`
	ImageURL          string            `json:"image_url,omitempty"`
	ImageMobileURL    string            `json:"image_mobile_url,omitempty"`
	BgColorStart      string            `json:"bg_color_start,omitempty"`
	BgColorEnd        string            `json:"bg_color_end,omitempty"`
	BgOverlayOpacity  float64           `json:"bg_overlay_opacity,omitempty"`
	PrimaryBtnText    string            `json:"primary_btn_text,omitempty"`
	PrimaryBtnURL     string            `json:"primary_btn_url,omitempty"`
	SecondaryBtnText  string            `json:"secondary_btn_text,omitempty"`
	SecondaryBtnURL   string            `json:"secondary_btn_url,omitempty"`
	Sequence          int               `json:"sequence,omitempty"`
	IsActive          bool              `json:"is_active,omitempty"`
	StartAt           time.Time         `json:"start_at,omitempty"`
	EndAt             time.Time         `json:"end_at,omitempty"`
	AutoSlideInterval int               `json:"auto_slide_interval,omitempty"`
	// Legacy compatibility fields
	LinkURL    string    `json:"link_url,omitempty"`
	LinkTarget string    `json:"link_target,omitempty"`
	SortOrder  int       `json:"sort_order,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
	UpdateTime time.Time `json:"update_time,omitempty"`
}

// PortalCustomPageDTO mirrors entity.PortalCustomPage without importing the entity package.
type PortalCustomPageDTO struct {
	ID             string    `json:"id,omitempty"`
	Title          string    `json:"title,omitempty"`
	Slug           string    `json:"slug,omitempty"`
	Type           string    `json:"type,omitempty"`
	ContentFormat  string    `json:"content_format,omitempty"`
	Content        string    `json:"content,omitempty"`
	Layout         string    `json:"layout,omitempty"`
	IsPublished    bool      `json:"is_published,omitempty"`
	PublishedAt    time.Time `json:"published_at,omitempty"`
	SeoTitle       string    `json:"seo_title,omitempty"`
	SeoDescription string    `json:"seo_description,omitempty"`
	FeaturedImage  string    `json:"featured_image,omitempty"`
	ViewCount      int64     `json:"view_count,omitempty"`
	// Legacy compatibility fields
	SortOrder  int       `json:"sort_order,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
	UpdateTime time.Time `json:"update_time,omitempty"`
}

// ==================== Ad DTOs ====================

// AdPlacementDTO mirrors entity.AdPlacement without importing the entity package.
type AdPlacementDTO struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	MaxAds      int    `json:"max_ads,omitempty"`
	IsActive    bool   `json:"is_active,omitempty"`
	Sequence    int    `json:"sequence,omitempty"`
	// Legacy compatibility fields
	CreateTime  time.Time `json:"create_time,omitempty"`
	UpdateTime  time.Time `json:"update_time,omitempty"`
}

// AdDTO mirrors entity.Ad without importing the entity package.
type AdDTO struct {
	ID             string            `json:"id,omitempty"`
	PlacementID    string            `json:"placement_id,omitempty"`
	Title          string            `json:"title,omitempty"`
	TitleI18n      map[string]string `json:"title_i18n,omitempty"`
	ImageURL       string            `json:"image_url,omitempty"`
	ImageMobileURL string            `json:"image_mobile_url,omitempty"`
	LinkURL        string            `json:"link_url,omitempty"`
	LinkTarget     string            `json:"link_target,omitempty"`
	BadgeText      string            `json:"badge_text,omitempty"`
	Priority       int               `json:"priority,omitempty"`
	IsActive       bool              `json:"is_active,omitempty"`
	StartAt        time.Time         `json:"start_at,omitempty"`
	EndAt          time.Time         `json:"end_at,omitempty"`
	Impressions    int64             `json:"impressions,omitempty"`
	Clicks         int64             `json:"clicks,omitempty"`
	// Legacy compatibility fields
	SortOrder  int       `json:"sort_order,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
	UpdateTime time.Time `json:"update_time,omitempty"`
}

// AdClickLogDTO mirrors entity.AdClickLog without importing the entity package.
type AdClickLogDTO struct {
	ID          string `json:"id,omitempty"`
	AdID        string `json:"ad_id,omitempty"`
	PlacementID string `json:"placement_id,omitempty"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Referer     string `json:"referer,omitempty"`
	// Legacy compatibility fields
	IPAddress  string    `json:"ip_address,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

// ==================== Comment DTOs ====================

// CommentUserDTO carries user data for comment edges without importing entity.
type CommentUserDTO struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Slug     string `json:"slug,omitempty"`
}

// CommentDTO mirrors entity.Comment without importing the entity package.
type CommentDTO struct {
	ID         string    `json:"id,omitempty"`
	Content    string    `json:"content,omitempty"`
	MediaID    string    `json:"media_id,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	ParentID   string    `json:"parent_id,omitempty"`
	RootID     string    `json:"root_id,omitempty"`
	Status     string    `json:"status,omitempty"`
	LikeCount  int       `json:"like_count,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
	UpdateTime time.Time `json:"update_time,omitempty"`
	// Edge data
	UserName   string `json:"user_name,omitempty"`
	UserAvatar string `json:"user_avatar,omitempty"`
}
