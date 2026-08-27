package service

import (
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	contentdal "origadmin/application/origstudio/internal/features/content/dal"
)

type ExploreHandler struct {
	exploreQS *contentdal.ExploreQueryService
}

func NewExploreHandler(exploreQS *contentdal.ExploreQueryService) *ExploreHandler {
	return &ExploreHandler{exploreQS: exploreQS}
}

func (h *ExploreHandler) RegisterRoutes(r http2.Router) {
	explore := r.Group("/explore")
	{
		explore.GET("/trending", h.trending())
	}
}

func (h *ExploreHandler) trending() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		_ = ctx.QueryVarDefault("period", "week")
		limit, _ := strconv.Atoi(ctx.QueryVarDefault("limit", "50"))

		reqCtx := ctx.Request().Context()

		items, err := h.exploreQS.GetTrending(reqCtx, limit)
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}

		result := make([]interface{}, 0, len(items))
		for _, item := range items {
			result = append(result, map[string]any{
				"id":           item.ID,
				"short_token":  item.ShortToken,
				"title":        item.Title,
				"description":  item.Description,
				"thumbnail":    item.Thumbnail,
				"duration":     item.Duration,
				"view_count":   item.ViewCount,
				"like_count":   item.LikeCount,
				"published_at": item.PublishedAt,
			})
		}

		http2.OK(ctx, map[string]any{
			"items": result,
			"total": len(result),
		})
		return nil
	}
}
