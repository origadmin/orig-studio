package service

import (
	"origadmin/application/origcms/internal/handler"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/helpers/repo"
	mediabiz "origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dto"
)

// SearchHandler handles search-related routes.
type SearchHandler struct {
	mediaUC *mediabiz.MediaUseCase
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(mediaUC *mediabiz.MediaUseCase) *SearchHandler {
	return &SearchHandler{mediaUC: mediaUC}
}

func (h *SearchHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	search := r.Group("/search")
	{
		search.GET("", server.GinHandlerToHTTP(h.search))
		search.GET("/suggestions", server.GinHandlerToHTTP(h.suggestions))
	}
}

// search performs a search on media.
// GET /search?q=keyword&page=1&page_size=20
func (h *SearchHandler) search(c *gin.Context) {
	keyword := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 20
	}

	// Normalize pagination parameters
	page, pageSize = repo.NormalizePagination(page, pageSize)

	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  keyword,
		},
	}

	// Handle tags filtering
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		opts.Tags = tags
	}

	medias, total, err := h.mediaUC.ListMedias(c.Request.Context(), opts)
	if err != nil {
		server.Fail(c, 50000, err.Error())
		return
	}

	// Build pagination response
	response := server.PageResponse[interface{}]{
		Code:    0,
		Message: "ok",
	}
	response.Data.Items = make([]interface{}, len(medias))
	for i, media := range medias {
		response.Data.Items[i] = media
	}
	response.Data.Total = int64(total)
	response.Data.Page = page
	response.Data.PageSize = pageSize

	c.JSON(200, response)
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
		server.OK(c, gin.H{"suggestions": []string{}})
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
		server.OK(c, gin.H{"suggestions": []string{}})
		return
	}

	suggestions := make([]string, 0, len(medias))
	for _, m := range medias {
		if m.Title != "" {
			suggestions = append(suggestions, m.Title)
		}
	}

	server.OK(c, gin.H{"suggestions": suggestions})
}
