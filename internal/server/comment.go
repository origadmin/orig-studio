package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/comment"
)

func RegisterCommentRoutes(group *gin.RouterGroup, client *entity.Client, jwtMgr *auth.Manager) {
	// Public routes (no auth required)
	publicComments := group.Group("/comments")
	{
		// GET /comments - List comments with filtering and pagination (PUBLIC)
		publicComments.GET("", func(c *gin.Context) {
			ctx := c.Request.Context()

			mediaID := c.Query("media_id")
			userID := c.Query("user_id")
			parentID := c.Query("parent_id")
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

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

			total, err := query.Count(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to count comments"})
				return
			}

			items, err := query.
				Limit(pageSize).
				Offset((page - 1) * pageSize).
				Order(entity.Desc(comment.FieldAddDate)).
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
				comments[i] = convertCommentToResponse(item)
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"total":     total,
					"comments":  comments,
					"page":      page,
					"page_size": pageSize,
				},
			})
		})

		// GET /comments/:id - Get single comment (PUBLIC)
		publicComments.GET("/:id", func(c *gin.Context) {
			ctx := c.Request.Context()
			id := c.Param("id")

			commentObj, err := client.Comment.Get(ctx, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Comment not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": convertCommentToResponse(commentObj),
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
				SetUserID(claims.UserID).
				SetStatus("PENDING")

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
				"data": gin.H{"comment": convertCommentToResponse(commentObj)},
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
				updateBuilder = updateBuilder.SetStatus(input.Comment.Status)
			}

			commentObj, err := updateBuilder.Save(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to update comment"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{"comment": convertCommentToResponse(commentObj)},
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
	registerCommentLikesRoutes(group, jwtMgr)
}

func convertCommentToResponse(item *entity.Comment) gin.H {
	resp := gin.H{
		"id":          item.ID,
		"content":     item.Text,
		"status":      item.Status,
		"create_time": item.AddDate.Format(time.RFC3339),
		"update_time": item.AddDate.Format(time.RFC3339),
		"like_count":  0,
		"is_liked":    false,
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

// Comment Likes Routes (Stub - TODO: Implement with database table)
func registerCommentLikesRoutes(group *gin.RouterGroup, jwtMgr *auth.Manager) {
	commentLikes := group.Group("/comments/:id")
	{
		// GET /comments/:id/likes - Get comment like status (PUBLIC)
		commentLikes.GET("/likes", func(c *gin.Context) {
			commentID := c.Param("id")
			if commentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "comment ID required"})
				return
			}

			// TODO: Replace with actual database query when comment_likes table is created
			// For now, return default values (stub implementation)
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"like_count": 0,
					"is_liked":   false,
					"is_disliked": false,
				},
			})
		})

		// POST /comments/:id/likes - Toggle like (AUTH REQUIRED)
		commentLikes.POST("/likes", JWTMiddleware(jwtMgr), func(c *gin.Context) {
			commentID := c.Param("id")
			if commentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "comment ID required"})
				return
			}

			// TODO: Implement actual toggle logic with database
			// For now, return stub response
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"like_count": 1,
					"is_liked":   true,
					"is_disliked": false,
				},
			})
		})

		// POST /comments/:id/dislikes - Toggle dislike (AUTH REQUIRED)
		commentLikes.POST("/dislikes", JWTMiddleware(jwtMgr), func(c *gin.Context) {
			commentID := c.Param("id")
			if commentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "comment ID required"})
				return
			}

			// TODO: Implement actual toggle logic with database
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"like_count": 0,
					"is_liked":   false,
					"is_disliked": true,
				},
			})
		})
	}
}
