package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	cm "origadmin/application/origcms/internal/data/entity/comment"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
)

// CommentHandler handles /api/v1/comments routes.
type CommentHandler struct {
	client *entity.Client
	jwt *auth.Manager
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(client *entity.Client, jwt *auth.Manager) *CommentHandler {
	return &CommentHandler{client: client, jwt: jwt}
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
	limit := 20
	page := 1
	if l := c.Query("page_size"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	offset := (page - 1) * limit

	query := h.client.Comment.Query().
		Limit(limit).
		Offset(offset).
		Order(entity.Desc(cm.FieldAddDate))

	// Filter by media_id if provided
	if mediaIdStr := c.Query("media_id"); mediaIdStr != "" {
		if mediaId, err := strconv.Atoi(mediaIdStr); err == nil {
			query.Where(cm.MediaIDEQ(mediaId))
		}
	}

	items, err := query.All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, _ := h.client.Comment.Query().Count(c.Request.Context())
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

	comments, err := h.client.Comment.Query().
		Where(cm.MediaIDEQ(mediaId)).
		WithUser().
		WithReplies(func(rq *entity.CommentQuery) {
			rq.WithUser()
		}).
		Order(entity.Asc(cm.FieldAddDate)).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, _ := h.client.Comment.Query().
		Where(cm.MediaIDEQ(mediaId)).
		Count(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"list":     comments,
		"total":    count,
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

	comment, err := h.client.Comment.Create().
		SetText(input.Text).
		SetMediaID(input.MediaID).
		SetUserID(int(claims.UserID)).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Handle nested reply: set parent edge if parent_id provided
	if input.ParentID != nil && *input.ParentID > 0 {
		parent, pErr := h.client.Comment.Get(c.Request.Context(), *input.ParentID)
		if pErr == nil && parent.MediaID == input.MediaID {
			h.client.Comment.UpdateOne(comment).AddReplies(parent).Exec(c.Request.Context())
		}
	}

	// Update media comment count
	h.client.Media.UpdateOneID(input.MediaID).
		AddCommentCount(1).
		Save(c.Request.Context())

	// Load user for response
	comment, _ = h.client.Comment.Query().
		Where(cm.IDEQ(comment.ID)).
		WithUser().
		Only(c.Request.Context())

	// Notify media owner about new comment (if commenter is not the owner)
	go func() {
		bgCtx := context.Background()
		media, mErr := h.client.Media.Get(bgCtx, input.MediaID)
		if mErr == nil && media.UserID != int(claims.UserID) {
			h.client.Notification.Create().
				SetAction("new_comment").
				SetNotify(false).
				SetMethod("comment").
				SetUserID(media.UserID).
				Save(bgCtx)
		}
	}()

	c.JSON(http.StatusCreated, comment)
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

	comment, err := h.client.Comment.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	isAuthor := comment.UserID == int(claims.UserID)
	isAdmin := claims.IsStaff

	if !isAuthor && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only update own comments"})
		return
	}

	var input struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err = h.client.Comment.UpdateOneID(id).
		SetText(input.Text).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	comment, _ = h.client.Comment.Query().
		Where(cm.IDEQ(id)).
		WithUser().
		Only(c.Request.Context())

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

	comment, err := h.client.Comment.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	isAuthor := comment.UserID == int(claims.UserID)
	isAdmin := claims.IsStaff

	if !isAuthor && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "can only delete own comments"})
		return
	}

	// Update media comment count
	if comment.MediaID > 0 {
		h.client.Media.UpdateOneID(comment.MediaID).
			AddCommentCount(-1).
			Save(c.Request.Context())
	}

	err = h.client.Comment.DeleteOneID(id).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
