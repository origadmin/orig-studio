/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Interaction module - handles likes, favorites, subscriptions, shares
 */

package service

import (
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origstudio/internal/helpers/http"
	ginadapter "origadmin/application/origstudio/internal/helpers/http/gin"
	"origadmin/application/origstudio/internal/helpers/repo"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/server"
)

type InteractionHandler struct {
	jwtMgr         *auth.Manager
	likeFavoriteUC *biz.LikeFavoriteUseCase
}

func NewInteractionHandler(
	jwtMgr *auth.Manager,
	likeFavoriteUC *biz.LikeFavoriteUseCase,
) *InteractionHandler {
	return &InteractionHandler{
		jwtMgr:         jwtMgr,
		likeFavoriteUC: likeFavoriteUC,
	}
}

func (h *InteractionHandler) RegisterRoutes(r http2.Router) {
	interactions := r.Group("/interactions")
	{
		h.registerLikes(interactions)
		h.registerFavorites(interactions)
		h.registerSubscriptions(interactions)
		h.registerShares(interactions)
	}
}

func (h *InteractionHandler) registerLikes(r http2.Router) {
	likes := r.Group("/likes")
	{
		likes.GET("/status", server.WithJWTCtx(h.jwtMgr, h.getLikeStatusBatch()))
		likes.GET("", server.WithJWTCtx(h.jwtMgr, h.getLikes()))
		likes.POST("", server.WithJWTCtx(h.jwtMgr, h.toggleLike()))
	}
}

func (h *InteractionHandler) registerFavorites(r http2.Router) {
	favorites := r.Group("/favorites")
	{
		favorites.GET("/check", server.WithJWTCtx(h.jwtMgr, h.checkFavorite()))
		favorites.GET("", server.WithJWTCtx(h.jwtMgr, h.getFavorites()))
		favorites.POST("", server.WithJWTCtx(h.jwtMgr, h.toggleFavorite()))
	}
}

func (h *InteractionHandler) registerSubscriptions(r http2.Router) {
	subscriptions := r.Group("/subscriptions")
	{
		subscriptions.GET("/count", server.WithJWTCtx(h.jwtMgr, h.getSubscriptionCount()))
		subscriptions.GET("", server.WithJWTCtx(h.jwtMgr, h.getSubscriptions()))
	}

	followers := r.Group("/followers")
	{
		followers.GET("/count", server.WithJWTCtx(h.jwtMgr, h.getFollowerCount()))
		followers.GET("", server.WithJWTCtx(h.jwtMgr, h.getFollowers()))
	}
}

func (h *InteractionHandler) registerShares(r http2.Router) {
	shares := r.Group("/shares")
	{
		shares.POST("", server.WithJWTCtx(h.jwtMgr, h.createShare()))
	}
}

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
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}

func (h *InteractionHandler) getLikeStatusBatch() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{"status": map[string]bool{}})
		return nil
	}
}

func (h *InteractionHandler) getFavorites() http2.HandlerFunc {
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

func (h *InteractionHandler) toggleFavorite() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}

func (h *InteractionHandler) checkFavorite() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{"is_favorite": false})
		return nil
	}
}

func (h *InteractionHandler) getSubscriptions() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists {
			http2.Fail(ctx, http2.ErrUnauthorized, "unauthorized")
			return nil
		}
		claims := val.(*auth.Claims)

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

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
		http2.OK(ctx, gin.H{"count": 0})
		return nil
	}
}

func (h *InteractionHandler) getFollowers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		val, exists := gc.Get("claims")
		if !exists {
			http2.Fail(ctx, http2.ErrUnauthorized, "unauthorized")
			return nil
		}
		claims := val.(*auth.Claims)

		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		page, pageSize = repo.NormalizeHTTPPagination(page, pageSize)

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
		http2.OK(ctx, gin.H{"count": 0})
		return nil
	}
}

func (h *InteractionHandler) createShare() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{"success": true})
		return nil
	}
}
