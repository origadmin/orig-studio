package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
)

// PlaylistHandler handles /api/v1/playlists routes.
type PlaylistHandler struct {
	uc  *biz.PlaylistChannelUseCase
	jwt *auth.Manager
}

// NewPlaylistHandler creates a new PlaylistHandler.
func NewPlaylistHandler(uc *biz.PlaylistChannelUseCase, jwt *auth.Manager) *PlaylistHandler {
	return &PlaylistHandler{uc: uc, jwt: jwt}
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListPlaylists(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	p, err := h.uc.GetPlaylist(c.Request.Context(), id)
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListUserPlaylists(c.Request.Context(), int(claims.UserID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  items,
		"total": total,
	})
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
		Name        string `json:"name" binding:"required,max=100"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := &biz.Playlist{
		Name:        input.Name,
		Description: input.Description,
		UserID:      int(claims.UserID),
		IsPublic:    input.IsPublic,
	}

	created, err := h.uc.CreatePlaylist(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
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

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := &biz.Playlist{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		IsPublic:    input.IsPublic,
	}

	updated, err := h.uc.UpdatePlaylist(c.Request.Context(), p, int(claims.UserID), claims.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
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

	err = h.uc.DeletePlaylist(c.Request.Context(), id, int(claims.UserID), claims.IsStaff)
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

	err = h.uc.AddMediaToPlaylist(c.Request.Context(), id, input.MediaID, int(claims.UserID), claims.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	_, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}
	_, err = strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	// removeMedia not fully implemented in UseCase yet
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented in UseCase"})
}
