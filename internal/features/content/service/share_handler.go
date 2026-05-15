package service

import (
	"net/url"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origstudio/internal/helpers/http"
	ginadapter "origadmin/application/origstudio/internal/helpers/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/features/content/biz"
)

type ShareHandler struct {
	uc  *biz.LikeFavoriteUseCase
	jwt *auth.Manager
}

func NewShareHandler(uc *biz.LikeFavoriteUseCase, jwt *auth.Manager) *ShareHandler {
	return &ShareHandler{uc: uc, jwt: jwt}
}

func (h *ShareHandler) RegisterRoutes(r http2.Router) {
	r.GET("/medias/:id/shares", h.getShareUrl())
	r.POST("/medias/:id/shares", server.WithJWTCtx(h.jwt, h.recordShare()))
}

type SocialShareLinks struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	Twitter  string `json:"twitter"`
	Facebook string `json:"facebook"`
	LinkedIn string `json:"linkedin"`
	WhatsApp string `json:"whatsapp"`
	Telegram string `json:"telegram"`
}

func (h *ShareHandler) getShareUrl() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		mediaID := gc.Param("id")

		shareUrl := gc.Request.Host + "/watch/" + mediaID
		if len(shareUrl) > 0 && shareUrl[0] != 'h' {
			shareUrl = "https://" + shareUrl
		}

		title := gc.DefaultQuery("title", "Check out this video!")
		encodedUrl := url.QueryEscape(shareUrl)
		encodedTitle := url.QueryEscape(title)

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

func (h *ShareHandler) recordShare() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}
