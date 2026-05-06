package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	systemdal "origadmin/application/origcms/internal/features/system/dal"
	mediabiz "origadmin/application/origcms/internal/features/media/biz"
)

// StatsHandler handles stats-related routes.
type StatsHandler struct {
	mediaUC        *mediabiz.MediaUseCase
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase
	statsRepo      *systemdal.StatsRepo
	jwt            *auth.Manager
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(mediaUC *mediabiz.MediaUseCase, likeFavoriteUC *contentbiz.LikeFavoriteUseCase, statsRepo *systemdal.StatsRepo, jwt *auth.Manager) *StatsHandler {
	return &StatsHandler{
		mediaUC:        mediaUC,
		likeFavoriteUC: likeFavoriteUC,
		statsRepo:      statsRepo,
		jwt:            jwt,
	}
}

func (h *StatsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := ginadapter.NewStdRouterAdapter(rg)
	stats := r.Group("/stats")
	{
		stats.GET("/dashboard", server.WithJWT(h.jwt, h.getDashboardStats()))
		stats.GET("/media", server.WithJWT(h.jwt, h.getMediaStats()))
		stats.GET("/user", server.WithJWT(h.jwt, h.getUserStats()))
	}
}

// getDashboardStats returns dashboard statistics.
// GET /stats/dashboard
func (h *StatsHandler) getDashboardStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		stats, err := h.statsRepo.GetDashboardStats(r.Context())
		if err != nil {
			server.Fail(gc, 500, "Failed to get dashboard stats: "+err.Error())
			return
		}
		server.OK(gc, stats)
	}
}

// getMediaStats returns media statistics.
// GET /stats/media
func (h *StatsHandler) getMediaStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		stats, err := h.statsRepo.GetMediaStats(r.Context())
		if err != nil {
			server.Fail(gc, 500, "Failed to get media stats: "+err.Error())
			return
		}
		server.OK(gc, stats)
	}
}

// getUserStats returns user statistics.
// GET /stats/user
func (h *StatsHandler) getUserStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		stats, err := h.statsRepo.GetUserStats(r.Context())
		if err != nil {
			server.Fail(gc, 500, "Failed to get user stats: "+err.Error())
			return
		}
		server.OK(gc, stats)
	}
}
