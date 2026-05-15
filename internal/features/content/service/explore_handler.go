package service

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origstudio/internal/helpers/http"
	ginadapter "origadmin/application/origstudio/internal/helpers/http/gin"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/media"
)

type ExploreHandler struct {
	entityClient *entity.Client
}

func NewExploreHandler(entityClient *entity.Client) *ExploreHandler {
	return &ExploreHandler{entityClient: entityClient}
}

func (h *ExploreHandler) RegisterRoutes(r http2.Router) {
	explore := r.Group("/explore")
	{
		explore.GET("/trending", h.trending())
	}
}

func (h *ExploreHandler) trending() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc.DefaultQuery("period", "week")
		limit, _ := strconv.Atoi(gc.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 100 {
			limit = 50
		}

		reqCtx := ctx.Request().Context()

		medias, err := h.entityClient.Media.Query().
			Limit(limit).
			Order(entity.Desc(media.FieldViewCount)).
			All(reqCtx)
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}

		items := make([]interface{}, 0, len(medias))
		for _, m := range medias {
			publishedAt := ""
			if !m.PublishedAt.IsZero() {
				publishedAt = m.PublishedAt.Format(time.RFC3339)
			}

			items = append(items, gin.H{
				"id":           m.ID,
				"short_token":  m.ShortToken,
				"title":        m.Title,
				"description":  m.Description,
				"thumbnail":    m.Thumbnail,
				"duration":     m.Duration,
				"view_count":   m.ViewCount,
				"like_count":   m.LikeCount,
				"published_at": publishedAt,
			})
		}

		http2.OK(ctx, gin.H{
			"items": items,
			"total": len(items),
		})
		return nil
	}
}
