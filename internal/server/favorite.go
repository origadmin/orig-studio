package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
)

// FavoriteHandler handles /api/v1/favorites routes.
type FavoriteHandler struct {
	uc  *biz.LikeFavoriteUseCase
	jwt *auth.Manager
}

// NewFavoriteHandler creates a new FavoriteHandler.
func NewFavoriteHandler(uc *biz.LikeFavoriteUseCase, jwt *auth.Manager) *FavoriteHandler {
	return &FavoriteHandler{uc: uc, jwt: jwt}
}

func (h *FavoriteHandler) Register(group *gin.RouterGroup) {
	// Original routes for backward compatibility
	favGroup := group.Group("/favorites")
	{
		// Protected routes — all favorite operations require auth
		favGroup.Use(JWTMiddleware(h.jwt))

		favGroup.GET("", h.listFavorites)
		favGroup.POST("", h.toggleFavorite)
		favGroup.DELETE("/:id", h.removeFavorite)

		// Check if current user has favorited a specific media
		favGroup.GET("/check/:mediaId", h.checkFavorite)
	}
}

// listFavorites returns favorites for the authenticated user.
func (h *FavoriteHandler) listFavorites(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	items, err := h.uc.ListUserFavorites(c.Request.Context(), int(claims.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": items})
}

// toggleFavorite toggles favorite status for a media item.
// POST body: {"media_id": int} → {"favorited": bool, "is_favorited": bool}
func (h *FavoriteHandler) toggleFavorite(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		MediaID int `json:"media_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.uc.ToggleFavorite(c.Request.Context(), int(claims.UserID), input.MediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favorited":    stats.IsFavorited,
		"is_favorited": stats.IsFavorited,
	})
}

// removeFavorite is deprecated in favor of toggleFavorite, but keeping for compatibility
func (h *FavoriteHandler) removeFavorite(c *gin.Context) {
	_, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// NOTE: biz.Favorite doesn't easily support deleting by record ID if we want to update media counts.
	// For now, we'll just return a placeholder or implement it properly.
	c.JSON(http.StatusNotImplemented, gin.H{"error": "use toggleFavorite instead"})
}

// checkFavorite checks if the current user has favorited a specific media.
// GET /favorites/check/:mediaId → {"favorited": bool}
func (h *FavoriteHandler) checkFavorite(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	stats, err := h.uc.GetMediaStats(c.Request.Context(), int(claims.UserID), mediaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"favorited": stats.IsFavorited})
}

// getFavoriteStatus returns favorite count and current user's status for a media.
// GET /media/:mediaId/favorite → {"is_favorited": bool}
func (h *FavoriteHandler) getFavoriteStatus(c *gin.Context) {
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	var isFavorited bool
	val, ok := c.Get("claims")
	if ok && val != nil {
		claims := val.(*auth.Claims)
		stats, err := h.uc.GetMediaStats(c.Request.Context(), int(claims.UserID), mediaId)
		if err == nil {
			isFavorited = stats.IsFavorited
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"is_favorited": isFavorited,
	})
}

// toggleFavoriteStatus toggles favorite status for a media item.
// POST /media/:mediaId/favorite → {"is_favorited": bool}
func (h *FavoriteHandler) toggleFavoriteStatus(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	favorited, err := h.uc.ToggleFavorite(c.Request.Context(), int(claims.UserID), mediaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_favorited": favorited,
	})
}
