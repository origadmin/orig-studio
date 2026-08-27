package dto

import "time"

type AdPlacementDTO struct {
	ID           string    `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	Slug         string    `json:"slug,omitempty"`
	Type         string    `json:"type,omitempty"`
	Description  string    `json:"description,omitempty"`
	Width        int       `json:"width,omitempty"`
	Height       int       `json:"height,omitempty"`
	MaxAds       int       `json:"max_ads,omitempty"`
	IsActive     bool      `json:"is_active,omitempty"`
	Sequence     int       `json:"sequence,omitempty"`
	CreativeCount int      `json:"creative_count,omitempty"`
	CreateTime   time.Time `json:"create_time,omitempty"`
	UpdateTime   time.Time `json:"update_time,omitempty"`
}

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
	SortOrder      int               `json:"sort_order,omitempty"`
	StartTime      time.Time         `json:"start_time,omitempty"`
	EndTime        time.Time         `json:"end_time,omitempty"`
	CreateTime     time.Time         `json:"create_time,omitempty"`
	UpdateTime     time.Time         `json:"update_time,omitempty"`
}

type AdCreativeDTO struct {
	ID             string            `json:"id,omitempty"`
	Title          string            `json:"title,omitempty"`
	TitleI18n      map[string]string `json:"title_i18n,omitempty"`
	ImageURL       string            `json:"image_url,omitempty"`
	ImageMobileURL string            `json:"image_mobile_url,omitempty"`
	LinkURL        string            `json:"link_url,omitempty"`
	LinkTarget     string            `json:"link_target,omitempty"`
	BadgeText      string            `json:"badge_text,omitempty"`
	IsActive       bool              `json:"is_active,omitempty"`
	Priority       int               `json:"priority,omitempty"`
	Impressions    int64             `json:"impressions,omitempty"`
	Clicks         int64             `json:"clicks,omitempty"`
	CreateTime     time.Time         `json:"create_time,omitempty"`
	UpdateTime     time.Time         `json:"update_time,omitempty"`
}

type AdPlacementWithAdsDTO struct {
	AdPlacementDTO
	Ads       []*AdDTO         `json:"ads,omitempty"`
	Creatives []*AdCreativeDTO `json:"creatives,omitempty"`
}

type AdClickLogDTO struct {
	ID          string    `json:"id,omitempty"`
	AdID        string    `json:"ad_id,omitempty"`
	PlacementID string    `json:"placement_id,omitempty"`
	IP          string    `json:"ip,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Referer     string    `json:"referer,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	CreateTime  time.Time `json:"create_time,omitempty"`
}