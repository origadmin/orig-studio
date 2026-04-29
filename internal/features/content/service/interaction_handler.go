/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Interaction module - handles likes, favorites, subscriptions, shares
 *
 * DEPRECATED: This module is deprecated. Interaction functionality has been
 * moved to their respective resource modules:
 * - Likes/Favorites/Shares → /medias/:id/*
 * - Subscriptions → /channels/:id/*
 * - User interactions → /users/:id/*
 */

package service

import (
	"origadmin/application/origcms/internal/server"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/features/content/biz"
)

// InteractionHandler handles all user interaction routes
// DEPRECATED: Use resource-specific handlers instead
type InteractionHandler struct {
	jwtMgr         *auth.Manager
	likeFavoriteUC *biz.LikeFavoriteUseCase
}

// NewInteractionHandler creates a new InteractionHandler
// DEPRECATED: Use resource-specific handlers instead
func NewInteractionHandler(
	jwtMgr *auth.Manager,
	likeFavoriteUC *biz.LikeFavoriteUseCase,
) *InteractionHandler {
	return &InteractionHandler{
		jwtMgr:         jwtMgr,
		likeFavoriteUC: likeFavoriteUC,
	}
}

// RegisterRoutes registers all interaction routes
// DEPRECATED: Use resource-specific routes instead
func (h *InteractionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	interactions := rg.Group("/interactions")
	{
		// ========== 1. Static sub-routes (alphabetical order) ==========
		// No static sub-routes at root level

		// ========== 2. Collection operations ==========
		// No collection operations at root level

		// ========== 3. Likes sub-module ==========
		h.registerLikes(interactions)

		// ========== 4. Favorites sub-module ==========
		h.registerFavorites(interactions)

		// ========== 5. Subscriptions sub-module ==========
		h.registerSubscriptions(interactions)

		// ========== 6. Shares sub-module ==========
		h.registerShares(interactions)
	}
}

// registerLikes handles all like-related routes
func (h *InteractionHandler) registerLikes(g *gin.RouterGroup) {
	likes := g.Group("/likes")
	{
		// Static routes first
		likes.GET("/status", server.JWTMiddleware(h.jwtMgr), h.getLikeStatusBatch())

		// Collection routes
		likes.GET("", server.JWTMiddleware(h.jwtMgr), h.getLikes())
		likes.POST("", server.JWTMiddleware(h.jwtMgr), h.toggleLike())
	}
}

// registerFavorites handles all favorite-related routes
func (h *InteractionHandler) registerFavorites(g *gin.RouterGroup) {
	favorites := g.Group("/favorites")
	{
		// Static routes first
		favorites.GET("/check", server.JWTMiddleware(h.jwtMgr), h.checkFavorite())

		// Collection routes
		favorites.GET("", server.JWTMiddleware(h.jwtMgr), h.getFavorites())
		favorites.POST("", server.JWTMiddleware(h.jwtMgr), h.toggleFavorite())
	}
}

// registerSubscriptions handles all subscription-related routes
func (h *InteractionHandler) registerSubscriptions(g *gin.RouterGroup) {
	subscriptions := g.Group("/subscriptions")
	{
		// Static routes first
		subscriptions.GET("/count", server.JWTMiddleware(h.jwtMgr), h.getSubscriptionCount())

		// Collection routes
		subscriptions.GET("", server.JWTMiddleware(h.jwtMgr), h.getSubscriptions())
	}

	followers := g.Group("/followers")
	{
		// Static routes first
		followers.GET("/count", server.JWTMiddleware(h.jwtMgr), h.getFollowerCount())

		// Collection routes
		followers.GET("", server.JWTMiddleware(h.jwtMgr), h.getFollowers())
	}
}

// registerShares handles all share-related routes
func (h *InteractionHandler) registerShares(g *gin.RouterGroup) {
	shares := g.Group("/shares")
	{
		// Collection routes only
		shares.POST("", server.JWTMiddleware(h.jwtMgr), h.createShare())
	}
}

// ==================== Like Handlers ====================

func (h *InteractionHandler) getLikes() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	}
}

func (h *InteractionHandler) toggleLike() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement toggle like
		server.OK(c, gin.H{"success": true})
	}
}

func (h *InteractionHandler) getLikeStatusBatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement batch like status
		server.OK(c, gin.H{"status": map[string]bool{}})
	}
}

// ==================== Favorite Handlers ====================

func (h *InteractionHandler) getFavorites() gin.HandlerFunc {
	return func(c *gin.Context) {
		server.OK(c, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	}
}

func (h *InteractionHandler) toggleFavorite() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement toggle favorite
		server.OK(c, gin.H{"success": true})
	}
}

func (h *InteractionHandler) checkFavorite() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement check favorite
		server.OK(c, gin.H{"is_favorite": false})
	}
}

// ==================== Subscription Handlers ====================

func (h *InteractionHandler) getSubscriptions() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("claims")
		if !exists {
			server.Fail(c, server.ErrUnauthorized, "unauthorized")
			return
		}
		claims := val.(*auth.Claims)

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		// TODO: Implement get subscriptions from use case
		_ = claims
		_ = page
		_ = pageSize

		server.OK(c, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *InteractionHandler) getSubscriptionCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get subscription count
		server.OK(c, gin.H{"count": 0})
	}
}

func (h *InteractionHandler) getFollowers() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("claims")
		if !exists {
			server.Fail(c, server.ErrUnauthorized, "unauthorized")
			return
		}
		claims := val.(*auth.Claims)

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		// TODO: Implement get followers from use case
		_ = claims
		_ = page
		_ = pageSize

		server.OK(c, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *InteractionHandler) getFollowerCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get follower count
		server.OK(c, gin.H{"count": 0})
	}
}

// ==================== Share Handlers ====================

func (h *InteractionHandler) createShare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement create share
		server.OK(c, gin.H{"success": true})
	}
}
