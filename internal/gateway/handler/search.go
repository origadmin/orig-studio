/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package handler

import (
	"context"

	"github.com/origadmin/runtime/log"
	mediav1 "origadmin/application/origcms/api/v1/services/media"
	"origadmin/application/origcms/internal/gateway/client"
)

// SearchHandler handles cross-service search aggregation.
// Migrated from: internal/svc-portal/biz/portal.go Search
type SearchHandler struct {
	clients *client.Clients
	log     *log.Helper
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(clients *client.Clients, logger log.Logger) *SearchHandler {
	return &SearchHandler{
		clients: clients,
		log:     log.NewHelper(log.With(logger, "module", "gateway.handler.search")),
	}
}

// SearchRequest defines search parameters.
type SearchRequest struct {
	Query    string `json:"query"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
}

// SearchResponse contains search results.
type SearchResponse struct {
	Videos []*MediaItem `json:"videos"`
	Total  int32        `json:"total"`
}

// Search performs a media search and returns matching results.
func (h *SearchHandler) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	mediaResp, err := h.clients.Media.ListMedias(ctx, &mediav1.ListMediasRequest{
		Keyword:  req.Query,
		Page:     req.Page,
		PageSize: pageSize,
	})
	if err != nil {
		h.log.Errorf("search failed for query %q: %v", req.Query, err)
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

	return &SearchResponse{
		Videos: items,
		Total:  mediaResp.Total,
	}, nil
}
