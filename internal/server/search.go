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
		search.GET("/suggestions", h.suggestions)
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

// suggestions returns search suggestions based on query
// GET /search/suggestions?q=keyword&limit=10
func (h *SearchHandler) suggestions(c *gin.Context) {
	keyword := c.Query("q")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 20 {
		limit = 10
	}

	if keyword == "" {
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	// Get media titles as suggestions
	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     1,
			PageSize: int32(limit),
			Keyword:  keyword,
		},
	}

	medias, _, err := h.mediaUC.ListMedias(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	suggestions := make([]string, 0, len(medias))
	for _, m := range medias {
		if m.Title != "" {
			suggestions = append(suggestions, m.Title)
		}
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}
