package service

import (
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/domain/types"
)

type FeedHandler struct {
	uc *biz.FeedUseCase
}

func NewFeedHandler(uc *biz.FeedUseCase) *FeedHandler {
	return &FeedHandler{uc: uc}
}

func (h *FeedHandler) RegisterRoutes(r http2.Router) {
	r.GET("/feed", h.getFeed())
}

type FeedResponse struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalCount int       `json:"total_count"`
	Sections   []Section `json:"sections"`
}

type Section struct {
	Title string           `json:"title"`
	Type  string           `json:"type"`
	Items []*biz.MediaInfo `json:"items"`
}

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
		page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

		medias, total, err := h.uc.ListLatest(ctx.Request().Context(), page, pageSize)
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, "failed to fetch feed")
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
