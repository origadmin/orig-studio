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
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/dto"
)

type UserHandler struct {
	uc  *biz.UserUseCase
	jwt *auth.Manager
}

func NewUserHandler(uc *biz.UserUseCase, jwt *auth.Manager) *UserHandler {
	return &UserHandler{uc: uc, jwt: jwt}
}

func (h *UserHandler) Register(r handler.Router) {
	users := r.Group("/users")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		// Current user endpoints
		users.GET("/me", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			claims := c.Get("claims").(*auth.Claims)
			if claims == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			u, err := h.uc.GetUser(r.Context(), claims.GetUserID())
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, u)
		}))

		users.PUT("/me", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			claims := c.Get("claims").(*auth.Claims)
			if claims == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			var input struct {
				Nickname string `json:"nickname"`
				Email    string `json:"email" binding:"omitempty,email"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			u, err := h.uc.GetUser(r.Context(), claims.GetUserID())
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}

			if input.Nickname != "" {
				u.Nickname = input.Nickname
			}
			if input.Email != "" {
				u.Email = input.Email
			}

			updated, err := h.uc.UpdateUser(r.Context(), u)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, updated)
		}))

		users.PUT("/me/password", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			claims := c.Get("claims").(*auth.Claims)
			if claims == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			var input struct {
				OldPassword string `json:"old_password" binding:"required"`
				NewPassword string `json:"new_password" binding:"required,min=6"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Verify old password
			if err := h.uc.VerifyPassword(r.Context(), claims.GetUserID(), input.OldPassword); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid old password"})
				return
			}

			// Update password
			// TODO: Implement UpdatePassword
			// hashedPassword, err := h.uc.HashPassword(input.NewPassword)
			// if err != nil {
			//  c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			//  return
			// }

			// if err := h.uc.UpdatePassword(r.Context(), claims.GetUserID(), hashedPassword); err != nil {
			//  c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			//  return
			// }

			c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
		}))

		// List users
		users.GET("", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			limit, _ := strconv.Atoi(c.Query("limit"))
			if limit == 0 {
				limit = 20
			}
			page, _ := strconv.Atoi(c.Query("page"))
			if page == 0 {
				page = 1
			}

			items, total, err := h.uc.ListUsers(r.Context(), &dto.UserQueryOption{
				QueryOption: repo.QueryOption{
					Page:     int32(page),
					PageSize: int32(limit),
				},
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"items":     items,
				"total":     total,
				"page":      page,
				"page_size": limit,
			})
		})

		// Create user
		users.POST("", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			var input struct {
				Username string `json:"username" binding:"required"`
				Email    string `json:"email" binding:"required,email"`
				Password string `json:"password" binding:"required,min=6"`
				Name     string `json:"name"`
			}
			if err := c.Bind(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			hashedPassword, _ := h.uc.HashPassword(input.Password)
			u, err := h.uc.CreateUser(r.Context(), &types.User{
				Username: input.Username,
				Email:    input.Email,
				Nickname: input.Name, // Use Name as Nickname if Name is not in User
			}, hashedPassword)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": u})
		})

		// ================================
		// 2. NESTED RESOURCE ROUTES
		// ================================
		// User related resources
		users.GET("/:id/playlists", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			var userID string
			if id == "me" {
				if claims := c.Get("claims"); claims != nil {
					userID = claims.(*auth.Claims).GetUserID()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				userID = id
			}

			// TODO: Implement playlist listing
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "playlists": []interface{}{}})
		})

		// Get user by username
		users.GET("/username/:username", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			username := c.Param("username")
			if username == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username"})
				return
			}

			u, err := h.uc.GetUserByUsername(r.Context(), username)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, u)
		})

		users.GET("/:id/favorites", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			var userID string
			if id == "me" {
				if claims := c.Get("claims"); claims != nil {
					userID = claims.(*auth.Claims).GetUserID()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				userID = id
			}

			// TODO: Implement favorites listing
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "favorites": []interface{}{}})
		}))

		users.GET("/:id/likes", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			var userID string
			if id == "me" {
				if claims := c.Get("claims"); claims != nil {
					userID = claims.(*auth.Claims).GetUserID()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				userID = id
			}

			// TODO: Implement likes listing
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "likes": []interface{}{}})
		}))

		users.GET("/:id/subscriptions", WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			var userID string
			if id == "me" {
				if claims := c.Get("claims"); claims != nil {
					userID = claims.(*auth.Claims).GetUserID()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				userID = id
			}

			page, _ := strconv.Atoi(c.Query("page"))
			if page == 0 {
				page = 1
			}
			pageSize, _ := strconv.Atoi(c.Query("page_size"))
			if pageSize == 0 {
				pageSize = 20
			}

			list, total, err := h.uc.GetSubscriptions(
				r.Context(),
				userID,
				page,
				pageSize,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"items":     list,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
		}))

		users.GET("/:id/followers", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			var userID string
			if id == "me" {
				if claims := c.Get("claims"); claims != nil {
					userID = claims.(*auth.Claims).GetUserID()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				userID = id
			}

			page, _ := strconv.Atoi(c.Query("page"))
			if page == 0 {
				page = 1
			}
			pageSize, _ := strconv.Atoi(c.Query("page_size"))
			if pageSize == 0 {
				pageSize = 20
			}

			list, total, err := h.uc.GetSubscribers(
				r.Context(),
				userID,
				page,
				pageSize,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"items":     list,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
		})

		users.GET("/:id/stats", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			var userID string
			if id == "me" {
				if claims := c.Get("claims"); claims != nil {
					userID = claims.(*auth.Claims).GetUserID()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				userID = id
			}

			// TODO: Implement user stats
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "stats": gin.H{}})
		})

		// User channels
		users.GET("/:id/channels", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			// Handle "me" as special case
			if id == "me" {
				if c.Get("claims") == nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			}
			// No need to parse as int64, use string directly

			limit, _ := strconv.Atoi(c.Query("limit"))
			if limit == 0 {
				limit = 100
			}
			page, _ := strconv.Atoi(c.Query("page"))
			if page == 0 {
				page = 1
			}

			// TODO: Implement ListUserChannels
			c.JSON(http.StatusOK, gin.H{
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
		users.GET("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			u, err := h.uc.GetUser(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, u)
		})

		// Delete user
		users.DELETE("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id := c.Param("id")
			err := h.uc.DeleteUser(r.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}

	// Move subscription routes to channels module
	// These will be handled in the channels module
}
