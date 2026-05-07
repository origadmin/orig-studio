package service

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/comment"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
)

// CommentHandler handles comment-related HTTP endpoints.
type CommentHandler struct {
	client        *entity.Client
	jwtMgr        *auth.Manager
	commentLikeUC *contentbiz.CommentLikeUseCase
	moderationUC  *contentbiz.CommentModerationUseCase
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(
	client *entity.Client,
	jwtMgr *auth.Manager,
	commentLikeUC *contentbiz.CommentLikeUseCase,
	moderationUC *contentbiz.CommentModerationUseCase,
) *CommentHandler {
	return &CommentHandler{
		client:        client,
		jwtMgr:        jwtMgr,
		commentLikeUC: commentLikeUC,
		moderationUC:  moderationUC,
	}
}

// RegisterRoutes registers the handler's routes.
func (h *CommentHandler) RegisterRoutes(r http2.Router) {
	// Public routes (no auth required)
	publicComments := r.Group("/comments")
	{
		// GET /comments - List comments with filtering and pagination (PUBLIC)
		publicComments.GET("", server.WithOptionalJWTCtx(h.jwtMgr, server.GinHandlerToHandlerFunc(h.listComments)))

		// GET /comments/:id - Get single comment (PUBLIC)
		publicComments.GET("/:id", server.WithOptionalJWTCtx(h.jwtMgr, server.GinHandlerToHandlerFunc(h.getComment)))
	}

	// Authenticated routes (JWT required for write operations)
	authComments := r.Group("/comments")
	authComments.Use(server.JWTMiddlewareCtx(h.jwtMgr))
	{
		// POST /comments - Create comment (AUTH REQUIRED)
		authComments.POST("", server.GinHandlerToHandlerFunc(h.createComment))

		// PUT /comments/:id - Update comment (AUTH REQUIRED)
		authComments.PUT("/:id", server.GinHandlerToHandlerFunc(h.updateComment))

		// DELETE /comments/:id - Delete comment (AUTH REQUIRED)
		authComments.DELETE("/:id", server.GinHandlerToHandlerFunc(h.deleteComment))
	}

	// Register Comment Likes routes
	h.registerCommentLikesRoutes(r)
}

func (h *CommentHandler) listComments(c *gin.Context) {
	ctx := c.Request.Context()

	mediaID := c.Query("media_id")
	userID := c.Query("user_id")
	parentID := c.Query("parent_id")
	sortBy := c.DefaultQuery("sort_by", "create_time")
	order := c.DefaultQuery("order", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var currentUserID string
	if claims, ok := server.GetClaims(c); ok {
		currentUserID = claims.GetUserID()
	}

	// Normalize pagination parameters
	page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

	query := h.client.Comment.Query()
	if mediaID != "" {
		query = query.Where(comment.MediaID(mediaID))
	}
	if userID != "" {
		query = query.Where(comment.UserID(userID))
	}
	if parentID != "" {
		query = query.Where(comment.HasParentWith(comment.ID(parentID)))
	}

	if claims, ok := server.GetClaims(c); ok && (claims.IsStaff || claims.Role == "admin") {
	} else if currentUserID != "" {
		query = query.Where(comment.Or(
			comment.StatusEQ(comment.StatusAPPROVED),
			comment.UserID(currentUserID),
		))
	} else {
		query = query.Where(comment.StatusEQ(comment.StatusAPPROVED))
	}

	total, err := query.Count(ctx)
	if err != nil {
		server.Fail(c, 500, "Failed to count comments")
		return
	}

	switch sortBy {
	case "like_count":
		if order == "asc" {
			query = query.Order(entity.Asc(comment.FieldLikeCount))
		} else {
			query = query.Order(entity.Desc(comment.FieldLikeCount))
		}
	default:
		if order == "asc" {
			query = query.Order(entity.Asc(comment.FieldAddDate))
		} else {
			query = query.Order(entity.Desc(comment.FieldAddDate))
		}
	}

	items, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		WithUser().
		WithParent(func(pq *entity.CommentQuery) {
			pq.WithUser()
		}).
		All(ctx)
	if err != nil {
		server.Fail(c, 500, "Failed to fetch comments")
		return
	}

	comments := make([]gin.H, len(items))
	for i, item := range items {
		comments[i] = convertCommentToResponse(item, currentUserID, h.commentLikeUC, ctx)
	}

	server.Page(c, comments, int64(total), page, pageSize)
}

func (h *CommentHandler) getComment(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	commentObj, err := h.client.Comment.Get(ctx, id)
	if err != nil {
		server.Fail(c, 404, "Comment not found")
		return
	}

	var currentUserID string
	if claims, ok := server.GetClaims(c); ok {
		currentUserID = claims.GetUserID()
	}

	server.OK(c, convertCommentToResponse(commentObj, currentUserID, h.commentLikeUC, ctx))
}

func (h *CommentHandler) createComment(c *gin.Context) {
	ctx := c.Request.Context()

	claimsVal, exists := c.Get("claims")
	if !exists {
		server.Fail(c, 401, "Authentication required")
		return
	}
	claims := claimsVal.(*auth.Claims)

	var input struct {
		Comment struct {
			Content  string `json:"content"`
			MediaID  string `json:"media_id,omitempty"`
			ParentID string `json:"parent_id,omitempty"`
		} `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, server.ErrBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if input.Comment.Content == "" {
		server.Fail(c, 400, "Content is required")
		return
	}

	createBuilder := h.client.Comment.Create().
		SetText(input.Comment.Content).
		SetUserID(claims.GetUserID()).
		SetStatus(comment.Status(h.moderationUC.GetInitialStatus(ctx)))

	if input.Comment.MediaID != "" {
		createBuilder = createBuilder.SetMediaID(input.Comment.MediaID)
	}
	if input.Comment.ParentID != "" {
		createBuilder = createBuilder.SetParentID(input.Comment.ParentID)
	}

	commentObj, err := createBuilder.Save(ctx)
	if err != nil {
		server.Fail(c, server.ErrInternal, "Failed to create comment: "+err.Error())
		return
	}

	server.Created(c, convertCommentToResponse(commentObj, claims.GetUserID(), h.commentLikeUC, ctx))
}

func (h *CommentHandler) updateComment(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var input struct {
		Comment struct {
			Content string `json:"content,omitempty"`
			Status  string `json:"status,omitempty"`
		} `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		server.Fail(c, 400, "Invalid request body")
		return
	}

	updateBuilder := h.client.Comment.UpdateOneID(id)
	if input.Comment.Content != "" {
		updateBuilder = updateBuilder.SetText(input.Comment.Content)
	}
	if input.Comment.Status != "" {
		updateBuilder = updateBuilder.SetStatus(comment.Status(input.Comment.Status))
	}

	commentObj, err := updateBuilder.Save(ctx)
	if err != nil {
		server.Fail(c, 500, "Failed to update comment")
		return
	}

	server.OK(c, convertCommentToResponse(commentObj, "", h.commentLikeUC, ctx))
}

func (h *CommentHandler) deleteComment(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	err := h.client.Comment.DeleteOneID(id).Exec(ctx)
	if err != nil {
		server.Fail(c, 404, "Comment not found")
		return
	}

	server.OK(c, nil)
}

func (h *CommentHandler) registerCommentLikesRoutes(r http2.Router) {
	commentLikes := r.Group("/comments/:id")
	{
		commentLikes.GET("/likes", server.WithOptionalJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gc := ginadapter.GetGinContext(r)
			commentID := gc.Param("id")
			if commentID == "" {
				server.Fail(gc, 400, "comment ID required")
				return
			}

			userID := ""
			if claims, ok := server.GetClaims(gc); ok {
				userID = claims.GetUserID()
			}

			stats, err := h.commentLikeUC.GetStats(r.Context(), userID, commentID)
			if err != nil {
				server.Fail(gc, 500, "failed to get comment likes")
				return
			}

			server.OK(gc, stats)
		})))

		commentLikes.POST("/likes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gc := ginadapter.GetGinContext(r)
			commentID := gc.Param("id")
			if commentID == "" {
				server.Fail(gc, 400, "comment ID required")
				return
			}

			userID := ""
			if claims, ok := server.GetClaims(gc); ok {
				userID = claims.GetUserID()
			}
			if userID == "" {
				server.Fail(gc, 401, "unauthorized")
				return
			}

			stats, err := h.commentLikeUC.ToggleLike(r.Context(), userID, commentID)
			if err != nil {
				server.Fail(gc, 500, "failed to toggle like")
				return
			}

			server.OK(gc, stats)
		})))

		commentLikes.POST("/dislikes", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gc := ginadapter.GetGinContext(r)
			commentID := gc.Param("id")
			if commentID == "" {
				server.Fail(gc, 400, "comment ID required")
				return
			}

			userID := ""
			if claims, ok := server.GetClaims(gc); ok {
				userID = claims.GetUserID()
			}
			if userID == "" {
				server.Fail(gc, 401, "unauthorized")
				return
			}

			stats, err := h.commentLikeUC.ToggleDislike(r.Context(), userID, commentID)
			if err != nil {
				server.Fail(gc, 500, "failed to toggle dislike")
				return
			}

			server.OK(gc, stats)
		})))
	}
}

func convertCommentToResponse(item *entity.Comment, currentUserID string, commentLikeUC *contentbiz.CommentLikeUseCase, ctx context.Context) gin.H {
	var likeCount int64
	var isLiked bool
	if commentLikeUC != nil && item.ID != "" {
		stats, err := commentLikeUC.GetStats(ctx, currentUserID, item.ID)
		if err == nil && stats != nil {
			likeCount = stats.LikeCount
			isLiked = stats.IsLiked
		}
	}

	resp := gin.H{
		"id":          item.ID,
		"content":     item.Text,
		"status":      item.Status,
		"create_time": item.AddDate.Format(time.RFC3339),
		"update_time": item.AddDate.Format(time.RFC3339),
		"like_count":  likeCount,
		"is_liked":    isLiked,
		"is_reply":    item.Edges.Parent != nil,
	}

	if item.MediaID != "" {
		resp["media_id"] = item.MediaID
	}
	if item.UserID != "" {
		resp["user_id"] = item.UserID
	}
	if item.Edges.User != nil {
		u := item.Edges.User
		resp["username"] = u.Username
		if u.Logo != "" {
			resp["avatar"] = u.Logo
		}
	}
	if item.Edges.Parent != nil {
		p := item.Edges.Parent
		resp["reply_to_comment_id"] = p.ID
		resp["reply_to_content"] = truncateText(p.Text, 100)
		if p.Edges.User != nil {
			resp["reply_to_username"] = p.Edges.User.Username
		}
	} else {
		resp["parent_id"] = nil
	}

	return resp
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}
