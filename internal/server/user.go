package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-user/biz"
)

type UserHandler struct {
	uc  *biz.UserUseCase
	jwt *auth.Manager
}

func NewUserHandler(uc *biz.UserUseCase, jwt *auth.Manager) *UserHandler {
	return &UserHandler{uc: uc, jwt: jwt}
}

func (h *UserHandler) Register(group *gin.RouterGroup) {
	users := group.Group("/users")
	{
		// List users
		users.GET("", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

			items, total, err := h.uc.ListUsers(c.Request.Context())
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
		})

		// Get user by ID
		users.GET("/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			u, err := h.uc.GetUser(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, u)
		})

		// Create user
		users.POST("", func(c *gin.Context) {
			var input struct {
				Username string `json:"username" binding:"required"`
				Email    string `json:"email" binding:"required,email"`
				Password string `json:"password" binding:"required,min=6"`
				Name     string `json:"name"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			hashedPassword, _ := h.uc.HashPassword(input.Password)
			u, err := h.uc.CreateUser(c.Request.Context(), &types.User{
				Username: input.Username,
				Email:    input.Email,
				Nickname: input.Name, // Use Name as Nickname if Name is not in User
			}, hashedPassword)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, u)
		})

		// Delete user
		users.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			err = h.uc.DeleteUser(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})

		// Subscription routes
		users.GET("/:id/subscription", func(c *gin.Context) {
			userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			var currentUserId int64 = 0
			if claims, ok := c.Get("claims"); ok {
				currentUserId = claims.(*auth.Claims).UserID
			}

			isSubscribed := false
			if currentUserId > 0 && currentUserId != userId {
				isSubscribed = false
			}

			c.JSON(http.StatusOK, gin.H{
				"is_subscribed":    isSubscribed,
				"subscriber_count": 0,
			})
		})

		users.POST("/:id/subscribe", JWTMiddleware(h.jwt), func(c *gin.Context) {
			_, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			_, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		users.DELETE("/:id/subscribe", JWTMiddleware(h.jwt), func(c *gin.Context) {
			_, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			_, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}

	// Subscriptions and followers routes
	group.GET("/subscriptions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"list":      []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	})

	group.GET("/followers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"list":      []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	})
}
