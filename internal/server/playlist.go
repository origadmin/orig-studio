package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
)

func RegisterPlaylistRoutes(group *gin.RouterGroup, client *entity.Client) {
	playlists := group.Group("/playlists")
	{
		playlists.GET("", func(c *gin.Context) {
			items, err := client.Playlist.Query().All(c.Request.Context())
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}
			OK(c, gin.H{
				"items":     items,
				"total":     len(items),
				"page":      1,
				"page_size": len(items),
			})
		})

		playlists.POST("", func(c *gin.Context) {
			var input struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				UserID      string `json:"user_id"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			p, err := client.Playlist.Create().
				SetTitle(input.Title).
				SetDescription(input.Description).
				SetUserID(input.UserID).
				Save(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, p)
		})
	}
}
