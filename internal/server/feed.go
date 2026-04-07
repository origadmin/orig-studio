/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/svc-content/biz"
)

// FeedHandler handles feed-related HTTP endpoints
type FeedHandler struct {
	uc *biz.FeedUseCase
}

func NewFeedHandler(uc *biz.FeedUseCase) *FeedHandler {
	return &FeedHandler{uc: uc}
}

func (h *FeedHandler) Register(group *gin.RouterGroup) {
	group.GET("/feed", h.GetFeed)
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
	Title string           `json:"title"`
	Type  string           `json:"type"`
	Items []*biz.MediaInfo `json:"items"`
}

// GetFeed godoc: GET /api/v1/feed
func (h *FeedHandler) GetFeed(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	medias, total, err := h.uc.ListLatest(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
		return
	}

	c.JSON(http.StatusOK, FeedResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		Sections: []Section{
			{
				Title: "Recommended for You",
				Type:  "recommended",
				Items: medias,
			},
		},
	})
}
