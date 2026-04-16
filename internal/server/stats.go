package server

import (
	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/svc-content/biz"
	"origadmin/application/origcms/internal/svc-system/data"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
)

// StatsHandler handles stats-related routes.
type StatsHandler struct {
	mediaUC        *mediabiz.MediaUseCase
	likeFavoriteUC *biz.LikeFavoriteUseCase
	statsRepo      *data.StatsRepo
	jwt            *auth.Manager
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(mediaUC *mediabiz.MediaUseCase, likeFavoriteUC *biz.LikeFavoriteUseCase, statsRepo *data.StatsRepo, jwt *auth.Manager) *StatsHandler {
	return &StatsHandler{
		mediaUC:        mediaUC,
		likeFavoriteUC: likeFavoriteUC,
		statsRepo:      statsRepo,
		jwt:            jwt,
	}
}

func (h *StatsHandler) Register(r handler.Router) {
	stats := r.Group("/stats")
	{
		// Note: We can't use Use() directly with the Router interface
		// We'll need to apply middleware to each route individually
		stats.GET("/dashboard", WithJWT(h.jwt, GinHandlerToHTTP(h.getDashboardStats)))
		stats.GET("/media", WithJWT(h.jwt, GinHandlerToHTTP(h.getMediaStats)))
		stats.GET("/user", WithJWT(h.jwt, GinHandlerToHTTP(h.getUserStats)))
	}
}

// getDashboardStats returns dashboard statistics.
// GET /stats/dashboard
func (h *StatsHandler) getDashboardStats(c *gin.Context) {
	stats, err := h.statsRepo.GetDashboardStats(c)
	if err != nil {
		Fail(c, 500, "Failed to get dashboard stats: "+err.Error())
		return
	}
	OK(c, stats)
}

// getMediaStats returns media statistics.
// GET /stats/media
func (h *StatsHandler) getMediaStats(c *gin.Context) {
	stats, err := h.statsRepo.GetMediaStats(c)
	if err != nil {
		Fail(c, 500, "Failed to get media stats: "+err.Error())
		return
	}
	OK(c, stats)
}

// getUserStats returns user statistics.
// GET /stats/user
func (h *StatsHandler) getUserStats(c *gin.Context) {
	stats, err := h.statsRepo.GetUserStats(c)
	if err != nil {
		Fail(c, 500, "Failed to get user stats: "+err.Error())
		return
	}
	OK(c, stats)
}
