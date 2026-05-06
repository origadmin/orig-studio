package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	pb "origadmin/application/origcms/api/gen/v1/media"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dto"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
)

// SearchHandler handles search-related routes.
type SearchHandler struct {
	mediaUC *biz.MediaUseCase
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(mediaUC *biz.MediaUseCase) *SearchHandler {
	return &SearchHandler{mediaUC: mediaUC}
}

func (h *SearchHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := ginadapter.NewStdRouterAdapter(rg)
	search := r.Group("/search")
	{
		search.GET("", h.search())
		search.GET("/suggestions", h.suggestions())
	}
}

// search performs a search on media.
// GET /search?q=keyword&page=1&page_size=20
func (h *SearchHandler) search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		keyword := gc.Query("q")
		pageStr := gc.DefaultQuery("page", "1")
		pageSizeStr := gc.DefaultQuery("page_size", "20")

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
		if tagsStr := gc.Query("tags"); tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			opts.Tags = tags
		}

		medias, total, err := h.mediaUC.ListMedias(r.Context(), opts)
		if err != nil {
			server.Fail(gc, 50000, err.Error())
			return
		}

		totalPages := int32(0)
		if pageSize > 0 {
			totalPages = (total + int32(pageSize) - 1) / int32(pageSize)
		}
		server.OK(gc, &pb.ListMediasResponse{
			Total:      total,
			Items:      medias,
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalPages: totalPages,
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// suggestions returns search suggestions based on query
// GET /search/suggestions?q=keyword&limit=10
func (h *SearchHandler) suggestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		keyword := gc.Query("q")
		limitStr := gc.DefaultQuery("limit", "10")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			limit = 10
		}

		if keyword == "" {
			server.OK(gc, &pb.GetSearchSuggestionsResponse{
				Suggestions: []string{},
			})
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

		medias, _, err := h.mediaUC.ListMedias(r.Context(), opts)
		if err != nil {
			server.OK(gc, &pb.GetSearchSuggestionsResponse{
				Suggestions: []string{},
			})
			return
		}

		suggestions := make([]string, 0, len(medias))
		for _, m := range medias {
			if m.Title != "" {
				suggestions = append(suggestions, m.Title)
			}
		}

		server.OK(gc, &pb.GetSearchSuggestionsResponse{
			Suggestions: suggestions,
		})
	}
}
