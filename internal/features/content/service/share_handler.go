package service

import (
	"encoding/json"
	"net/url"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/features/content/biz"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type ShareHandler struct {
	uc       *biz.LikeFavoriteUseCase
	jwt      *auth.Manager
	settingUC *systembiz.SettingUseCase
}

func NewShareHandler(uc *biz.LikeFavoriteUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase) *ShareHandler {
	return &ShareHandler{uc: uc, jwt: jwt, settingUC: settingUC}
}

func (h *ShareHandler) RegisterRoutes(r http2.Router) {
	r.GET("/medias/:token/shares", h.getShareUrl())
	r.POST("/medias/:token/shares", server.WithJWTCtx(h.jwt, h.recordShare()))
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
		mediaID := ctx.Var("token")

		shareUrl := ctx.Request().Host + "/watch/" + mediaID
		if len(shareUrl) > 0 && shareUrl[0] != 'h' {
			shareUrl = "https://" + shareUrl
		}

		title := ctx.QueryVarDefault("title", "Check out this video!")
		encodedUrl := url.QueryEscape(shareUrl)
		encodedTitle := url.QueryEscape(title)

		enabled := h.enabledSharePlatforms(ctx)
		socialLinks := SocialShareLinks{
			Url:   shareUrl,
			Title: title,
		}
		if enabled["twitter"] {
			socialLinks.Twitter = "https://twitter.com/intent/tweet?url=" + encodedUrl + "&text=" + encodedTitle
		}
		if enabled["facebook"] {
			socialLinks.Facebook = "https://www.facebook.com/sharer/sharer.php?u=" + encodedUrl
		}
		if enabled["linkedin"] {
			socialLinks.LinkedIn = "https://www.linkedin.com/sharing/share-offsite/?url=" + encodedUrl
		}
		if enabled["whatsapp"] {
			socialLinks.WhatsApp = "https://wa.me/?text=" + encodedTitle + "%20" + encodedUrl
		}
		if enabled["telegram"] {
			socialLinks.Telegram = "https://t.me/share/url?url=" + encodedUrl + "&text=" + encodedTitle
		}

		http2.OK(ctx, socialLinks)
		return nil
	}
}

// sharePlatformKeys mirrors the canonical set in system/service (twitter,
// facebook, whatsapp, telegram, linkedin, weibo). Kept local to avoid a cross
// import just for the slice; stay in sync if platforms change.
var sharePlatformKeys = []string{"twitter", "facebook", "whatsapp", "telegram", "linkedin", "weibo"}

func (h *ShareHandler) enabledSharePlatforms(ctx http2.Context) map[string]bool {
	out := make(map[string]bool, len(sharePlatformKeys))
	for _, k := range sharePlatformKeys {
		out[k] = true
	}
	if h.settingUC == nil {
		return out
	}
	raw := h.settingUC.Get(ctx, "share_platforms")
	if raw == "" {
		return out
	}
	var cfg map[string]bool
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return out
	}
	for k, v := range cfg {
		out[k] = v
	}
	return out
}

func (h *ShareHandler) recordShare() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, map[string]any{"success": true})
		return nil
	}
}
