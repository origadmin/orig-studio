/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * System module - handles stats and settings
 */

package service

import (
	"context"
	"encoding/json"
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemdto "origadmin/application/origstudio/internal/features/system/dto"
	systemData "origadmin/application/origstudio/internal/features/system/dal"
)

type SystemHandler struct {
	jwtMgr    *auth.Manager
	statsRepo *systemData.StatsRepo
	settingUC *systembiz.SettingUseCase
	emailUC   *systembiz.EmailUseCase
}

func NewSystemHandler(
	jwtMgr *auth.Manager,
	statsRepo *systemData.StatsRepo,
	settingUC *systembiz.SettingUseCase,
	emailUC *systembiz.EmailUseCase,
) *SystemHandler {
	return &SystemHandler{
		jwtMgr:    jwtMgr,
		statsRepo: statsRepo,
		settingUC: settingUC,
		emailUC:   emailUC,
	}
}

func (h *SystemHandler) RegisterRoutes(r http2.Router) {
	system := r.Group("/system")
	{
		h.registerSettings(system)
	}

	config := r.Group("/config")
	{
		config.GET("", h.getPublicConfig())
	}
}

func (h *SystemHandler) registerSettings(g http2.Router) {
	settings := g.Group("/settings")
	{
		settings.GET("", h.getSettings())
		settings.PUT("", h.updateSettings())
		settings.GET("/:key", h.getSettingByKey())
		settings.POST("/:key/reset", h.resetSetting())
		settings.GET("/storage/capabilities", h.getStorageCapabilities())
		settings.GET("/email/status", h.getEmailStatus())
		settings.POST("/email/test", h.sendTestEmail())
	}
}

// ==================== Settings Handlers ====================

func (h *SystemHandler) getSettings() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		items, err := h.settingUC.ListAll(ctx.Request().Context())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		grouped := make(map[string][]*systemdto.SettingDTO)
		for _, item := range items {
			masked := h.settingUC.MaskSensitive(item)
			cat := string(item.Category)
			grouped[cat] = append(grouped[cat], masked)
		}

		return server.OKCtx(ctx, grouped)
	}
}

func (h *SystemHandler) updateSettings() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		var req struct {
			Settings []struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value"`
			} `json:"settings" binding:"required"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		var updated []*systemdto.SettingDTO
		for _, item := range req.Settings {
			existing, err := h.settingUC.GetByKey(ctx.Request().Context(), item.Key)
			if err != nil {
				return server.FailCtx(ctx, server.ErrNotFound, "setting not found: "+item.Key)
			}

			if err := ValidateSettingValue(item.Value, existing.Type); err != nil {
				return server.FailCtx(ctx, server.ErrBadRequest, "invalid value for "+item.Key+": "+err.Error())
			}

			if item.Key == "homepage_layout" {
				validLayouts := map[string]bool{"auto": true, "video": true, "article": true, "mixed": true, "welcome": true, "doc": true}
				if !validLayouts[item.Value] {
					return server.FailCtx(ctx, server.ErrBadRequest, "invalid value for homepage_layout: must be one of: auto, video, article, mixed, doc, welcome")
				}
			}

			s := &systemdto.SettingDTO{
				Key:           existing.Key,
				Value:         item.Value,
				Type:          existing.Type,
				Category:      existing.Category,
				Description:   existing.Description,
				IsSensitive:   existing.IsSensitive,
				FallbackValue: existing.FallbackValue,
				IsBuiltin:     existing.IsBuiltin,
			}
			result, err := h.settingUC.Upsert(ctx.Request().Context(), s)
			if err != nil {
				return server.FailCtx(ctx, server.ErrInternal, err.Error())
			}
			updated = append(updated, result)
		}

		return server.OKCtx(ctx, updated)
	}
}

func (h *SystemHandler) getSettingByKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		key := ctx.Var("key")
		if key == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "key is required")
		}

		s, err := h.settingUC.GetByKey(ctx.Request().Context(), key)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "setting not found")
		}

		masked := h.settingUC.MaskSensitive(s)
		return server.OKCtx(ctx, masked)
	}
}

func (h *SystemHandler) resetSetting() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		key := ctx.Var("key")
		if key == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "key is required")
		}

		s, err := h.settingUC.ResetToDefault(ctx.Request().Context(), key)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "setting not found")
		}

		masked := h.settingUC.MaskSensitive(s)
		return server.OKCtx(ctx, masked)
	}
}

// ==================== Public Config Endpoint ====================

func (h *SystemHandler) getPublicConfig() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		publicSettings := h.settingUC.GetPublicSettings(ctx.Request().Context())
		return server.OKCtx(ctx, publicSettings)
	}
}

// ==================== Portal Config Endpoint ====================

type portalModuleConfig struct {
	Articles bool `json:"articles"`
	Videos   bool `json:"videos"`
	Music    bool `json:"music"`
}

type portalSiteConfig struct {
	SiteName          string   `json:"site_name"`
	SiteDescription   string   `json:"site_description"`
	PrimaryURL        string   `json:"primary_url"`
	AllowedURLs       []string `json:"allowed_urls"`
	SiteLogoURL       string   `json:"site_logo_url"`
	AllowRegistration bool     `json:"allow_registration"`
	AllowUpload       bool     `json:"allow_upload"`
}

type portalConfigResponse struct {
	Modules portalModuleConfig `json:"modules"`
	Layout  string             `json:"layout"`
	Site    portalSiteConfig   `json:"site"`
	Share   map[string]bool    `json:"share"`
}

func (h *SystemHandler) getPortalConfig() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		rctx := ctx.Request().Context()

		modules := portalModuleConfig{
			Articles: getBoolWithDefault(h.settingUC, rctx, "module_articles", true),
			Videos:   getBoolWithDefault(h.settingUC, rctx, "module_videos", true),
			Music:    getBoolWithDefault(h.settingUC, rctx, "module_music", false),
		}

		configuredLayout := h.settingUC.Get(rctx, "homepage_layout")
		if configuredLayout == "" {
			configuredLayout = "auto"
		}
		layout := resolveLayout(modules, configuredLayout)

		site := portalSiteConfig{
			SiteName:          h.settingUC.Get(rctx, "site_name"),
			SiteDescription:   h.settingUC.Get(rctx, "site_description"),
			PrimaryURL:        h.settingUC.Get(rctx, "primary_url"),
			SiteLogoURL:       h.settingUC.Get(rctx, "site_logo_url"),
			AllowRegistration: getBoolWithDefault(h.settingUC, rctx, "allow_registration", true),
			AllowUpload:       getBoolWithDefault(h.settingUC, rctx, "allow_upload", true),
		}
		if urls := h.settingUC.Get(rctx, "base_urls"); urls != "" {
			_ = json.Unmarshal([]byte(urls), &site.AllowedURLs)
		}

		return server.OKCtx(ctx, portalConfigResponse{
			Modules: modules,
			Layout:  layout,
			Site:    site,
			Share:   parseSharePlatforms(h.settingUC, rctx),
		})
	}
}

// sharePlatformKeys is the canonical set of link-based share platforms.
var sharePlatformKeys = []string{"twitter", "facebook", "whatsapp", "telegram", "linkedin", "weibo"}

// parseSharePlatforms reads the share_platforms setting (JSON map of
// platform->enabled) and returns a map containing every known platform key,
// defaulting missing/unparseable entries to enabled so the share dialog never
// silently drops a platform.
func parseSharePlatforms(uc *systembiz.SettingUseCase, ctx context.Context) map[string]bool {
	out := make(map[string]bool, len(sharePlatformKeys))
	for _, k := range sharePlatformKeys {
		out[k] = true
	}
	raw := uc.Get(ctx, "share_platforms")
	if raw == "" {
		return out
	}
	var cfg map[string]bool
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return out
	}
	for k, v := range cfg {
		out[k] = v
	}
	return out
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

// ==================== Storage & Email Handlers ====================

func (h *SystemHandler) getStorageCapabilities() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.settingUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "settings service not available")
		}

		rctx := ctx.Request().Context()
		s3Configured := h.settingUC.Get(rctx, "s3_endpoint") != "" &&
			h.settingUC.Get(rctx, "s3_bucket") != "" &&
			h.settingUC.Get(rctx, "s3_access_key") != ""

		currentType := h.settingUC.Get(rctx, "storage_type")
		if currentType == "" {
			currentType = "local"
		}

		return server.OKCtx(ctx, map[string]any{
			"current_type":     currentType,
			"available_types":  []string{"local"},
			"s3_configured":    s3Configured,
			"s3_available":     s3Configured,
			"hybrid_available": s3Configured,
		})
	}
}

func (h *SystemHandler) getEmailStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.emailUC == nil {
			return server.OKCtx(ctx, map[string]any{"configured": false})
		}
		cfg := h.emailUC.GetSMTPConfig(ctx.Request().Context())
		return server.OKCtx(ctx, cfg)
	}
}

func (h *SystemHandler) sendTestEmail() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if h.emailUC == nil {
			return server.FailCtx(ctx, server.ErrInternal, "email service not available")
		}

		var req struct {
			To string `json:"to" binding:"required,email"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if err := h.emailUC.SendTestEmail(ctx.Request().Context(), req.To); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, "Failed to send test email: "+err.Error())
		}
		return server.OKCtx(ctx, map[string]any{"message": "Test email sent"})
	}
}

// ==================== Validation Helpers ====================

func ValidateSettingValue(value string, typ systemdto.SettingType) error {
	switch typ {
	case systemdto.SettingTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return err
		}
	case systemdto.SettingTypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return err
		}
	case systemdto.SettingTypeString:
	case systemdto.SettingTypeJSON:
	}
	return nil
}
