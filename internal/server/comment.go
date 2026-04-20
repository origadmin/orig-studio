package server

import (
	"net/http"
	"strconv"

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
				All(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to fetch comments"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"total":     total,
					"comments":  items,
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

			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": commentObj})
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
				"data": gin.H{"comment": commentObj},
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
				"data": gin.H{"comment": commentObj},
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
}
