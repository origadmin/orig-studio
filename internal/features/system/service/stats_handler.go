package service

import (
	"origadmin/application/origcms/internal/handler"
	"github.com/gin-gonic/gin"

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
	r := handler.NewGinRouterAdapter(rg)
	stats := r.Group("/stats")
	{
		// Note: We can't use Use() directly with the Router interface
		// We'll need to apply middleware to each route individually
		stats.GET("/dashboard", server.WithJWT(h.jwt, server.GinHandlerToHTTP(h.getDashboardStats)))
		stats.GET("/media", server.WithJWT(h.jwt, server.GinHandlerToHTTP(h.getMediaStats)))
		stats.GET("/user", server.WithJWT(h.jwt, server.GinHandlerToHTTP(h.getUserStats)))
	}
}

// getDashboardStats returns dashboard statistics.
// GET /stats/dashboard
func (h *StatsHandler) getDashboardStats(c *gin.Context) {
	stats, err := h.statsRepo.GetDashboardStats(c)
	if err != nil {
		server.Fail(c, 500, "Failed to get dashboard stats: "+err.Error())
		return
	}
	server.OK(c, stats)
}

// getMediaStats returns media statistics.
// GET /stats/media
func (h *StatsHandler) getMediaStats(c *gin.Context) {
	stats, err := h.statsRepo.GetMediaStats(c)
	if err != nil {
		server.Fail(c, 500, "Failed to get media stats: "+err.Error())
		return
	}
	server.OK(c, stats)
}

// getUserStats returns user statistics.
// GET /stats/user
func (h *StatsHandler) getUserStats(c *gin.Context) {
	stats, err := h.statsRepo.GetUserStats(c)
	if err != nil {
		server.Fail(c, 500, "Failed to get user stats: "+err.Error())
		return
	}
	server.OK(c, stats)
}
