package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
)

type CommentHandler struct {
	client *entity.Client
}

func NewCommentHandler(client *entity.Client) *CommentHandler {
	return &CommentHandler{client: client}
}

func (h *CommentHandler) Register(group *gin.RouterGroup) {
	comments := group.Group("/comments")
	{
		comments.GET("", h.listComments())
		comments.POST("", h.createComment())
		comments.DELETE("/:id", h.deleteComment())
	}
}

func (h *CommentHandler) listComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := h.client.Comment.Query().All(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"list": items})
	}
}

func (h *CommentHandler) createComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Text    string `json:"text"`
			MediaID int    `json:"media_id"`
			UserID  int    `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		comment, err := h.client.Comment.Create().
			SetText(input.Text).
			SetMediaID(input.MediaID).
			SetUserID(input.UserID).
			Save(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, comment)
	}
}

func (h *CommentHandler) deleteComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		err = h.client.Comment.DeleteOneID(id).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
