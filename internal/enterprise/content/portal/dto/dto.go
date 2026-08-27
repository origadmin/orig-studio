package dto

import "time"

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
	Target      string            `json:"target,omitempty"`
	SortOrder   int               `json:"sort_order,omitempty"`
	IsActive    bool              `json:"is_active,omitempty"`
	CreateTime  time.Time         `json:"create_time,omitempty"`
	UpdateTime  time.Time         `json:"update_time,omitempty"`
}

type PortalBannerDTO struct {
	ID                string            `json:"id,omitempty"`
	Title             string            `json:"title,omitempty"`
	TitleI18n         map[string]string `json:"title_i18n,omitempty"`
	Subtitle          string            `json:"subtitle,omitempty"`
	SubtitleI18n      map[string]string `json:"subtitle_i18n,omitempty"`
	BadgeText         string            `json:"badge_text,omitempty"`
	ImageURL          string            `json:"image_url,omitempty"`
	ImageMobileURL    string            `json:"image_mobile_url,omitempty"`
	VideoURL          string            `json:"video_url,omitempty"`
	BgColorStart      string            `json:"bg_color_start,omitempty"`
	BgColorEnd        string            `json:"bg_color_end,omitempty"`
	BgOverlayOpacity  float64           `json:"bg_overlay_opacity,omitempty"`
	PrimaryBtnText    string            `json:"primary_btn_text,omitempty"`
	PrimaryBtnURL     string            `json:"primary_btn_url,omitempty"`
	SecondaryBtnText  string            `json:"secondary_btn_text,omitempty"`
	SecondaryBtnURL   string            `json:"secondary_btn_url,omitempty"`
	Sequence          int               `json:"sequence,omitempty"`
	IsActive          bool              `json:"is_active,omitempty"`
	StartAt           *time.Time        `json:"start_at,omitempty"`
	EndAt             *time.Time        `json:"end_at,omitempty"`
	ClearEndAt        bool              `json:"-"`
	ClearStartAt      bool              `json:"-"`
	AutoSlideInterval int               `json:"auto_slide_interval,omitempty"`
	Type              string            `json:"type,omitempty"`
	Count             int               `json:"count,omitempty"`
	CategoryID        string            `json:"category_id,omitempty"`
	DisplayMode       string            `json:"display_mode,omitempty"`
	LinkURL           string            `json:"link_url,omitempty"`
	LinkTarget        string            `json:"link_target,omitempty"`
	SortOrder         int               `json:"sort_order,omitempty"`
	StartTime         *time.Time        `json:"start_time,omitempty"`
	EndTime           *time.Time        `json:"end_time,omitempty"`
	CreateTime        *time.Time        `json:"create_time,omitempty"`
	UpdateTime        *time.Time        `json:"update_time,omitempty"`
}

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
	SortOrder      int       `json:"sort_order,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty"`
	UpdateTime     time.Time `json:"update_time,omitempty"`
}