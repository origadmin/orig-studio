package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-content/biz"
)

// LikeHandler handles /api/v1/likes routes.
type LikeHandler struct {
	uc  *biz.LikeFavoriteUseCase
	jwt *auth.Manager
}

// NewLikeHandler creates a new LikeHandler.
func NewLikeHandler(uc *biz.LikeFavoriteUseCase, jwt *auth.Manager) *LikeHandler {
	return &LikeHandler{uc: uc, jwt: jwt}
}

func (h *LikeHandler) Register(group *gin.RouterGroup) {
	// Original routes for backward compatibility
	likes := group.Group("/likes")
	{
		// Protected routes — all like operations require auth
		likes.Use(JWTMiddleware(h.jwt))

		likes.POST("", h.toggleLike)
		likes.GET("/media/:mediaId", h.getMediaLikes)

		// Check if current user has liked a specific media
		likes.GET("/check/:mediaId", h.checkLike)
	}
}

// toggleLike toggles like/dislike status for a media item.
// POST body: {"media_id": int, "type": "like"|"dislike"} → {"liked": bool, "disliked": bool}
func (h *LikeHandler) toggleLike(c *gin.Context) {
	val, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	claims := val.(*auth.Claims)

	var input struct {
		MediaID int    `json:"media_id" binding:"required"`
		Type    string `json:"type" binding:"required,oneof=like dislike"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.uc.ToggleLike(c.Request.Context(), int(claims.UserID), input.MediaID, input.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated stats for response
	stats, _ := h.uc.GetMediaStats(c.Request.Context(), int(claims.UserID), input.MediaID)
	c.JSON(http.StatusOK, gin.H{
		"liked":    stats.UserLikeType == "like",
		"disliked": stats.UserLikeType == "dislike",
	})
}

// getMediaLikes returns like/dislike counts and current user's status for a media.
// GET /likes/media/:mediaId → {"likes": N, "dislikes": N, "user_liked": "none"|"liked"|"disliked"}
func (h *LikeHandler) getMediaLikes(c *gin.Context) {
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	var userID int
	val, ok := c.Get("claims")
	if ok && val != nil {
		claims := val.(*auth.Claims)
		userID = int(claims.UserID)
	}

	stats, err := h.uc.GetMediaStats(c.Request.Context(), userID, mediaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"likes":      stats.LikeCount,
		"dislikes":   stats.DislikeCount,
		"user_liked": stats.UserLikeType,
	})
}

// checkLike checks if the current user has liked a specific media.
// GET /likes/check/:mediaId → {"liked": bool, "type": "none"|"like"|"disliked"}
func (h *LikeHandler) checkLike(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"liked": stats.UserLikeType != "none",
		"type":  stats.UserLikeType,
	})
}

// getLikeStatus returns like count and current user's status for a media.
// GET /media/:mediaId/like → {"is_liked": bool, "is_disliked": bool, "like_count": int, "dislike_count": int}
func (h *LikeHandler) getLikeStatus(c *gin.Context) {
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}

	var userID int
	val, ok := c.Get("claims")
	if ok && val != nil {
		claims := val.(*auth.Claims)
		userID = int(claims.UserID)
	}

	stats, err := h.uc.GetMediaStats(c.Request.Context(), userID, mediaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_liked":      stats.UserLikeType == "like",
		"is_disliked":   stats.UserLikeType == "dislike",
		"like_count":    stats.LikeCount,
		"dislike_count": stats.DislikeCount,
	})
}

// toggleLikeStatus toggles like status for a media item.
// POST /media/:mediaId/like → {"is_liked": bool, "is_disliked": bool, "like_count": int, "dislike_count": int}
func (h *LikeHandler) toggleLikeStatus(c *gin.Context) {
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

	_, err = h.uc.ToggleLike(c.Request.Context(), int(claims.UserID), mediaId, "like")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated stats for response
	stats, _ := h.uc.GetMediaStats(c.Request.Context(), int(claims.UserID), mediaId)
	c.JSON(http.StatusOK, gin.H{
		"is_liked":      stats.UserLikeType == "like",
		"is_disliked":   stats.UserLikeType == "dislike",
		"like_count":    stats.LikeCount,
		"dislike_count": stats.DislikeCount,
	})
}

// toggleDislikeStatus toggles dislike status for a media item.
// POST /media/:mediaId/dislike → {"is_liked": bool, "is_disliked": bool, "like_count": int, "dislike_count": int}
func (h *LikeHandler) toggleDislikeStatus(c *gin.Context) {
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

	_, err = h.uc.ToggleLike(c.Request.Context(), int(claims.UserID), mediaId, "dislike")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated stats for response
	stats, _ := h.uc.GetMediaStats(c.Request.Context(), int(claims.UserID), mediaId)
	c.JSON(http.StatusOK, gin.H{
		"is_liked":      stats.UserLikeType == "like",
		"is_disliked":   stats.UserLikeType == "dislike",
		"like_count":    stats.LikeCount,
		"dislike_count": stats.DislikeCount,
	})
}
