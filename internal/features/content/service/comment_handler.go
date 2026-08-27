package service

import (
	"context"
	"strconv"
	"time"

	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	contentdal "origadmin/application/origstudio/internal/features/content/dal"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

// CommentHandler handles comment-related HTTP endpoints.
type CommentHandler struct {
	commentQS     *contentdal.CommentQueryService
	jwtMgr        *auth.Manager
	commentLikeUC *contentbiz.CommentLikeUseCase
	moderationUC  *contentbiz.CommentModerationUseCase
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(
	commentQS *contentdal.CommentQueryService,
	jwtMgr *auth.Manager,
	commentLikeUC *contentbiz.CommentLikeUseCase,
	moderationUC *contentbiz.CommentModerationUseCase,
) *CommentHandler {
	return &CommentHandler{
		commentQS:     commentQS,
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
		publicComments.GET("", server.WithOptionalJWTCtx(h.jwtMgr, h.listComments))
		publicComments.GET("/:id", server.WithOptionalJWTCtx(h.jwtMgr, h.getComment))
	}

	// Authenticated routes (JWT required for write operations)
	authComments := r.Group("/comments")
	authComments.Use(server.JWTMiddlewareCtx(h.jwtMgr))
	{
		authComments.POST("", h.createComment)
		authComments.PUT("/:id", h.updateComment)
		authComments.DELETE("/:id", h.deleteComment)
	}

	// Register routes at top-level for media-scoped comments.
	// NOTE: Use explicit r.GET / r.POST (NOT r.Group + "" suffix) because route
	// adapters may treat "/medias/:token/comments" + "" as "/medias/:token/comments/"
	// (trailing slash) which fails to match POST /api/v1/medias/<id>/comments.
	// Media-scoped comment creation (authenticated) lives at a dedicated
	// path (POST /api/v1/medias/:token/comments) to avoid route-shadowing against
	// the gRPC-gateway POST /api/v1/comments endpoint, which expects a flat
	// protobuf payload with int64 media_id and cannot process UUID media ids.
	r.GET("/medias/:token/comments", server.WithOptionalJWTCtx(h.jwtMgr, h.listMediaComments))
	r.POST("/medias/:token/comments", server.WithJWTCtx(h.jwtMgr, h.createMediaComment))

	// Register Comment Likes routes
	h.registerCommentLikesRoutes(r)
}

func (h *CommentHandler) listComments(ctx http2.Context) error {
	// 兼容多种命名：老内容用 snake_case(media_id/content_id, int64)，
	// 新 UUID/shortToken 内容前端只发 camelCase(mediaId/contentId)。
	// 原实现只读 media_id，导致 UUID 内容过滤失效、返回全站评论(BUG-156)。
	mediaID := ctx.QueryVar("media_id")
	if mediaID == "" {
		mediaID = ctx.QueryVar("mediaId")
	}
	if mediaID == "" {
		mediaID = ctx.QueryVar("contentId")
	}
	if mediaID == "" {
		mediaID = ctx.QueryVar("content_id")
	}
	userID := ctx.QueryVar("user_id")
	parentID := ctx.QueryVar("parent_id")
	rootOnly := ctx.QueryVar("root_only") == "true" || ctx.QueryVar("root_only") == "1"
	sortBy := ctx.QueryVarDefault("sort_by", "create_time")
	order := ctx.QueryVarDefault("order", "desc")
	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))

	var currentUserID string
	isAdmin := false
	if claims, ok := server.GetClaimsCtx(ctx); ok {
		currentUserID = claims.GetUserID()
		isAdmin = claims.IsAdmin()
	}

	// Resolve media_id (supports both UUID and short_token)
	if mediaID != "" {
		mediaID = h.resolveMediaID(ctx, mediaID)
	}

	// Normalize pagination parameters
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

	result, err := h.commentQS.ListComments(ctx, contentdal.CommentListParams{
		MediaID:  mediaID,
		UserID:   userID,
		ParentID: parentID,
		RootOnly: rootOnly,
		SortBy:   sortBy,
		Order:    order,
		Page:     page,
		PageSize: pageSize,
		IsAdmin:  isAdmin,
	})
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to fetch comments")
	}

	comments := make([]map[string]any, len(result.Items))
	for i, item := range result.Items {
		comments[i] = convertCommentListItemToResponse(item, currentUserID, h.commentLikeUC, ctx)
	}

	server.PageCtx(ctx, comments, int64(result.Total), page, pageSize)
	return nil
}

func (h *CommentHandler) listMediaComments(ctx http2.Context) error {
	mediaToken := ctx.Var("token")
	// Resolve short_token to UUID if needed
	mediaID := h.resolveMediaID(ctx, mediaToken)
	q := ctx.Request().URL.Query()
	q.Set("media_id", mediaID)
	ctx.Request().URL.RawQuery = q.Encode()
	return h.listComments(ctx)
}

// resolveMediaID resolves a media ID or short_token to the internal UUID.
// If resolution fails, returns the original idOrToken as-is.
func (h *CommentHandler) resolveMediaID(ctx context.Context, idOrToken string) string {
	if h.commentQS == nil {
		return idOrToken
	}
	return h.commentQS.ResolveMediaID(ctx, idOrToken)
}

// checkMediaExists verifies that a media exists by UUID.
func (h *CommentHandler) checkMediaExists(ctx context.Context, mediaID string) error {
	if h.commentQS == nil {
		return nil
	}
	return h.commentQS.CheckMediaExists(ctx, mediaID)
}

// incrementCommentCount updates the media's comment count by delta.
func (h *CommentHandler) incrementCommentCount(ctx context.Context, mediaID string, delta int) {
	if h.commentQS != nil && mediaID != "" {
		_ = h.commentQS.IncrementCommentCount(ctx, mediaID, delta)
	}
}

// createMediaComment creates a comment scoped to a media id or short token from
// the URL path :token variable. It mirrors createComment but infers the media
// id from the URL (supporting both UUIDs and short tokens) so the request body
// only needs to carry {comment:{content, parent_id}}. This route lives at a
// dedicated path to avoid shadowing conflicts with the gRPC-gateway endpoint.
//
// Body compatibility (dual-format parsing so both HTTP-native and gRPC-style
// payloads are accepted):
//   - HTTP handler format: {comment:{content|text, parent_id}}
//   - gRPC flat format:    {text|content, parent_id}
func (h *CommentHandler) createMediaComment(ctx http2.Context) error {
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		return server.FailCtx(ctx, 401, "Authentication required")
	}
	claims := claimsVal.(*auth.Claims)
	mediaToken := ctx.Var("token")

	// Resolve short_token to UUID
	mediaID := h.resolveMediaID(ctx, mediaToken)

	// Dual-format struct: captures both nested {comment} wrapper and flat
	// top-level fields so either payload shape binds correctly.
	var input struct {
		// Nested format: {comment:{content|text, parent_id}}  (HTTP handler)
		Comment *struct {
			Content  string `json:"content"`
			Text     string `json:"text"`
			ParentID string `json:"parent_id,omitempty"`
		} `json:"comment,omitempty"`
		// Flat format: {text|content, parent_id}  (gRPC-style / legacy clients)
		Text     string `json:"text"`
		Content  string `json:"content"`
		ParentID string `json:"parent_id,omitempty"`
	}
	if err := ctx.Bind(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, "Invalid request body: "+err.Error())
	}

	// Merge fields from either format. Nested {comment.*} takes precedence when
	// both shapes are present in the same payload.
	content := ""
	parentID := ""
	if input.Comment != nil {
		if input.Comment.Content != "" {
			content = input.Comment.Content
		} else if input.Comment.Text != "" {
			content = input.Comment.Text
		}
		parentID = input.Comment.ParentID
	}
	if content == "" {
		if input.Content != "" {
			content = input.Content
		} else if input.Text != "" {
			content = input.Text
		}
	}
	if parentID == "" {
		parentID = input.ParentID
	}
	if content == "" {
		return server.FailCtx(ctx, 400, "comment text cannot be empty")
	}

	// Verify media exists
	if mediaID != "" {
		if err := h.checkMediaExists(ctx, mediaID); err != nil {
			return server.FailCtx(ctx, 404, "media not found")
		}
	}

	initialStatus := "APPROVED"
	if h.moderationUC != nil {
		initialStatus = h.moderationUC.GetInitialStatus(ctx)
	}

	item, err := h.commentQS.CreateComment(ctx, contentdal.CreateCommentParams{
		Content:  content,
		UserID:   claims.GetUserID(),
		MediaID:  mediaID,
		ParentID: parentID,
		Status:   initialStatus,
	})
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "Failed to create comment: "+err.Error())
	}

	// Update media comment count
	h.incrementCommentCount(ctx, mediaID, 1)

	return server.CreatedCtx(ctx, convertCommentListItemToResponse(item, claims.GetUserID(), h.commentLikeUC, ctx))
}

func (h *CommentHandler) getComment(ctx http2.Context) error {
	id := ctx.Var("id")

	item, err := h.commentQS.GetComment(ctx, id)
	if err != nil {
		return server.FailCtx(ctx, 404, "Comment not found")
	}

	var currentUserID string
	if claims, ok := server.GetClaimsCtx(ctx); ok {
		currentUserID = claims.GetUserID()
	}

	return server.OKCtx(ctx, convertCommentListItemToResponse(item, currentUserID, h.commentLikeUC, ctx))
}

func (h *CommentHandler) createComment(ctx http2.Context) error {
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		return server.FailCtx(ctx, 401, "Authentication required")
	}
	claims := claimsVal.(*auth.Claims)

	// Dual-format struct. Accepts both the HTTP-native nested wrapper and the
	// gRPC-gateway flat payload so requests routed to either handler (or bound
	// via fallback proto endpoints) parse successfully regardless of which
	// middleware matched first.
	//
	// Media-id resolution supports all four naming variants (snake/camel ×
	// media/content) because the frontend passes each under both conventions.
	var input struct {
		// Nested format: {comment:{content|text, media_id|mediaId|content_id|contentId, parent_id}}
		Comment *struct {
			Content   string `json:"content"`
			Text      string `json:"text"`
			MediaID   string `json:"media_id,omitempty"`
			MediaId   string `json:"mediaId,omitempty"`
			ContentID string `json:"content_id,omitempty"`
			ContentId string `json:"contentId,omitempty"`
			ParentID  string `json:"parent_id,omitempty"`
		} `json:"comment,omitempty"`
		// Flat format: {text|content, media_id|mediaId|content_id|contentId, parent_id}
		Text      string `json:"text"`
		Content   string `json:"content"`
		MediaID   string `json:"media_id,omitempty"`
		MediaId   string `json:"mediaId,omitempty"`
		ContentID string `json:"content_id,omitempty"`
		ContentId string `json:"contentId,omitempty"`
		ParentID  string `json:"parent_id,omitempty"`
	}
	if err := ctx.Bind(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, "Invalid request body: "+err.Error())
	}

	// Merge content from either shape and either field name. Nested wins when
	// both are supplied (clients are expected to pick one shape consistently,
	// but we tolerate ambiguity by preferring the explicit comment wrapper).
	content := ""
	parentID := ""
	mediaID := ""
	if input.Comment != nil {
		if input.Comment.Content != "" {
			content = input.Comment.Content
		} else if input.Comment.Text != "" {
			content = input.Comment.Text
		}
		parentID = input.Comment.ParentID
		if mediaID == "" {
			switch {
			case input.Comment.MediaID != "":
				mediaID = input.Comment.MediaID
			case input.Comment.MediaId != "":
				mediaID = input.Comment.MediaId
			case input.Comment.ContentID != "":
				mediaID = input.Comment.ContentID
			case input.Comment.ContentId != "":
				mediaID = input.Comment.ContentId
			}
		}
	}
	if content == "" {
		if input.Content != "" {
			content = input.Content
		} else if input.Text != "" {
			content = input.Text
		}
	}
	if parentID == "" {
		parentID = input.ParentID
	}
	if mediaID == "" {
		switch {
		case input.MediaID != "":
			mediaID = input.MediaID
		case input.MediaId != "":
			mediaID = input.MediaId
		case input.ContentID != "":
			mediaID = input.ContentID
		case input.ContentId != "":
			mediaID = input.ContentId
		}
	}
	if content == "" {
		return server.FailCtx(ctx, 400, "comment text cannot be empty")
	}

	// Resolve media_id (supports both UUID and short_token)
	if mediaID != "" {
		mediaID = h.resolveMediaID(ctx, mediaID)
	}

	// Verify media exists
	if mediaID != "" {
		if err := h.checkMediaExists(ctx, mediaID); err != nil {
			return server.FailCtx(ctx, 404, "media not found")
		}
	}

	initialStatus := "APPROVED"
	if h.moderationUC != nil {
		initialStatus = h.moderationUC.GetInitialStatus(ctx)
	}

	item, err := h.commentQS.CreateComment(ctx, contentdal.CreateCommentParams{
		Content:  content,
		UserID:   claims.GetUserID(),
		MediaID:  mediaID,
		ParentID: parentID,
		Status:   initialStatus,
	})
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "Failed to create comment: "+err.Error())
	}

	// Update media comment count
	h.incrementCommentCount(ctx, mediaID, 1)

	return server.CreatedCtx(ctx, convertCommentListItemToResponse(item, claims.GetUserID(), h.commentLikeUC, ctx))
}

func (h *CommentHandler) updateComment(ctx http2.Context) error {
	id := ctx.Var("id")

	var input struct {
		Comment struct {
			Content string `json:"content,omitempty"`
			Status  string `json:"status,omitempty"`
		} `json:"comment"`
	}
	if err := ctx.Bind(&input); err != nil {
		return server.FailCtx(ctx, 400, "Invalid request body")
	}

	item, err := h.commentQS.UpdateComment(ctx, id, contentdal.UpdateCommentParams{
		Content: input.Comment.Content,
		Status:  input.Comment.Status,
	})
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to update comment")
	}

	return server.OKCtx(ctx, convertCommentListItemToResponse(item, "", h.commentLikeUC, ctx))
}

func (h *CommentHandler) deleteComment(ctx http2.Context) error {
	id := ctx.Var("id")

	// Get comment first to retrieve mediaID for count decrement
	comment, err := h.commentQS.GetComment(ctx, id)
	if err != nil {
		return server.FailCtx(ctx, 404, "Comment not found")
	}

	err = h.commentQS.DeleteComment(ctx, id)
	if err != nil {
		return server.FailCtx(ctx, 500, "Failed to delete comment")
	}

	// Decrement media comment count
	if comment.MediaID != "" {
		h.incrementCommentCount(ctx, comment.MediaID, -1)
	}

	return server.OKCtx(ctx, nil)
}

func (h *CommentHandler) registerCommentLikesRoutes(r http2.Router) {
	commentLikes := r.Group("/comments/:id")
	{
		commentLikes.GET("/likes", server.WithOptionalJWTCtx(h.jwtMgr, func(ctx http2.Context) error {
			commentID := ctx.Var("id")
			if commentID == "" {
				return server.FailCtx(ctx, 400, "comment ID required")
			}

			userID := ""
			if claims, ok := server.GetClaimsCtx(ctx); ok {
				userID = claims.GetUserID()
			}

			stats, err := h.commentLikeUC.GetStats(ctx, userID, commentID)
			if err != nil {
				return server.FailCtx(ctx, 500, "failed to get comment likes")
			}

			return server.OKCtx(ctx, stats)
		}))

		commentLikes.POST("/likes", server.WithJWTCtx(h.jwtMgr, func(ctx http2.Context) error {
			commentID := ctx.Var("id")
			if commentID == "" {
				return server.FailCtx(ctx, 400, "comment ID required")
			}

			userID := ""
			if claims, ok := server.GetClaimsCtx(ctx); ok {
				userID = claims.GetUserID()
			}
			if userID == "" {
				return server.FailCtx(ctx, 401, "unauthorized")
			}

			stats, err := h.commentLikeUC.ToggleLike(ctx, userID, commentID)
			if err != nil {
				return server.FailCtx(ctx, 500, "failed to toggle like")
			}

			return server.OKCtx(ctx, stats)
		}))

		commentLikes.POST("/dislikes", server.WithJWTCtx(h.jwtMgr, func(ctx http2.Context) error {
			commentID := ctx.Var("id")
			if commentID == "" {
				return server.FailCtx(ctx, 400, "comment ID required")
			}

			userID := ""
			if claims, ok := server.GetClaimsCtx(ctx); ok {
				userID = claims.GetUserID()
			}
			if userID == "" {
				return server.FailCtx(ctx, 401, "unauthorized")
			}

			stats, err := h.commentLikeUC.ToggleDislike(ctx, userID, commentID)
			if err != nil {
				return server.FailCtx(ctx, 500, "failed to toggle dislike")
			}

			return server.OKCtx(ctx, stats)
		}))
	}
}

func convertCommentListItemToResponse(item *contentdal.CommentListItem, currentUserID string, commentLikeUC *contentbiz.CommentLikeUseCase, ctx context.Context) map[string]any {
	var likeCount int64
	var isLiked bool
	if commentLikeUC != nil && item.ID != "" {
		stats, err := commentLikeUC.GetStats(ctx, currentUserID, item.ID)
		if err == nil && stats != nil {
			likeCount = stats.LikeCount
			isLiked = stats.IsLiked
		}
	}

	resp := map[string]any{
		"id":          item.ID,
		"content":     item.Content,
		"status":      item.Status,
		"create_time": item.CreateTime.Format(time.RFC3339),
		"update_time": item.UpdateTime.Format(time.RFC3339),
		"like_count":  likeCount,
		"is_liked":    isLiked,
		"is_reply":    item.IsReply,
	}

	if item.MediaID != "" {
		resp["media_id"] = item.MediaID
	}
	if item.UserID != "" {
		resp["user_id"] = item.UserID
	}
	if item.Username != "" {
		resp["username"] = item.Username
	}
	if item.Avatar != "" {
		resp["avatar"] = item.Avatar
	}
	if item.ParentID != "" {
		resp["reply_to_comment_id"] = item.ParentID
		resp["reply_to_content"] = item.ParentContent
		if item.ParentUsername != "" {
			resp["reply_to_username"] = item.ParentUsername
		}
	} else {
		resp["parent_id"] = nil
	}

	return resp
}
