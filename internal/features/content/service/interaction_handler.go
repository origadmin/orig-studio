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
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/helpers/repo"

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
func (h *InteractionHandler) RegisterRoutes(r http2.Router) {
	interactions := r.Group("/interactions")
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
func (h *InteractionHandler) registerLikes(r http2.Router) {
	likes := r.Group("/likes")
	{
		// Static routes first
		likes.GET("/status", server.WithJWTCtx(h.jwtMgr, h.getLikeStatusBatch()))

		// Collection routes
		likes.GET("", server.WithJWTCtx(h.jwtMgr, h.getLikes()))
		likes.POST("", server.WithJWTCtx(h.jwtMgr, h.toggleLike()))
	}
}

// registerFavorites handles all favorite-related routes
func (h *InteractionHandler) registerFavorites(r http2.Router) {
	favorites := r.Group("/favorites")
	{
		// Static routes first
		favorites.GET("/check", server.WithJWTCtx(h.jwtMgr, h.checkFavorite()))

		// Collection routes
		favorites.GET("", server.WithJWTCtx(h.jwtMgr, h.getFavorites()))
		favorites.POST("", server.WithJWTCtx(h.jwtMgr, h.toggleFavorite()))
	}
}

// registerSubscriptions handles all subscription-related routes
func (h *InteractionHandler) registerSubscriptions(r http2.Router) {
	subscriptions := r.Group("/subscriptions")
	{
		// Static routes first
		subscriptions.GET("/count", server.WithJWTCtx(h.jwtMgr, h.getSubscriptionCount()))

		// Collection routes
		subscriptions.GET("", server.WithJWTCtx(h.jwtMgr, h.getSubscriptions()))
	}

	followers := r.Group("/followers")
	{
		// Static routes first
		followers.GET("/count", server.WithJWTCtx(h.jwtMgr, h.getFollowerCount()))

		// Collection routes
		followers.GET("", server.WithJWTCtx(h.jwtMgr, h.getFollowers()))
	}
}

// registerShares handles all share-related routes
func (h *InteractionHandler) registerShares(r http2.Router) {
	shares := r.Group("/shares")
	{
		// Collection routes only
		shares.POST("", server.WithJWTCtx(h.jwtMgr, h.createShare()))
	}
}

// ==================== Like Handlers ====================

func (h *InteractionHandler) getLikes() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
		return nil
	}
}

func (h *InteractionHandler) toggleLike() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement toggle like
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}

func (h *InteractionHandler) getLikeStatusBatch() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement batch like status
		http2.OK(ctx, gin.H{"status": map[string]bool{}})
		return nil
	}
}

// ==================== Favorite Handlers ====================

func (h *InteractionHandler) getFavorites() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		http2.OK(ctx, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      1,
			"page_size": 20,
		})
		return nil
	}
}

func (h *InteractionHandler) toggleFavorite() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement toggle favorite
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}

func (h *InteractionHandler) checkFavorite() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement check favorite
		http2.OK(ctx, gin.H{"is_favorite": false})
		return nil
	}
}

// ==================== Subscription Handlers ====================

func (h *InteractionHandler) getSubscriptions() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
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

		http2.OK(ctx, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
		return nil
	}
}

func (h *InteractionHandler) getSubscriptionCount() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement get subscription count
		http2.OK(ctx, gin.H{"count": 0})
		return nil
	}
}

func (h *InteractionHandler) getFollowers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
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

		http2.OK(ctx, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
		return nil
	}
}

func (h *InteractionHandler) getFollowerCount() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement get follower count
		http2.OK(ctx, gin.H{"count": 0})
		return nil
	}
}

// ==================== Share Handlers ====================

func (h *InteractionHandler) createShare() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		_ = gc
		// TODO: Implement create share
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}
