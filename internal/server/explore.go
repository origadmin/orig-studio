package server

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/data/entity/media"
)

type ExploreHandler struct {
	entityClient *entity.Client
}

func NewExploreHandler(entityClient *entity.Client) *ExploreHandler {
	return &ExploreHandler{entityClient: entityClient}
}

func (h *ExploreHandler) Register(r handler.Router) {
	explore := r.Group("/explore")
	{
		explore.GET("/trending", GinHandlerToHTTP(h.trending))
	}
}

func (h *ExploreHandler) trending(c *gin.Context) {
	_ = c.DefaultQuery("period", "week")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	ctx := c.Request.Context()

	medias, err := h.entityClient.Media.Query().
		Limit(limit).
		Order(entity.Desc(media.FieldViewCount)).
		All(ctx)
	if err != nil {
		Fail(c, 50000, err.Error())
		return
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

	response := PageResponse[interface{}]{}
	response.Data.Items = items
	response.Data.Total = int64(len(items))

	c.JSON(200, response)
}
