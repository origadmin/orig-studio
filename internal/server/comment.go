package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
)

func RegisterCommentRoutes(group *gin.RouterGroup, client *entity.Client) {
	comments := group.Group("/comments")
	{
		comments.GET("", func(c *gin.Context) {
			// Simply list all comments for now
			items, err := client.Comment.Query().All(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"list": items})
		})

		comments.POST("", func(c *gin.Context) {
			var input struct {
				Text    string `json:"text"`
				MediaID int    `json:"media_id"`
				UserID  int    `json:"user_id"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			comment, err := client.Comment.Create().
				SetText(input.Text).
				SetMediaID(input.MediaID).
				SetUserID(input.UserID).
				Save(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, comment)
		})
	}
}
