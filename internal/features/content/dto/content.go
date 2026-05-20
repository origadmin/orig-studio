/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package dto provides data transfer objects for the content feature module.
package dto

// FeedResponse is the DTO for feed API responses.
type FeedResponse struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalCount int       `json:"total_count"`
	Sections   []Section `json:"sections"`
}

// Section is the DTO for a feed section.
type Section struct {
	Title string      `json:"title"`
	Type  string      `json:"type"`
	Items interface{} `json:"items"` // []*biz.MediaInfo - uses interface{} to avoid circular import
}

// CommentListItem is the DTO for a comment in admin list responses.
// Field names align with the frontend admin Comments page expectations (B087).
type CommentListItem struct {
	ID                string            `json:"id"`
	Content           string            `json:"content"`
	Status            string            `json:"status"`
	MediaID           string            `json:"media_id"`
	UserID            string            `json:"user_id"`
	Username          string            `json:"username,omitempty"`
	Avatar            string            `json:"avatar,omitempty"`
	LikeCount         int               `json:"like_count"`
	ReplyCount        int               `json:"reply_count"`
	ReportCount       int               `json:"report_count"`
	IsSpam            bool              `json:"is_spam"`
	CreateTime        string            `json:"create_time"`
	Media             *CommentMediaItem `json:"media,omitempty"`
	ModeratedBy       string            `json:"moderated_by,omitempty"`
	ModeratedAt       string            `json:"moderated_at,omitempty"`
	ParentID          string            `json:"parent_id,omitempty"`
	Depth             int               `json:"depth"`
	HasReplies        bool              `json:"has_replies"`
	Children          []CommentListItem `json:"children,omitempty"`
	HasPendingReports bool              `json:"has_pending_reports"`
}

// CommentMediaItem is the nested media object in admin comment list responses.
type CommentMediaItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// CommentStatsDTO is the DTO for comment statistics.
type CommentStatsDTO struct {
	Pending         int `json:"pending"`
	Approved        int `json:"approved"`
	Rejected        int `json:"rejected"`
	Blocked         int `json:"blocked"`
	Total           int `json:"total"`
	ReportedPending int `json:"reported_pending"`
}

// ModerationResultDTO is the DTO for moderation action results.
type ModerationResultDTO struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	ModeratedBy string `json:"moderated_by"`
	ModeratedAt string `json:"moderated_at"`
	ReportCount int    `json:"report_count,omitempty"`
}

// BatchResultDTO is the DTO for batch moderation results.
type BatchResultDTO struct {
	UpdatedCount int    `json:"updated_count"`
	SkippedCount int    `json:"skipped_count"`
	Message      string `json:"message"`
}

// CommentReportDTO is the DTO for a comment report.
type CommentReportDTO struct {
	ID          string `json:"id"`
	CommentID   string `json:"comment_id"`
	ReporterID  string `json:"reporter_id"`
	Reason      string `json:"reason"`
	CreateTime   string `json:"create_time"`
	Description string `json:"description,omitempty"`
	Username    string `json:"username,omitempty"`
	Status      string `json:"status"`
}

// CommentReportsResultDTO is the DTO for comment reports response.
type CommentReportsResultDTO struct {
	CommentID   string             `json:"comment_id"`
	ReportCount int                `json:"report_count"`
	Reports     []CommentReportDTO `json:"reports"`
}

// ReportResultDTO is the DTO for report submission result.
type ReportResultDTO struct {
	Message     string `json:"message"`
	ReportCount int    `json:"report_count"`
	Status      string `json:"status"`
}

// DismissReportsResultDTO is the DTO for dismiss reports result.
type DismissReportsResultDTO struct {
	CommentID      string `json:"comment_id"`
	DismissedCount int    `json:"dismissed_count"`
	ReportCount    int    `json:"report_count"`
	Message        string `json:"message"`
}

// MediaReportRequest is the DTO for media report requests.
type MediaReportRequest struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// MediaReportResultDTO is the DTO for media report results.
type MediaReportResultDTO struct {
	Message     string `json:"message"`
	ReportCount int    `json:"report_count"`
	Status      string `json:"status"`
}

// SocialShareLinks is the DTO for social media share links.
type SocialShareLinks struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	Twitter  string `json:"twitter"`
	Facebook string `json:"facebook"`
	LinkedIn string `json:"linkedin"`
	WhatsApp string `json:"whatsapp"`
	Telegram string `json:"telegram"`
}
