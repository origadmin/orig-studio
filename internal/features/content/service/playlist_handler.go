package service

import (
	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/server"
)

// PlaylistHandler handles playlist-related HTTP endpoints.
type PlaylistHandler struct {
	client *entity.Client
}

// NewPlaylistHandler creates a new PlaylistHandler.
func NewPlaylistHandler(client *entity.Client) *PlaylistHandler {
	return &PlaylistHandler{client: client}
}

// RegisterRoutes registers the handler's routes.
func (h *PlaylistHandler) RegisterRoutes(rg *gin.RouterGroup) {
	playlists := rg.Group("/playlists")
	{
		playlists.GET("", h.listPlaylists)
		playlists.POST("", h.createPlaylist)
	}
}

func (h *PlaylistHandler) listPlaylists(c *gin.Context) {
	items, err := h.client.Playlist.Query().All(c.Request.Context())
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}
	server.OK(c, gin.H{
		"items":     items,
		"total":     len(items),
		"page":      1,
		"page_size": len(items),
	})
}

func (h *PlaylistHandler) createPlaylist(c *gin.Context) {
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		UserID      string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	p, err := h.client.Playlist.Create().
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetUserID(input.UserID).
		Save(c.Request.Context())
	if err != nil {
		server.Fail(c, server.ErrInternal, err.Error())
		return
	}

	server.Created(c, p)
}
