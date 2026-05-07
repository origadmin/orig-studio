/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * System module - handles stats and settings
 */

package service

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/setting"
	"origadmin/application/origcms/internal/server"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemData "origadmin/application/origcms/internal/features/system/dal"
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

func (h *SystemHandler) RegisterRoutes(r http2.Router) {
	system := r.Group("/system")
	{
		h.registerStats(system)
		h.registerSettings(system)
	}

	config := r.Group("/config")
	{
		config.GET("", server.HTTPToHandlerFunc(h.getPublicConfig()))
	}

	portal := r.Group("/portal")
	{
		portal.GET("/config", server.HTTPToHandlerFunc(h.getPortalConfig()))
	}
}

func (h *SystemHandler) registerStats(g http2.Router) {
	stats := g.Group("/stats")
	{
		stats.GET("/dashboard", server.HTTPToHandlerFunc(h.getDashboardStats()))
		stats.GET("/media", server.HTTPToHandlerFunc(h.getMediaStats()))
		stats.GET("/traffic", server.HTTPToHandlerFunc(h.getTrafficStats()))
		stats.GET("/users", server.HTTPToHandlerFunc(h.getUserStats()))
	}
}

func (h *SystemHandler) registerSettings(g http2.Router) {
	settings := g.Group("/settings")
	{
		settings.GET("", server.HTTPToHandlerFunc(h.getSettings()))
		settings.PUT("", server.HTTPToHandlerFunc(h.updateSettings()))
		settings.GET("/:key", server.HTTPToHandlerFunc(h.getSettingByKey()))
		settings.POST("/:key/reset", server.HTTPToHandlerFunc(h.resetSetting()))
	}
}

// ==================== Stats Handlers ====================

func (h *SystemHandler) getDashboardStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.statsRepo == nil {
			server.Fail(gc, server.ErrInternal, "stats service not available")
			return
		}

		stats, err := h.statsRepo.GetDashboardStats(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, stats)
	}
}

func (h *SystemHandler) getMediaStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.statsRepo == nil {
			server.Fail(gc, server.ErrInternal, "stats service not available")
			return
		}

		stats, err := h.statsRepo.GetMediaStats(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, stats)
	}
}

func (h *SystemHandler) getUserStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.statsRepo == nil {
			server.Fail(gc, server.ErrInternal, "stats service not available")
			return
		}

		stats, err := h.statsRepo.GetUserStats(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, stats)
	}
}

func (h *SystemHandler) getTrafficStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))

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

// ==================== Settings Handlers ====================

func (h *SystemHandler) getSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		items, err := h.settingUC.ListAll(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		grouped := make(map[string][]*entity.Setting)
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			cat := string(item.Category)
			grouped[cat] = append(grouped[cat], masked)
		}

		server.OK(gc, grouped)
	}
}

func (h *SystemHandler) updateSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		var req struct {
			Settings []struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value"`
			} `json:"settings" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		var updated []*entity.Setting
		for _, item := range req.Settings {
			existing, err := h.settingUC.GetByKey(r.Context(), item.Key)
			if err != nil {
				server.Fail(gc, server.ErrNotFound, "setting not found: "+item.Key)
				return
			}

			if err := ValidateSettingValue(item.Value, existing.Type); err != nil {
				server.Fail(gc, server.ErrBadRequest, "invalid value for "+item.Key+": "+err.Error())
				return
			}

			if item.Key == "homepage_layout" {
				validLayouts := map[string]bool{"auto": true, "video": true, "article": true, "mixed": true, "welcome": true, "doc": true}
				if !validLayouts[item.Value] {
					server.Fail(gc, server.ErrBadRequest, "invalid value for homepage_layout: must be one of: auto, video, article, mixed, doc, welcome")
					return
				}
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
			result, err := h.settingUC.Upsert(r.Context(), s)
			if err != nil {
				server.Fail(gc, server.ErrInternal, err.Error())
				return
			}
			updated = append(updated, result)
		}

		server.OK(gc, updated)
	}
}

func (h *SystemHandler) getSettingByKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		key := gc.Param("key")
		if key == "" {
			server.Fail(gc, server.ErrBadRequest, "key is required")
			return
		}

		s, err := h.settingUC.GetByKey(r.Context(), key)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "setting not found")
			return
		}

		masked := h.settingUC.MaskSensitive(s)
		server.OK(gc, masked)
	}
}

func (h *SystemHandler) resetSetting() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		key := gc.Param("key")
		if key == "" {
			server.Fail(gc, server.ErrBadRequest, "key is required")
			return
		}

		s, err := h.settingUC.ResetToDefault(r.Context(), key)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "setting not found")
			return
		}

		masked := h.settingUC.MaskSensitive(s)
		server.OK(gc, masked)
	}
}

// ==================== Public Config Endpoint ====================

func (h *SystemHandler) getPublicConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		publicSettings := h.settingUC.GetPublicSettings(r.Context())
		server.OK(gc, publicSettings)
	}
}

// ==================== Portal Config Endpoint ====================

type portalModuleConfig struct {
	Articles bool `json:"articles"`
	Videos   bool `json:"videos"`
	Music    bool `json:"music"`
}

type portalSiteConfig struct {
	SiteName          string `json:"site_name"`
	SiteDescription   string `json:"site_description"`
	AllowRegistration bool   `json:"allow_registration"`
	AllowUpload       bool   `json:"allow_upload"`
}

type portalConfigResponse struct {
	Modules portalModuleConfig `json:"modules"`
	Layout  string             `json:"layout"`
	Site    portalSiteConfig   `json:"site"`
}

func (h *SystemHandler) getPortalConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if h.settingUC == nil {
			server.Fail(gc, server.ErrInternal, "settings service not available")
			return
		}

		ctx := r.Context()

		modules := portalModuleConfig{
			Articles: getBoolWithDefault(h.settingUC, ctx, "module_articles", true),
			Videos:   getBoolWithDefault(h.settingUC, ctx, "module_videos", true),
			Music:    getBoolWithDefault(h.settingUC, ctx, "module_music", false),
		}

		configuredLayout := h.settingUC.Get(ctx, "homepage_layout")
		if configuredLayout == "" {
			configuredLayout = "auto"
		}
		layout := resolveLayout(modules, configuredLayout)

		site := portalSiteConfig{
			SiteName:          h.settingUC.Get(ctx, "site_name"),
			SiteDescription:   h.settingUC.Get(ctx, "site_description"),
			AllowRegistration: getBoolWithDefault(h.settingUC, ctx, "allow_registration", true),
			AllowUpload:       getBoolWithDefault(h.settingUC, ctx, "allow_upload", true),
		}

		server.OK(gc, portalConfigResponse{
			Modules: modules,
			Layout:  layout,
			Site:    site,
		})
	}
}

func getBoolWithDefault(uc *systembiz.SettingUseCase, ctx context.Context, key string, defaultValue bool) bool {
	val := uc.Get(ctx, key)
	if val == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return b
}

func resolveLayout(modules portalModuleConfig, configuredLayout string) string {
	if configuredLayout != "auto" {
		if configuredLayout == "video" && !modules.Videos {
			return "welcome"
		}
		if configuredLayout == "article" && !modules.Articles {
			return "welcome"
		}
		if configuredLayout == "mixed" && (!modules.Videos || !modules.Articles) {
			return "welcome"
		}
		if configuredLayout == "doc" && !modules.Articles {
			return "welcome"
		}
		return configuredLayout
	}

	activeCount := 0
	if modules.Articles {
		activeCount++
	}
	if modules.Videos {
		activeCount++
	}
	if modules.Music {
		activeCount++
	}

	switch {
	case activeCount == 0:
		return "welcome"
	case modules.Videos && !modules.Articles:
		return "video"
	case modules.Articles && !modules.Videos:
		return "doc"
	default:
		return "mixed"
	}
}

// ==================== Validation Helpers ====================

func ValidateSettingValue(value string, typ setting.Type) error {
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
