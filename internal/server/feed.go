/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
)

// FeedHandler handles feed-related HTTP endpoints
type FeedHandler struct {
	db *entity.Client
}

// NewFeedHandler creates a new FeedHandler
func NewFeedHandler(db *entity.Client) *FeedHandler {
	return &FeedHandler{db: db}
}

// FeedResponse represents the feed response structure
type FeedResponse struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	Sections   []Section   `json:"sections"`
}

// Section represents a feed section
type Section struct {
	Title string      `json:"title"`
	Type  string      `json:"type"`
	Items []MediaItem `json:"items"`
}

// MediaItem represents a media item in feed
type MediaItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    int    `json:"duration"`
	ViewCount   int    `json:"view_count"`
	URL         string `json:"url"`
}

// GetFeed godoc: GET /api/v1/feed
func (h *FeedHandler) GetFeed(c *gin.Context) {
	// 获取媒体列表作为 feed
	medias, err := h.db.Media.Query().
		Limit(20).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
		return
	}

	// 如果没有媒体，返回空sections
	if len(medias) == 0 {
		c.JSON(http.StatusOK, FeedResponse{
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
			Sections:   []Section{},
		})
		return
	}

	// 转换媒体为 MediaItem
	items := make([]MediaItem, len(medias))
	for i, m := range medias {
		items[i] = MediaItem{
			ID:          int(m.ID),
			Title:       m.Title,
			Description: m.Description,
			Thumbnail:  m.Thumbnail,
			Duration:   m.Duration,
			ViewCount:  int(m.ViewCount),
			URL:        "/v/" + string(rune(m.ID)),
		}
	}

	c.JSON(http.StatusOK, FeedResponse{
		Page:       1,
		PageSize:   20,
		TotalCount: len(medias),
		Sections: []Section{
			{
				Title: "Recommended for You",
				Type:  "recommended",
				Items: items,
			},
		},
	})
}