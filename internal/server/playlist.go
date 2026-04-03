package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
)

type PlaylistHandler struct {
	client *entity.Client
}

func NewPlaylistHandler(client *entity.Client) *PlaylistHandler {
	return &PlaylistHandler{client: client}
}

func (h *PlaylistHandler) Register(group *gin.RouterGroup) {
	playlists := group.Group("/playlists")
	{
		playlists.GET("", h.listPlaylists())
		playlists.POST("", h.createPlaylist())
		playlists.DELETE("/:id", h.deletePlaylist())
	}
}

func (h *PlaylistHandler) listPlaylists() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := h.client.Playlist.Query().All(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"list": items})
	}
}

func (h *PlaylistHandler) createPlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			UserID      int    `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p, err := h.client.Playlist.Create().
			SetTitle(input.Title).
			SetDescription(input.Description).
			SetUserID(input.UserID).
			Save(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, p)
	}
}

func (h *PlaylistHandler) deletePlaylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = h.client.Playlist.DeleteOneID(id).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
