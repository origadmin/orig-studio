/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package handler provides aggregation handlers for the API gateway.
// These handlers replace the former svc-portal aggregation logic.
package handler

import (
	"context"

	"github.com/origadmin/runtime/log"
	mediav1 "origadmin/application/origcms/api/v1/services/media"
	"origadmin/application/origcms/internal/gateway/client"
)

// FeedHandler handles homepage feed aggregation.
// Migrated from: internal/svc-portal/biz/portal.go GetHomeFeed
type FeedHandler struct {
	clients *client.Clients
	log     *log.Helper
}

// NewFeedHandler creates a new FeedHandler.
func NewFeedHandler(clients *client.Clients, logger log.Logger) *FeedHandler {
	return &FeedHandler{
		clients: clients,
		log:     log.NewHelper(log.With(logger, "module", "gateway.handler.feed")),
	}
}

// HomeFeedRequest defines the request for home feed.
type HomeFeedRequest struct {
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
	OrderBy  string `json:"order_by"` // "views" | "latest"
}

// HomeFeedResponse defines the response for home feed.
type HomeFeedResponse struct {
	Sections []*FeedSection `json:"sections"`
}

// FeedSection is a grouped list of media items in the feed.
type FeedSection struct {
	Title string       `json:"title"`
	Type  string       `json:"type"`
	Items []*MediaItem `json:"items"`
}

// MediaItem is a simplified media representation for feed display.
type MediaItem struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnail_url"`
	ViewCount    int64  `json:"view_count"`
	AuthorID     int64  `json:"author_id"`
}

// GetHomeFeed returns aggregated trending and latest media for the homepage.
func (h *FeedHandler) GetHomeFeed(ctx context.Context, req *HomeFeedRequest) (*HomeFeedResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	orderBy := req.OrderBy
	if orderBy == "" {
		orderBy = "views"
	}

	mediaResp, err := h.clients.Media.ListMedias(ctx, &mediav1.ListMediasRequest{
		PageSize:   pageSize,
		OrderBy:    orderBy,
		Descending: true,
	})
	if err != nil {
		h.log.Errorf("failed to list medias for home feed: %v", err)
		return nil, err
	}

	items := make([]*MediaItem, 0, len(mediaResp.Medias))
	for _, m := range mediaResp.Medias {
		items = append(items, &MediaItem{
			ID:           m.Id,
			Title:        m.Title,
			Description:  m.Description,
			ThumbnailURL: m.Thumbnail,
			ViewCount:    m.ViewCount,
			AuthorID:     m.UserId,
		})
	}

	return &HomeFeedResponse{
		Sections: []*FeedSection{
			{
				Title: "Trending",
				Type:  "trending",
				Items: items,
			},
		},
	}, nil
}
