package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/comment"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
)

func RegisterCommentRoutes(group *gin.RouterGroup, client *entity.Client, jwtMgr *auth.Manager, commentLikeUC *contentbiz.CommentLikeUseCase, moderationUC *contentbiz.CommentModerationUseCase) {
	// Public routes (no auth required)
	publicComments := group.Group("/comments")
	{
		// GET /comments - List comments with filtering and pagination (PUBLIC)
		publicComments.GET("", OptionalJWTMiddleware(jwtMgr), func(c *gin.Context) {
			ctx := c.Request.Context()

			mediaID := c.Query("media_id")
			userID := c.Query("user_id")
			parentID := c.Query("parent_id")
			sortBy := c.DefaultQuery("sort_by", "created_at")
			order := c.DefaultQuery("order", "desc")
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			var currentUserID string
			if claims, ok := GetClaims(c); ok {
				currentUserID = claims.GetUserID()
			}

			if page < 1 {
				page = 1
			}
			if pageSize < 1 || pageSize > 100 {
				pageSize = 20
			}

			query := client.Comment.Query()
			if mediaID != "" {
				query = query.Where(comment.MediaID(mediaID))
			}
			if userID != "" {
				query = query.Where(comment.UserID(userID))
			}
			if parentID != "" {
				query = query.Where(comment.HasParentWith(comment.ID(parentID)))
			}

			if claims, ok := GetClaims(c); ok && (claims.IsStaff || claims.Role == "admin") {
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
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to count comments"})
				return
			}

			switch sortBy {
			case "like_count":
				if order == "asc" {
					query = query.Order(entity.Asc(comment.FieldAddDate))
				} else {
					query = query.Order(entity.Desc(comment.FieldAddDate))
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
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to fetch comments"})
				return
			}

			comments := make([]gin.H, len(items))
			for i, item := range items {
				comments[i] = convertCommentToResponse(item, currentUserID, commentLikeUC, ctx)
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"items":     comments,
					"total":     total,
					"page":      page,
					"page_size": pageSize,
				},
			})
		})

		// GET /comments/:id - Get single comment (PUBLIC)
		publicComments.GET("/:id", OptionalJWTMiddleware(jwtMgr), func(c *gin.Context) {
			ctx := c.Request.Context()
			id := c.Param("id")

			commentObj, err := client.Comment.Get(ctx, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Comment not found"})
				return
			}

			var currentUserID string
			if claims, ok := GetClaims(c); ok {
				currentUserID = claims.GetUserID()
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    convertCommentToResponse(commentObj, currentUserID, commentLikeUC, ctx),
			})
		})
	}

	// Authenticated routes (JWT required for write operations)
	authComments := group.Group("/comments")
	authComments.Use(JWTMiddleware(jwtMgr))
	{
		// POST /comments - Create comment (AUTH REQUIRED)
		authComments.POST("", func(c *gin.Context) {
			ctx := c.Request.Context()

			claimsVal, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Authentication required"})
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
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body: " + err.Error()})
				return
			}

			if input.Comment.Content == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Content is required"})
				return
			}

			createBuilder := client.Comment.Create().
				SetText(input.Comment.Content).
				SetUserID(claims.GetUserID()).
				SetStatus(comment.Status(moderationUC.GetInitialStatus(ctx)))

			if input.Comment.MediaID != "" {
				createBuilder = createBuilder.SetMediaID(input.Comment.MediaID)
			}
			if input.Comment.ParentID != "" {
				createBuilder = createBuilder.SetParentID(input.Comment.ParentID)
			}

			commentObj, err := createBuilder.Save(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create comment: " + err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"code":    0,
				"message": "success",
				"data":    gin.H{"comment": convertCommentToResponse(commentObj, claims.GetUserID(), commentLikeUC, ctx)},
			})
		})

		// PUT /comments/:id - Update comment (AUTH REQUIRED)
		authComments.PUT("/:id", func(c *gin.Context) {
			ctx := c.Request.Context()
			id := c.Param("id")

			var input struct {
				Comment struct {
					Content string `json:"content,omitempty"`
					Status  string `json:"status,omitempty"`
				} `json:"comment"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
				return
			}

			updateBuilder := client.Comment.UpdateOneID(id)
			if input.Comment.Content != "" {
				updateBuilder = updateBuilder.SetText(input.Comment.Content)
			}
			if input.Comment.Status != "" {
				updateBuilder = updateBuilder.SetStatus(comment.Status(input.Comment.Status))
			}

			commentObj, err := updateBuilder.Save(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to update comment"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    gin.H{"comment": convertCommentToResponse(commentObj, "", commentLikeUC, ctx)},
			})
		})

		// DELETE /comments/:id - Delete comment (AUTH REQUIRED)
		authComments.DELETE("/:id", func(c *gin.Context) {
			ctx := c.Request.Context()
			id := c.Param("id")

			err := client.Comment.DeleteOneID(id).Exec(ctx)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Comment not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
		})
	}

	// Register Comment Likes routes (stub)
	registerCommentLikesRoutes(group, jwtMgr, commentLikeUC)
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

func registerCommentLikesRoutes(group *gin.RouterGroup, jwtMgr *auth.Manager, commentLikeUC *contentbiz.CommentLikeUseCase) {
	commentLikes := group.Group("/comments/:id")
	{
		commentLikes.GET("/likes", func(c *gin.Context) {
			commentID := c.Param("id")
			if commentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "comment ID required"})
				return
			}

			userID := ""
			if claims, ok := GetClaims(c); ok {
				userID = claims.GetUserID()
			}

			stats, err := commentLikeUC.GetStats(c.Request.Context(), userID, commentID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get comment likes"})
				return
			}

			OK(c, stats)
		})

		commentLikes.POST("/likes", JWTMiddleware(jwtMgr), func(c *gin.Context) {
			commentID := c.Param("id")
			if commentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "comment ID required"})
				return
			}

			userID := ""
			if claims, ok := GetClaims(c); ok {
				userID = claims.GetUserID()
			}
			if userID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized"})
				return
			}

			stats, err := commentLikeUC.ToggleLike(c.Request.Context(), userID, commentID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to toggle like"})
				return
			}

			OK(c, stats)
		})

		commentLikes.POST("/dislikes", JWTMiddleware(jwtMgr), func(c *gin.Context) {
			commentID := c.Param("id")
			if commentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "comment ID required"})
				return
			}

			userID := ""
			if claims, ok := GetClaims(c); ok {
				userID = claims.GetUserID()
			}
			if userID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized"})
				return
			}

			stats, err := commentLikeUC.ToggleDislike(c.Request.Context(), userID, commentID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to toggle dislike"})
				return
			}

			OK(c, stats)
		})
	}
}
