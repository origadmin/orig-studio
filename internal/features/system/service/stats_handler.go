package service

import (
	"net/http"

	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	systemdal "origadmin/application/origstudio/internal/features/system/dal"
	mediabiz "origadmin/application/origstudio/internal/features/media/biz"

	http2 "origadmin/application/origstudio/internal/pkg/http"
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

func (h *StatsHandler) RegisterRoutes(r http2.Router) {
	stats := r.Group("/stats")
	{
		stats.GET("/dashboard", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.getDashboardStats())))
		stats.GET("/media", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.getMediaStats())))
		stats.GET("/user", server.WithJWTCtx(h.jwt, server.HTTPToHandlerFunc(h.getUserStats())))
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
