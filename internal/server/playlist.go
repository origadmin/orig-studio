package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	entitymedia "origadmin/application/origcms/internal/data/entity/media"
	mediaplaylist "origadmin/application/origcms/internal/data/entity/mediaplaylist"
	pl "origadmin/application/origcms/internal/data/entity/playlist"
)

// PlaylistHandler handles /api/v1/playlists routes.
type PlaylistHandler struct {
	client *entity.Client
	jwt    *auth.Manager
}

// NewPlaylistHandler creates a new PlaylistHandler.
func NewPlaylistHandler(client *entity.Client, jwt *auth.Manager) *PlaylistHandler {
	return &PlaylistHandler{client: client, jwt: jwt}
}

func (h *PlaylistHandler) Register(group *gin.RouterGroup) {
	playlists := group.Group("/playlists")
	{
		// Public read routes
		playlists.GET("", h.listPlaylists)
		playlists.GET("/:id", h.getPlaylist)

		// Protected write routes
		protected := playlists.Group("")
		protected.Use(JWTMiddleware(h.jwt))
		{
			protected.POST("", h.createPlaylist)
			protected.PUT("/:id", h.updatePlaylist)
			protected.DELETE("/:id", h.deletePlaylist)
			// Media management within playlist
			protected.POST("/:id/media", h.addMedia)
			protected.DELETE("/:id/media/:mediaId", h.removeMedia)

			// User's playlists
			protected.GET("/my", h.myPlaylists)
		}
	}
}

// listPlaylists returns all playlists with pagination.
func (h *PlaylistHandler) listPlaylists(c *gin.Context) {
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.client.Playlist.Query().
		Limit(limit).
		Offset(offset).
		Order(entity.Desc(pl.FieldAddDate)).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, _ := h.client.Playlist.Query().Count(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"list":  items,
		"total": total,
	})
}

// getPlaylist returns a single playlist with its media items.
func (h *PlaylistHandler) getPlaylist(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	p, err := h.client.Playlist.Query().
		Where(pl.IDEQ(id)).
		WithUser().
		WithMedia(func(mq *entity.MediaQuery) {
			mq.WithUser()
			mq.Order(entity.Asc("files_playlistmedia_ordering"))
		}).
		Only(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// myPlaylists returns playlists for the authenticated user.
func (h *PlaylistHandler) myPlaylists(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	items, err := h.client.Playlist.Query().
		Where(pl.UserIDEQ(int(claims.UserID))).
		WithUser().
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": items})
}

// createPlaylist creates a new playlist for the authenticated user.
func (h *PlaylistHandler) createPlaylist(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Title         string `json:"title" binding:"required,max=100"`
		Description   string `json:"description"`
		FriendlyToken string `json:"friendly_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.client.Playlist.Create().
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetFriendlyToken(input.FriendlyToken).
		SetUserID(int(claims.UserID)).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	p, _ = h.client.Playlist.Query().
		Where(pl.IDEQ(p.ID)).
		WithUser().
		Only(c.Request.Context())

	c.JSON(http.StatusCreated, p)
}

// updatePlaylist updates a playlist. Only owner can update.
func (h *PlaylistHandler) updatePlaylist(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	p, err := h.client.Playlist.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}
	if p.UserID != int(claims.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only update own playlist"})
		return
	}

	var input struct {
		Title         *string `json:"title"`
		Description   *string `json:"description"`
		FriendlyToken *string `json:"friendly_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := h.client.Playlist.UpdateOneID(id)
	if input.Title != nil {
		update.SetTitle(*input.Title)
	}
	if input.Description != nil {
		update.SetDescription(*input.Description)
	}
	if input.FriendlyToken != nil {
		update.SetFriendlyToken(*input.FriendlyToken)
	}

	p, err = update.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	p, _ = h.client.Playlist.Query().
		Where(pl.IDEQ(id)).
		WithUser().
		Only(c.Request.Context())

	c.JSON(http.StatusOK, p)
}

// deletePlaylist deletes a playlist. Only owner or admin.
func (h *PlaylistHandler) deletePlaylist(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	p, err := h.client.Playlist.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}
	isOwner := p.UserID == int(claims.UserID)
	isAdmin := claims.IsStaff

	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only delete own playlist"})
		return
	}

	err = h.client.Playlist.DeleteOneID(id).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// addMedia adds a media item to a playlist.
func (h *PlaylistHandler) addMedia(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	var input struct {
		MediaID int `json:"media_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Verify ownership
	p, err := h.client.Playlist.Get(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}
	if p.UserID != int(claims.UserID) && !claims.IsStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only modify own playlist"})
		return
	}

	// Verify media exists
	_, err = h.client.Media.Get(ctx, input.MediaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "media not found"})
		return
	}

	// Check if already in playlist
	existsMP, _ := h.client.MediaPlaylist.Query().
		Where(
			mediaplaylist.HasPlaylistWith(pl.IDEQ(id)),
			mediaplaylist.HasMediaWith(entitymedia.IDEQ(input.MediaID)),
		).Exist(ctx)
	if existsMP {
		c.JSON(http.StatusConflict, gin.H{"error": "media already in playlist"})
		return
	}

	// Get current max ordering for this playlist
	maxOrder := 0
	mpList, _ := h.client.MediaPlaylist.Query().
		Where(mediaplaylist.HasPlaylistWith(pl.IDEQ(id))).
		All(ctx)
	for _, mp := range mpList {
		if mp.Ordering > maxOrder {
			maxOrder = mp.Ordering
		}
	}

	// Create MediaPlaylist junction record
	_, err = h.client.MediaPlaylist.Create().
		SetPlaylistID(id).
		AddMediaIDs(input.MediaID).
		SetOrdering(maxOrder + 1).
		Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add media: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "media added to playlist",
		"playlist_id": id,
		"media_id":    input.MediaID,
	})
}

// removeMedia removes a media item from a playlist.
func (h *PlaylistHandler) removeMedia(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	ctx := c.Request.Context()

	p, err := h.client.Playlist.Get(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}
	if p.UserID != int(claims.UserID) && !claims.IsStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only modify own playlist"})
		return
	}

	// Delete from MediaPlaylist junction table
	_, err = h.client.MediaPlaylist.Delete().
		Where(
			mediaplaylist.HasPlaylistWith(pl.IDEQ(id)),
			mediaplaylist.HasMediaWith(entitymedia.IDEQ(mediaId)),
		).Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove media: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "media removed from playlist",
		"playlist_id": id,
		"media_id":    mediaId,
	})
}
