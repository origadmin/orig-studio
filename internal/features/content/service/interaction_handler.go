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
	"net/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/helpers/repo"
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
	r := ginadapter.NewStdRouterAdapter(g)
	likes := r.Group("/likes")
	{
		// Static routes first
		likes.GET("/status", server.WithJWT(h.jwtMgr, h.getLikeStatusBatch()))

		// Collection routes
		likes.GET("", server.WithJWT(h.jwtMgr, h.getLikes()))
		likes.POST("", server.WithJWT(h.jwtMgr, h.toggleLike()))
	}
}

// registerFavorites handles all favorite-related routes
func (h *InteractionHandler) registerFavorites(g *gin.RouterGroup) {
	r := ginadapter.NewStdRouterAdapter(g)
	favorites := r.Group("/favorites")
	{
		// Static routes first
		favorites.GET("/check", server.WithJWT(h.jwtMgr, h.checkFavorite()))

		// Collection routes
		favorites.GET("", server.WithJWT(h.jwtMgr, h.getFavorites()))
		favorites.POST("", server.WithJWT(h.jwtMgr, h.toggleFavorite()))
	}
}

// registerSubscriptions handles all subscription-related routes
func (h *InteractionHandler) registerSubscriptions(g *gin.RouterGroup) {
	r := ginadapter.NewStdRouterAdapter(g)
	subscriptions := r.Group("/subscriptions")
	{
		// Static routes first
		subscriptions.GET("/count", server.WithJWT(h.jwtMgr, h.getSubscriptionCount()))

		// Collection routes
		subscriptions.GET("", server.WithJWT(h.jwtMgr, h.getSubscriptions()))
	}

	followers := r.Group("/followers")
	{
		// Static routes first
		followers.GET("/count", server.WithJWT(h.jwtMgr, h.getFollowerCount()))

		// Collection routes
		followers.GET("", server.WithJWT(h.jwtMgr, h.getFollowers()))
	}
}

// registerShares handles all share-related routes
func (h *InteractionHandler) registerShares(g *gin.RouterGroup) {
	r := ginadapter.NewStdRouterAdapter(g)
	shares := r.Group("/shares")
	{
		// Collection routes only
		shares.POST("", server.WithJWT(h.jwtMgr, h.createShare()))
	}
}

// ==================== Like Handlers ====================

func (h *InteractionHandler) getLikes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	}
}

func (h *InteractionHandler) toggleLike() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement toggle like
		server.OK(gc, gin.H{"success": true})
	}
}

func (h *InteractionHandler) getLikeStatusBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement batch like status
		server.OK(gc, gin.H{"status": map[string]bool{}})
	}
}

// ==================== Favorite Handlers ====================

func (h *InteractionHandler) getFavorites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.OK(gc, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
	}
}

func (h *InteractionHandler) toggleFavorite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement toggle favorite
		server.OK(gc, gin.H{"success": true})
	}
}

func (h *InteractionHandler) checkFavorite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement check favorite
		server.OK(gc, gin.H{"is_favorite": false})
	}
}

// ==================== Subscription Handlers ====================

func (h *InteractionHandler) getSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		val, exists := gc.Get("claims")
		if !exists {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		claims := val.(*auth.Claims)

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		// TODO: Implement get subscriptions from use case
		_ = claims
		_ = page
		_ = pageSize

		server.OK(gc, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *InteractionHandler) getSubscriptionCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement get subscription count
		server.OK(gc, gin.H{"count": 0})
	}
}

func (h *InteractionHandler) getFollowers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		val, exists := gc.Get("claims")
		if !exists {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		claims := val.(*auth.Claims)

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		// Normalize pagination parameters
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

		// TODO: Implement get followers from use case
		_ = claims
		_ = page
		_ = pageSize

		server.OK(gc, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *InteractionHandler) getFollowerCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement get follower count
		server.OK(gc, gin.H{"count": 0})
	}
}

// ==================== Share Handlers ====================

func (h *InteractionHandler) createShare() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		// TODO: Implement create share
		server.OK(gc, gin.H{"success": true})
	}
}
