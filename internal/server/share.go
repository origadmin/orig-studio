package server

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
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

func (h *ShareHandler) Register(r handler.Router) {
	// Share routes are now defined in media.go with consistent :id parameter
}

// SocialShareLinks contains all social media share links
type SocialShareLinks struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	Twitter  string `json:"twitter"`
	Facebook string `json:"facebook"`
	LinkedIn string `json:"linkedin"`
	WhatsApp string `json:"whatsapp"`
	Telegram string `json:"telegram"`
}

// getShareUrl returns the share URL and social media links for a media item.
// GET /media/:mediaId/share → {"url": string, "twitter": string, ...}
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

	// Get title from query or use default
	title := c.DefaultQuery("title", "Check out this video!")
	encodedUrl := url.QueryEscape(shareUrl)
	encodedTitle := url.QueryEscape(title)

	// Generate social media share links
	socialLinks := SocialShareLinks{
		Url:      shareUrl,
		Title:    title,
		Twitter:  "https://twitter.com/intent/tweet?url=" + encodedUrl + "&text=" + encodedTitle,
		Facebook: "https://www.facebook.com/sharer/sharer.php?u=" + encodedUrl,
		LinkedIn: "https://www.linkedin.com/sharing/share-offsite/?url=" + encodedUrl,
		WhatsApp: "https://wa.me/?text=" + encodedTitle + "%20" + encodedUrl,
		Telegram: "https://t.me/share/url?url=" + encodedUrl + "&text=" + encodedTitle,
	}

	c.JSON(http.StatusOK, socialLinks)
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
