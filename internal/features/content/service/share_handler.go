package service

import (
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/server"
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

func (h *ShareHandler) RegisterRoutes(r http2.Router) {
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
func (h *ShareHandler) getShareUrl() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		mediaId, err := strconv.Atoi(gc.Param("mediaId"))
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "Invalid media ID")
			return nil
		}

		// Build share URL - assuming the frontend is at /watch/:mediaId
		// You may want to make this configurable
		shareUrl := gc.Request.Host + "/watch/" + strconv.Itoa(mediaId)
		// Add https:// if not present
		if len(shareUrl) > 0 && shareUrl[0] != 'h' {
			shareUrl = "https://" + shareUrl
		}

		// Get title from query or use default
		title := gc.DefaultQuery("title", "Check out this video!")
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

		http2.OK(ctx, socialLinks)
		return nil
	}
}

// recordShare records a share event.
// POST /media/:mediaId/share → {"success": bool}
func (h *ShareHandler) recordShare() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_, exists := gc.Get("claims")
		if !exists {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}

		_, err := strconv.Atoi(gc.Param("mediaId"))
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "Invalid media ID")
			return nil
		}

		// TODO: Implement share count increment in the future
		// For now, just return success

		http2.OK(ctx, gin.H{
			"success": true,
		})
		return nil
	}
}
