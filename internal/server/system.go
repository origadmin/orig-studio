/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * System module - handles stats and settings
 */

package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/setting"
	"origadmin/application/origcms/internal/handler"
	systembiz "origadmin/application/origcms/internal/svc-system/biz"
	systemData "origadmin/application/origcms/internal/svc-system/data"
)

type SystemHandler struct {
	jwtMgr    *auth.Manager
	statsRepo *systemData.StatsRepo
	settingUC *systembiz.SettingUseCase
}

func NewSystemHandler(
	jwtMgr *auth.Manager,
	statsRepo *systemData.StatsRepo,
	settingUC *systembiz.SettingUseCase,
) *SystemHandler {
	return &SystemHandler{
		jwtMgr:    jwtMgr,
		statsRepo: statsRepo,
		settingUC: settingUC,
	}
}

func (h *SystemHandler) Register(r handler.Router) {
	system := r.Group("/system")
	{
		h.registerStats(system)
		h.registerSettings(system)
	}

	config := r.Group("/config")
	{
		config.GET("", GinHandlerToHTTP(h.getPublicConfig()))
	}
}

func (h *SystemHandler) registerStats(g handler.Router) {
	stats := g.Group("/stats")
	{
		stats.GET("/dashboard", GinHandlerToHTTP(h.getDashboardStats()))
		stats.GET("/media", GinHandlerToHTTP(h.getMediaStats()))
		stats.GET("/traffic", GinHandlerToHTTP(h.getTrafficStats()))
		stats.GET("/users", GinHandlerToHTTP(h.getUserStats()))
	}
}

func (h *SystemHandler) registerSettings(g handler.Router) {
	settings := g.Group("/settings")
	{
		settings.GET("", GinHandlerToHTTP(h.getSettings()))
		settings.PUT("", GinHandlerToHTTP(h.updateSettings()))
		settings.GET("/:key", GinHandlerToHTTP(h.getSettingByKey()))
		settings.POST("/:key/reset", GinHandlerToHTTP(h.resetSetting()))
	}
}

// ==================== Stats Handlers ====================

func (h *SystemHandler) getDashboardStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.statsRepo == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stats service not available"})
			return
		}

		stats, err := h.statsRepo.GetDashboardStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (h *SystemHandler) getMediaStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.statsRepo == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stats service not available"})
			return
		}

		stats, err := h.statsRepo.GetMediaStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (h *SystemHandler) getUserStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.statsRepo == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stats service not available"})
			return
		}

		stats, err := h.statsRepo.GetUserStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (h *SystemHandler) getTrafficStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		_ = page
		_ = pageSize

		OK(c, gin.H{
			"items":     []interface{}{},
			"total":     0,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// ==================== Settings Handlers ====================

func (h *SystemHandler) getSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		items, err := h.settingUC.ListAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		grouped := make(map[string][]*entity.Setting)
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			cat := string(item.Category)
			grouped[cat] = append(grouped[cat], masked)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    grouped,
		})
	}
}

func (h *SystemHandler) updateSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		var req struct {
			Settings []struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value"`
			} `json:"settings" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrBadRequest, "message": err.Error()})
			return
		}

		var updated []*entity.Setting
		for _, item := range req.Settings {
			existing, err := h.settingUC.GetByKey(c.Request.Context(), item.Key)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"code": ErrNotFound, "message": "setting not found: " + item.Key},
				)
				return
			}

			if err := validateSettingValue(item.Value, existing.Type); err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{
						"code":    ErrBadRequest,
						"message": "invalid value for " + item.Key + ": " + err.Error(),
					},
				)
				return
			}

			s := &entity.Setting{
				Key:           existing.Key,
				Value:         item.Value,
				Type:          existing.Type,
				Category:      existing.Category,
				Description:   existing.Description,
				IsSensitive:   existing.IsSensitive,
				FallbackValue: existing.FallbackValue,
				IsBuiltin:     existing.IsBuiltin,
			}
			result, err := h.settingUC.Upsert(c.Request.Context(), s)
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{"code": ErrInternal, "message": err.Error()},
				)
				return
			}
			updated = append(updated, result)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    updated,
		})
	}
}

func (h *SystemHandler) getSettingByKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		key := c.Param("key")
		if key == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "key is required"},
			)
			return
		}

		s, err := h.settingUC.GetByKey(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "setting not found"})
			return
		}

		masked := h.settingUC.MaskSensitive(s)
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    masked,
		})
	}
}

func (h *SystemHandler) resetSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		key := c.Param("key")
		if key == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"code": ErrBadRequest, "message": "key is required"},
			)
			return
		}

		s, err := h.settingUC.ResetToDefault(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": ErrNotFound, "message": "setting not found"})
			return
		}

		masked := h.settingUC.MaskSensitive(s)
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    masked,
		})
	}
}

// ==================== Public Config Endpoint ====================

func (h *SystemHandler) getPublicConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.settingUC == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings service not available"})
			return
		}

		publicSettings := h.settingUC.GetPublicSettings(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    publicSettings,
		})
	}
}

// ==================== Validation Helpers ====================

func validateSettingValue(value string, typ setting.Type) error {
	switch typ {
	case setting.TypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return err
		}
	case setting.TypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return err
		}
	case setting.TypeString:
	case setting.TypeJSON:
	}
	return nil
}
