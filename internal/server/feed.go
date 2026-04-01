/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
	entitymedia "origadmin/application/origcms/internal/data/entity/media"
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
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalCount int       `json:"total_count"`
	Sections   []Section `json:"sections"`
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
	ViewCount int64  `json:"view_count"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	URL         string `json:"url"`
	Type      string `json:"type"`
}

// GetFeed godoc: GET /api/v1/feed
func (h *FeedHandler) GetFeed(c *gin.Context) {
	ctx := c.Request.Context()

	// Get latest active media, ordered by created_at desc
	medias, err := h.db.Media.Query().
		Where(entitymedia.StateEQ("active")).
		Order(entity.Desc("created_at")).
		Limit(20).
		WithUser().
		WithCategory().
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
		return
	}

	// If no media, return empty sections
	if len(medias) == 0 {
		c.JSON(http.StatusOK, FeedResponse{
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
			Sections:   []Section{},
		})
		return
	}

	// Convert media to MediaItem
	items := make([]MediaItem, len(medias))
	for i, m := range medias {
		username := ""
		if len(m.Edges.User) > 0 {
			username = m.Edges.User[0].Username
		}
		items[i] = MediaItem{
			ID:          int(m.ID),
			Title:       m.Title,
			Description: m.Description,
			Thumbnail: m.Thumbnail,
			Duration:  m.Duration,
			ViewCount: m.ViewCount,
			UserID:    m.UserID,
			Username:  username,
			URL:       fmt.Sprintf("/v/%d", m.ID),
			Type:      m.Type,
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