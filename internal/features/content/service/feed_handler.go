package service

import (
	"strconv"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
)

// FeedHandler handles feed-related HTTP endpoints
type FeedHandler struct {
	uc *biz.FeedUseCase
}

func NewFeedHandler(uc *biz.FeedUseCase) *FeedHandler {
	return &FeedHandler{uc: uc}
}

func (h *FeedHandler) RegisterRoutes(r http2.Router) {
	r.GET("/feed", h.getFeed())
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

// getFeed godoc: GET /api/v1/feed
func (h *FeedHandler) getFeed() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		page, _ := strconv.Atoi(gc.Query("page"))
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(gc.Query("page_size"))
		if pageSize == 0 {
			pageSize = 20
		}
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		medias, total, err := h.uc.ListLatest(ctx.Request().Context(), page, pageSize)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, "failed to fetch feed")
			return nil
		}

		http2.OK(ctx, FeedResponse{
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
		return nil
	}
}
