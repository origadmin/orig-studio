package service

import (
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/features/content/biz"
)

type ShareHandler struct {
	uc  *biz.LikeFavoriteUseCase
	jwt *auth.Manager
}

func NewShareHandler(uc *biz.LikeFavoriteUseCase, jwt *auth.Manager) *ShareHandler {
	return &ShareHandler{uc: uc, jwt: jwt}
}

func (h *ShareHandler) RegisterRoutes(r http2.Router) {
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
		mediaId, err := strconv.Atoi(gc.Param("mediaId"))
		if err != nil {
			http2.Fail(ctx, http2.ErrBadRequest, "Invalid media ID")
			return nil
		}

		shareUrl := gc.Request.Host + "/watch/" + strconv.Itoa(mediaId)
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
		gc := ginadapter.GinContextFromHTTP(ctx)
		_, exists := gc.Get("claims")
		if !exists {
			http2.Fail(ctx, http2.ErrUnauthorized, "unauthorized")
			return nil
		}

		_, err := strconv.Atoi(gc.Param("mediaId"))
		if err != nil {
			http2.Fail(ctx, http2.ErrBadRequest, "Invalid media ID")
			return nil
		}

		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}
