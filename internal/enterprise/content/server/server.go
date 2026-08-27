package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/wire"

	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	portalservice "origadmin/application/origstudio/internal/enterprise/content/portal/service"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	std "origadmin/application/origstudio/internal/pkg/http/std"
)

var ProviderSet = wire.NewSet(NewEnterpriseContentServer)

// EnterpriseContentServer wraps the enterprise content handler (portal)
// and provides an HTTP handler that registers its routes on a Gin engine.
type EnterpriseContentServer struct {
	portalHandler *portalservice.Handler
	settingUC     *systembiz.SettingUseCase
}

func NewEnterpriseContentServer(
	portalHandler *portalservice.Handler,
	settingUC *systembiz.SettingUseCase,
) *EnterpriseContentServer {
	return &EnterpriseContentServer{
		portalHandler: portalHandler,
		settingUC:     settingUC,
	}
}

// resolveShareBaseURL returns the absolute site base for share links.
//
// The base MUST come from the admin-configured "主Web地址" (primary_url), NOT
// from the inbound request host. The content service sits behind the gateway
// reverse proxy, so req.Host is the upstream target (e.g. content:8003) and
// would yield a broken link. Fallback order:
//  1. primary_url (the canonical main web address)        -> preferred
//  2. first entry of base_urls (allowed site URLs)        -> fallback
//  3. X-Forwarded-Host / -Proto injected by the gateway   -> last resort
//     (only for fresh deployments that have not set any URL yet)
func resolveShareBaseURL(settingUC *systembiz.SettingUseCase, ctx http2.Context) string {
	c := ctx.Request().Context()
	if v := strings.TrimSpace(settingUC.Get(c, "primary_url")); v != "" {
		return v
	}
	if raw := settingUC.Get(c, "base_urls"); raw != "" {
		var urls []string
		if err := json.Unmarshal([]byte(raw), &urls); err == nil {
			for _, u := range urls {
				if u = strings.TrimSpace(u); u != "" {
					return u
				}
			}
		}
	}
	// Last-resort fallback for unconfigured deployments.
	host := ctx.Request().Host
	if fwd := ctx.GetHeader("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	proto := "http"
	if p := ctx.GetHeader("X-Forwarded-Proto"); p != "" {
		proto = p
	}
	return proto + "://" + host
}

// RegisterRoutes registers all enterprise content routes on the given router.
func (s *EnterpriseContentServer) RegisterRoutes(r http2.Router) {
	s.portalHandler.RegisterRoutes(r)

	// BUG-267: share-link endpoint. The web client's ShareDialog calls
	// GET /api/v1/medias/{token}/shares and expects { url } — an absolute watch
	// link — to populate the copy/share box. The media gRPC GetMediaShares returns
	// { share_count, share_url } (relative, wrong shape), so we own this route on
	// the content service and return the correct {url}. The share URL base is
	// sourced from the admin-configured primary_url (主Web地址); see
	// resolveShareBaseURL. The watch URL uses the frontend's actual route:
	// /watch?v={short_token}.
	shareHandler := func(ctx http2.Context) error {
		token := ctx.Var("token")
		base := strings.TrimRight(resolveShareBaseURL(s.settingUC, ctx), "/")
		shareURL := base + "/watch?v=" + token
		http2.OK(ctx, map[string]any{
			"url":   shareURL,
			"title": ctx.QueryVarDefault("title", "Check out this video!"),
		})
		return nil
	}
	r.GET("/medias/:token/shares", shareHandler)
	// Optional "record share" — just ack so the client's share() call succeeds.
	r.POST("/medias/:token/shares", func(ctx http2.Context) error {
		http2.OK(ctx, map[string]any{"success": true})
		return nil
	})
}

// HTTPHandler returns an http.Handler that serves all enterprise content routes.
func (s *EnterpriseContentServer) HTTPHandler() http.Handler {
	router := std.NewRouter()
	apiV1 := router.Group("/api/v1")
	s.RegisterRoutes(apiV1)
	return router
}
