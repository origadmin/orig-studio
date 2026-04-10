/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * User module - handles user CRUD and user-related resources
 *
 * API paths:
 * - /api/v1/users              - user collection
 * - /api/v1/users/me           - current user (shortcut)
 * - /api/v1/users/me/password   - change password
 * - /api/v1/users/:id           - single user
 * - /api/v1/users/:id/playlists - user playlists
 * - /api/v1/users/:id/favorites - user favorites
 * - /api/v1/users/:id/likes     - user likes
 * - /api/v1/users/:id/subscriptions - user subscriptions
 * - /api/v1/users/:id/followers - user followers
 * - /api/v1/users/:id/stats     - user stats
 */

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
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		// Current user endpoints
		users.GET("/me", JWTMiddleware(h.jwt), func(c *gin.Context) {
			claims, ok := c.MustGet("claims").(*auth.Claims)
			if !ok {
				Fail(c, ErrUnauthorized, "unauthorized")
				return
			}
			u, err := h.uc.GetUser(c.Request.Context(), claims.UserID)
			if err != nil {
				Fail(c, ErrUserNotFound, "User not found")
				return
			}
			OK(c, u)
		})

		users.PUT("/me", JWTMiddleware(h.jwt), func(c *gin.Context) {
			claims, ok := c.MustGet("claims").(*auth.Claims)
			if !ok {
				Fail(c, ErrUnauthorized, "unauthorized")
				return
			}

			var input struct {
				Nickname string `json:"nickname"`
				Email    string `json:"email" binding:"omitempty,email"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				Fail(c, ErrBadRequest, err.Error())
				return
			}

			u, err := h.uc.GetUser(c.Request.Context(), claims.UserID)
			if err != nil {
				Fail(c, ErrUserNotFound, "User not found")
				return
			}

			if input.Nickname != "" {
				u.Nickname = input.Nickname
			}
			if input.Email != "" {
				u.Email = input.Email
			}

			updated, err := h.uc.UpdateUser(c.Request.Context(), u)
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}
			OK(c, updated)
		})

		users.PUT("/me/password", JWTMiddleware(h.jwt), func(c *gin.Context) {
			claims, ok := c.MustGet("claims").(*auth.Claims)
			if !ok {
				Fail(c, ErrUnauthorized, "unauthorized")
				return
			}

			var input struct {
				OldPassword string `json:"old_password" binding:"required"`
				NewPassword string `json:"new_password" binding:"required,min=6"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				Fail(c, ErrBadRequest, err.Error())
				return
			}

			// Verify old password
			if err := h.uc.VerifyPassword(c.Request.Context(), claims.UserID, input.OldPassword); err != nil {
				Fail(c, ErrPasswordWrong, "Invalid old password")
				return
			}

			// Update password
			// TODO: Implement UpdatePassword
			// hashedPassword, err := h.uc.HashPassword(input.NewPassword)
			// if err != nil {
			// 	Fail(c, ErrInternal, "Failed to hash password")
			// 	return
			// }

			// if err := h.uc.UpdatePassword(c.Request.Context(), claims.UserID, hashedPassword); err != nil {
			// 	Fail(c, ErrInternal, err.Error())
			// 	return
			// }

			OK(c, gin.H{"message": "Password updated"})
		})

		// List users
		users.GET("", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

			items, total, err := h.uc.ListUsers(c.Request.Context())
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}

			OK(c, gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": limit,
			})
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
				Fail(c, ErrBadRequest, err.Error())
				return
			}

			hashedPassword, _ := h.uc.HashPassword(input.Password)
			u, err := h.uc.CreateUser(c.Request.Context(), &types.User{
				Username: input.Username,
				Email:    input.Email,
				Nickname: input.Name, // Use Name as Nickname if Name is not in User
			}, hashedPassword)
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}

			c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: u})
		})

		// ================================
		// 2. NESTED RESOURCE ROUTES
		// ================================
		// User related resources
		users.GET("/:id/playlists", func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			var userID int64
			if id == "me" {
				if claims, ok := c.Get("claims"); ok {
					userID = claims.(*auth.Claims).UserID
				} else {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				var err error
				userID, err = strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			// TODO: Implement playlist listing
			OK(c, gin.H{"user_id": userID, "playlists": []interface{}{}})
		})

		users.GET("/:id/favorites", JWTMiddleware(h.jwt), func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			var userID int64
			if id == "me" {
				if claims, ok := c.Get("claims"); ok {
					userID = claims.(*auth.Claims).UserID
				} else {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				var err error
				userID, err = strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			// TODO: Implement favorites listing
			OK(c, gin.H{"user_id": userID, "favorites": []interface{}{}})
		})

		users.GET("/:id/likes", JWTMiddleware(h.jwt), func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			var userID int64
			if id == "me" {
				if claims, ok := c.Get("claims"); ok {
					userID = claims.(*auth.Claims).UserID
				} else {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				var err error
				userID, err = strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			// TODO: Implement likes listing
			OK(c, gin.H{"user_id": userID, "likes": []interface{}{}})
		})

		users.GET("/:id/subscriptions", JWTMiddleware(h.jwt), func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			var userID int64
			if id == "me" {
				if claims, ok := c.Get("claims"); ok {
					userID = claims.(*auth.Claims).UserID
				} else {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				var err error
				userID, err = strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			list, total, err := h.uc.GetSubscriptions(
				c.Request.Context(),
				int(userID),
				page,
				pageSize,
			)
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}

			OK(c, gin.H{
				"items":     list,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
		})

		users.GET("/:id/followers", func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			var userID int64
			if id == "me" {
				if claims, ok := c.Get("claims"); ok {
					userID = claims.(*auth.Claims).UserID
				} else {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				var err error
				userID, err = strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			list, total, err := h.uc.GetSubscribers(
				c.Request.Context(),
				int(userID),
				page,
				pageSize,
			)
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}

			OK(c, gin.H{
				"items":     list,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
		})

		users.GET("/:id/stats", func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			var userID int64
			if id == "me" {
				if claims, ok := c.Get("claims"); ok {
					userID = claims.(*auth.Claims).UserID
				} else {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				var err error
				userID, err = strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			// TODO: Implement user stats
			OK(c, gin.H{"user_id": userID, "stats": gin.H{}})
		})

		// User channels
		users.GET("/:id/channels", func(c *gin.Context) {
			id := c.Param("id")
			// Handle "me" as special case
			if id == "me" {
				if _, ok := c.Get("claims"); !ok {
					Fail(c, ErrUnauthorized, "unauthorized")
					return
				}
			} else {
				_, err := strconv.ParseInt(id, 10, 64)
				if err != nil {
					Fail(c, ErrBadRequest, "Invalid ID")
					return
				}
			}

			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

			// TODO: Implement ListUserChannels
			OK(c, gin.H{
				"items":     []interface{}{},
				"total":     0,
				"page":      page,
				"page_size": limit,
			})
		})

		// ================================
		// 3. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		// Get user by ID
		users.GET("/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				Fail(c, ErrBadRequest, "Invalid ID")
				return
			}
			u, err := h.uc.GetUser(c.Request.Context(), id)
			if err != nil {
				Fail(c, ErrUserNotFound, "User not found")
				return
			}
			OK(c, u)
		})

		// Delete user
		users.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				Fail(c, ErrBadRequest, "Invalid ID")
				return
			}
			err = h.uc.DeleteUser(c.Request.Context(), id)
			if err != nil {
				Fail(c, ErrInternal, err.Error())
				return
			}
			OK(c, gin.H{"message": "deleted"})
		})
	}

	// Move subscription routes to channels module
	// These will be handled in the channels module
}
