package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
)

// StatsHandler handles stats-related routes.
type StatsHandler struct {
	mediaUC        *mediabiz.MediaUseCase
	likeFavoriteUC *biz.LikeFavoriteUseCase
	jwt            *auth.Manager
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(mediaUC *mediabiz.MediaUseCase, likeFavoriteUC *biz.LikeFavoriteUseCase, jwt *auth.Manager) *StatsHandler {
	return &StatsHandler{
		mediaUC:        mediaUC,
		likeFavoriteUC: likeFavoriteUC,
		jwt:            jwt,
	}
}

func (h *StatsHandler) Register(group *gin.RouterGroup) {
	stats := group.Group("/stats")
	{
		stats.Use(JWTMiddleware(h.jwt))
		stats.GET("/dashboard", h.getDashboardStats)
	}
}

// getDashboardStats returns dashboard statistics.
// GET /stats/dashboard
func (h *StatsHandler) getDashboardStats(c *gin.Context) {
	// TODO: Implement proper stats using the use cases
	// For now, return mock data

	// Example response matching the frontend expectations
	c.JSON(http.StatusOK, gin.H{
		"total_media":    10,
		"total_users":    5,
		"total_views":    1000,
		"total_comments": 50,
		"top_media": []gin.H{
			{"id": 1, "title": "Sample Video 1", "view_count": 500},
			{"id": 2, "title": "Sample Video 2", "view_count": 300},
			{"id": 3, "title": "Sample Video 3", "view_count": 200},
		},
		"recent_activity": []gin.H{
			{"type": "upload", "message": "New video uploaded", "time": "2024-01-01T00:00:00Z"},
			{"type": "comment", "message": "New comment added", "time": "2024-01-02T00:00:00Z"},
		},
	})
}
