package service

import (
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemservice "origadmin/application/origcms/internal/features/system/service"
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
	// Apply ModuleGuard gin middleware via type assertion
	if adapter, ok := playlists.(*ginadapter.RouterAdapter); ok {
		adapter.Use(systemservice.ModuleGuard(h.settingUC, "module_videos"))
	}
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
		gc := ginadapter.GinContextFromHTTP(ctx)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		items, total, err := h.playlistUC.ListPlaylists(ctx.Request().Context(), page, pageSize)
		if err != nil {
			server.FailCtx(ctx, server.ErrInternal, err.Error())
			return nil
		}

		server.PageCtx(ctx, items, int64(total), page, pageSize)
		return nil
	}
}

// getPlaylistByToken returns a single playlist by short_token.
// Public playlists are accessible to everyone.
// Private playlists are only accessible to their owner (requires JWT).
func (h *PlaylistHandler) getPlaylistByToken() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		token := gc.Param("token")
		if token == "" {
			server.FailCtx(ctx, server.ErrBadRequest, "playlist token is required")
			return nil
		}

		playlist, err := h.playlistUC.GetPlaylistByShortToken(ctx.Request().Context(), token)
		if err != nil {
			server.FailCtx(ctx, server.ErrNotFound, "playlist not found")
			return nil
		}

		// Private playlists: only the owner can view them
		if !playlist.IsPublic {
			claims, ok := server.GetClaimsCtx(ctx)
			if !ok || claims.GetUserID() != playlist.UserID {
				server.FailCtx(ctx, server.ErrNotFound, "playlist not found")
				return nil
			}
		}

		server.OKCtx(ctx, gin.H{"playlist": playlist})
		return nil
	}
}
