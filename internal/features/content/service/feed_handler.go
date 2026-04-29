/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"origadmin/application/origcms/internal/handler"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/server"
)

// FeedHandler handles feed-related HTTP endpoints
type FeedHandler struct {
	uc *biz.FeedUseCase
}

func NewFeedHandler(uc *biz.FeedUseCase) *FeedHandler {
	return &FeedHandler{uc: uc}
}

func (h *FeedHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	r.GET("/feed", h.GetFeed)
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
func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	ctx := r.Context()
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	medias, total, err := h.uc.ListLatest(ctx, page, pageSize)
	if err != nil {
		server.Fail(c.GinContext(), server.ErrInternal, "failed to fetch feed")
		return
	}

	server.OK(c.GinContext(), FeedResponse{
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
