package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
)

// CommentHandler handles /api/v1/comments routes.
type CommentHandler struct {
	uc  *biz.CommentUseCase
	jwt *auth.Manager
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(uc *biz.CommentUseCase, jwt *auth.Manager) *CommentHandler {
	return &CommentHandler{uc: uc, jwt: jwt}
}

func (h *CommentHandler) Register(group *gin.RouterGroup) {
	comments := group.Group("/comments")
	{
		// Public read routes (media comments visible to all)
		comments.GET("", h.listComments)

		// Protected write routes
		protected := comments.Group("")
		protected.Use(JWTMiddleware(h.jwt))
		{
			protected.POST("", h.createComment)
			protected.PUT("/:id", h.updateComment)
			protected.DELETE("/:id", h.deleteComment)
			protected.GET("/media/:mediaId", h.listMediaComments)
		}
	}
}

// listComments returns all comments with optional media_id filter and pagination.
// GET /comments?media_id=123&page=1&page_size=20
func (h *CommentHandler) listComments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	mediaID, _ := strconv.Atoi(c.Query("media_id"))

	var items []*biz.Comment
	var total int
	var err error

	if mediaID > 0 {
		items, total, err = h.uc.ListMediaComments(c.Request.Context(), mediaID, page, limit)
	} else {
		// items, total, err = h.uc.ListAll(c.Request.Context(), page, limit)
		// For now, listComments without media_id is not fully implemented in UseCase
		c.JSON(http.StatusBadRequest, gin.H{"error": "media_id is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// listMediaComments returns comments for a specific media with nested replies.
// GET /comments/media/:mediaId
func (h *CommentHandler) listMediaComments(c *gin.Context) {
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	items, total, err := h.uc.ListMediaComments(c.Request.Context(), mediaId, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":     items,
		"total":    total,
		"media_id": mediaId,
	})
}

// createComment creates a new comment.
// POST body: {"text": string, "media_id": int, "parent_id": int (optional)}
func (h *CommentHandler) createComment(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		Text     string `json:"text" binding:"required"`
		MediaID  int    `json:"media_id" binding:"required"`
		ParentID *int   `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := &biz.Comment{
		Text:     input.Text,
		MediaID:  input.MediaID,
		UserID:   int(claims.UserID),
		ParentID: input.ParentID,
	}

	created, err := h.uc.CreateComment(c.Request.Context(), comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// updateComment updates a comment text. Only author or admin can update.
func (h *CommentHandler) updateComment(c *gin.Context) {
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
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.uc.UpdateComment(
		c.Request.Context(),
		id,
		int(claims.UserID),
		claims.IsStaff,
		input.Text,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// deleteComment deletes a comment. Only author or admin can delete.
func (h *CommentHandler) deleteComment(c *gin.Context) {
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

	err = h.uc.DeleteComment(c.Request.Context(), id, int(claims.UserID), claims.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
