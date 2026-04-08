package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
)

// ShareHandler handles share-related routes.
type ShareHandler struct {
	uc  *biz.LikeFavoriteUseCase
	jwt *auth.Manager
}

// NewShareHandler creates a new ShareHandler.
func NewShareHandler(uc *biz.LikeFavoriteUseCase, jwt *auth.Manager) *ShareHandler {
	return &ShareHandler{uc: uc, jwt: jwt}
}

func (h *ShareHandler) Register(group *gin.RouterGroup) {
	// Share routes are now defined in media.go with consistent :id parameter
}

// getShareUrl returns the share URL for a media item.
// GET /media/:mediaId/share → {"url": string}
func (h *ShareHandler) getShareUrl(c *gin.Context) {
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	// Build share URL - assuming the frontend is at /watch/:mediaId
	// You may want to make this configurable
	shareUrl := c.Request.Host + "/watch/" + strconv.Itoa(mediaId)
	// Add https:// if not present
	if len(shareUrl) > 0 && shareUrl[0] != 'h' {
		shareUrl = "https://" + shareUrl
	}

	c.JSON(http.StatusOK, gin.H{
		"url": shareUrl,
	})
}

// recordShare records a share event.
// POST /media/:mediaId/share → {"success": bool}
func (h *ShareHandler) recordShare(c *gin.Context) {
	_, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	_, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	// TODO: Implement share count increment in the future
	// For now, just return success

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
