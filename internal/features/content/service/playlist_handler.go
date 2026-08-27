package service

import (
	"strconv"

	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"
)

// PlaylistHandler handles public/portal playlist HTTP endpoints.
// User-scoped operations (CRUD, add/remove media) are handled by MeHandler under /me/playlists.
// Admin operations are handled by AdminHandler under /admin/playlists.
//
// Portal routes use short_token (not database id) as the public identifier,
// consistent with ChannelHandler's /:token pattern and the project's
// "no id exposure" design principle (see A005 analysis).
type PlaylistHandler struct {
	playlistUC *biz.PlaylistChannelUseCase
	settingUC  *systembiz.SettingUseCase
	jwt        *auth.Manager
}

// NewPlaylistHandler creates a new PlaylistHandler.
func NewPlaylistHandler(playlistUC *biz.PlaylistChannelUseCase, settingUC *systembiz.SettingUseCase, jwt *auth.Manager) *PlaylistHandler {
	return &PlaylistHandler{playlistUC: playlistUC, settingUC: settingUC, jwt: jwt}
}

// RegisterRoutes registers the handler's routes.
func (h *PlaylistHandler) RegisterRoutes(r http2.Router) {
	playlists := r.Group("/playlists")
	playlists.Use(systemservice.ModuleGuardCtx(h.settingUC, "module_videos"))
	{
		playlists.GET("", h.listPlaylists())
		// Use OptionalJWTMiddleware so that private playlists can be accessed
		// by their owner. Without this, GetClaims(c) always returns ok=false
		// for the portal route, causing 404 for any private playlist (B099).
		playlists.GET("/:token", server.WithOptionalJWTCtx(h.jwt, h.getPlaylistByToken()))
	}
}

// listPlaylists returns all public playlists with pagination (portal view).
func (h *PlaylistHandler) listPlaylists() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.playlistUC.ListPlaylists(ctx.Request().Context(), page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// getPlaylistByToken returns a single playlist by short_token.
// Public playlists are accessible to everyone.
// Private playlists are only accessible to their owner (requires JWT).
func (h *PlaylistHandler) getPlaylistByToken() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		token := ctx.Var("token")
		if token == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "playlist token is required")
		}

		playlist, err := h.playlistUC.GetPlaylistByShortToken(ctx.Request().Context(), token)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "playlist not found")
		}

		// Private playlists: only the owner can view them
		if !playlist.IsPublic {
			claims, ok := server.GetClaimsCtx(ctx)
			if !ok || claims.GetUserID() != playlist.UserID {
				return server.FailCtx(ctx, server.ErrNotFound, "playlist not found")
			}
		}

		return server.OKCtx(ctx, map[string]any{"playlist": playlist})
	}
}
