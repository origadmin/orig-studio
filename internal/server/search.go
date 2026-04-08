package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/helpers/repo"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// SearchHandler handles search-related routes.
type SearchHandler struct {
	mediaUC *mediabiz.MediaUseCase
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(mediaUC *mediabiz.MediaUseCase) *SearchHandler {
	return &SearchHandler{mediaUC: mediaUC}
}

func (h *SearchHandler) Register(group *gin.RouterGroup) {
	search := group.Group("/search")
	{
		search.GET("", h.search)
	}
}

// search performs a search on media.
// GET /search?q=keyword&page=1&page_size=20
func (h *SearchHandler) search(c *gin.Context) {
	keyword := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  keyword,
		},
	}

	medias, total, err := h.mediaUC.ListMedias(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      medias,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func boolPtr(b bool) *bool {
	return &b
}
