package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/content/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type PortalHandler struct {
	uc         *biz.PortalUseCase
	adUC       *biz.AdUseCase
	catUC      *biz.CategoryTagUseCase
	jwt        *auth.Manager
	settingUC  *systembiz.SettingUseCase
	featureUC  *systembiz.FeatureFlagUseCase
}

func NewPortalHandler(uc *biz.PortalUseCase, adUC *biz.AdUseCase, catUC *biz.CategoryTagUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase, featureUC *systembiz.FeatureFlagUseCase) *PortalHandler {
	return &PortalHandler{uc: uc, adUC: adUC, catUC: catUC, jwt: jwt, settingUC: settingUC, featureUC: featureUC}
}

func (h *PortalHandler) RegisterRoutes(r http2.Router) {
	adminNavItems := r.Group("/admin/nav-items")
	adminNavItems.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminNavItems.GET("", h.listNavItems())
		adminNavItems.POST("", h.createNavItem())
		adminNavItems.PUT("/:id", h.updateNavItem())
		adminNavItems.DELETE("/:id", h.deleteNavItem())
		adminNavItems.PUT("/reorder", h.reorderNavItems())
	}

	adminBanners := r.Group("/admin/banners")
	adminBanners.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminBanners.GET("", h.listBanners())
		adminBanners.POST("", h.createBanner())
		adminBanners.GET("/:id", h.getBanner())
		adminBanners.PUT("/:id", h.updateBanner())
		adminBanners.DELETE("/:id", h.deleteBanner())
		adminBanners.POST("/:id/toggle", h.toggleBanner())
	}

	adminPages := r.Group("/admin/pages")
	adminPages.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminPages.GET("", h.listCustomPages())
		adminPages.POST("", h.createCustomPage())
		adminPages.GET("/:id", h.getCustomPage())
		adminPages.PUT("/:id", h.updateCustomPage())
		adminPages.DELETE("/:id", h.deleteCustomPage())
	}

	pages := r.Group("/p")
	{
		pages.GET("/:slug", h.getPublicPageBySlug())
	}

	portal := r.Group("/portal")
	{
		portal.GET("/config", h.getPortalConfig())
	}
}

// ==================== Admin NavItem Handlers ====================

func (h *PortalHandler) listNavItems() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListNavItems(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, items)
	}
}

func (h *PortalHandler) createNavItem() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Type       string            `json:"type" binding:"required"`
			Label      string            `json:"label" binding:"required"`
			LabelI18n  map[string]string `json:"label_i18n"`
			URL        string            `json:"url"`
			TargetType string            `json:"target_type"`
			TargetID   string            `json:"target_id"`
			Icon       string            `json:"icon"`
			Color      string            `json:"color"`
			Sequence   int               `json:"sequence"`
			ParentID   string            `json:"parent_id"`
			IsVisible  *bool             `json:"is_visible"`
			OpenNewTab *bool             `json:"open_new_tab"`
			CSSClass   string            `json:"css_class"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		isVisible := true
		if input.IsVisible != nil {
			isVisible = *input.IsVisible
		}
		openNewTab := false
		if input.OpenNewTab != nil {
			openNewTab = *input.OpenNewTab
		}

		item := &dto.PortalNavItemDTO{
			Type:       input.Type,
			Label:      input.Label,
			LabelI18n:  input.LabelI18n,
			URL:        input.URL,
			TargetType: input.TargetType,
			TargetID:   input.TargetID,
			Icon:       input.Icon,
			Color:      input.Color,
			Sequence:   input.Sequence,
			ParentID:   input.ParentID,
			IsVisible:  isVisible,
			OpenNewTab: openNewTab,
			CSSClass:   input.CSSClass,
		}

		created, err := h.uc.CreateNavItem(ctx, item)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, created)
	}
}

func (h *PortalHandler) updateNavItem() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var input struct {
			Type       string            `json:"type"`
			Label      string            `json:"label"`
			LabelI18n  map[string]string `json:"label_i18n"`
			URL        string            `json:"url"`
			TargetType string            `json:"target_type"`
			TargetID   string            `json:"target_id"`
			Icon       string            `json:"icon"`
			Color      string            `json:"color"`
			Sequence   *int              `json:"sequence"`
			ParentID   string            `json:"parent_id"`
			IsVisible  *bool             `json:"is_visible"`
			OpenNewTab *bool             `json:"open_new_tab"`
			CSSClass   string            `json:"css_class"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existing, err := h.uc.GetNavItemByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "nav item not found")
		}
		_ = existing

		item := &dto.PortalNavItemDTO{
			ID: id,
		}
		if input.Type != "" {
			item.Type = input.Type
		}
		if input.Label != "" {
			item.Label = input.Label
		}
		if input.LabelI18n != nil {
			item.LabelI18n = input.LabelI18n
		}
		item.URL = input.URL
		item.TargetType = input.TargetType
		item.TargetID = input.TargetID
		item.Icon = input.Icon
		item.Color = input.Color
		if input.Sequence != nil {
			item.Sequence = *input.Sequence
		}
		item.ParentID = input.ParentID
		if input.IsVisible != nil {
			item.IsVisible = *input.IsVisible
		}
		if input.OpenNewTab != nil {
			item.OpenNewTab = *input.OpenNewTab
		}
		item.CSSClass = input.CSSClass

		updated, err := h.uc.UpdateNavItem(ctx, item)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *PortalHandler) deleteNavItem() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		if err := h.uc.DeleteNavItem(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

func (h *PortalHandler) reorderNavItems() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			IDs []string `json:"ids" binding:"required"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if err := h.uc.ReorderNavItems(ctx, input.IDs); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

// ==================== Admin Banner Handlers ====================

func (h *PortalHandler) listBanners() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListBanners(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, items)
	}
}

func (h *PortalHandler) createBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Title             string            `json:"title" binding:"required"`
			TitleI18n         map[string]string `json:"title_i18n"`
			Subtitle          string            `json:"subtitle"`
			SubtitleI18n      map[string]string `json:"subtitle_i18n"`
			BadgeText         string            `json:"badge_text"`
			ImageURL          string            `json:"image_url"`
			ImageMobileURL    string            `json:"image_mobile_url"`
			BgColorStart      string            `json:"bg_color_start"`
			BgColorEnd        string            `json:"bg_color_end"`
			BgOverlayOpacity  float64           `json:"bg_overlay_opacity"`
			PrimaryBtnText    string            `json:"primary_btn_text"`
			PrimaryBtnURL     string            `json:"primary_btn_url"`
			SecondaryBtnText  string            `json:"secondary_btn_text"`
			SecondaryBtnURL   string            `json:"secondary_btn_url"`
			Sequence          int               `json:"sequence"`
			IsActive          *bool             `json:"is_active"`
			StartAt           *time.Time        `json:"start_at"`
			EndAt             *time.Time        `json:"end_at"`
			AutoSlideInterval int               `json:"auto_slide_interval"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}

		b := &dto.PortalBannerDTO{
			Title:             input.Title,
			TitleI18n:         input.TitleI18n,
			Subtitle:          input.Subtitle,
			SubtitleI18n:      input.SubtitleI18n,
			BadgeText:         input.BadgeText,
			ImageURL:          input.ImageURL,
			ImageMobileURL:    input.ImageMobileURL,
			BgColorStart:      input.BgColorStart,
			BgColorEnd:        input.BgColorEnd,
			BgOverlayOpacity:  input.BgOverlayOpacity,
			PrimaryBtnText:    input.PrimaryBtnText,
			PrimaryBtnURL:     input.PrimaryBtnURL,
			SecondaryBtnText:  input.SecondaryBtnText,
			SecondaryBtnURL:   input.SecondaryBtnURL,
			Sequence:          input.Sequence,
			IsActive:          isActive,
			AutoSlideInterval: input.AutoSlideInterval,
		}
		if input.StartAt != nil {
			b.StartAt = *input.StartAt
		}
		if input.EndAt != nil {
			b.EndAt = *input.EndAt
		}

		created, err := h.uc.CreateBanner(ctx, b)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, created)
	}
}

func (h *PortalHandler) updateBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var input struct {
			Title             string            `json:"title"`
			TitleI18n         map[string]string `json:"title_i18n"`
			Subtitle          string            `json:"subtitle"`
			SubtitleI18n      map[string]string `json:"subtitle_i18n"`
			BadgeText         string            `json:"badge_text"`
			ImageURL          string            `json:"image_url"`
			ImageMobileURL    string            `json:"image_mobile_url"`
			BgColorStart      string            `json:"bg_color_start"`
			BgColorEnd        string            `json:"bg_color_end"`
			BgOverlayOpacity  float64           `json:"bg_overlay_opacity"`
			PrimaryBtnText    string            `json:"primary_btn_text"`
			PrimaryBtnURL     string            `json:"primary_btn_url"`
			SecondaryBtnText  string            `json:"secondary_btn_text"`
			SecondaryBtnURL   string            `json:"secondary_btn_url"`
			Sequence          *int              `json:"sequence"`
			IsActive          *bool             `json:"is_active"`
			StartAt           *time.Time        `json:"start_at"`
			EndAt             *time.Time        `json:"end_at"`
			AutoSlideInterval *int              `json:"auto_slide_interval"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		b := &dto.PortalBannerDTO{ID: id}
		if input.Title != "" {
			b.Title = input.Title
		}
		if input.TitleI18n != nil {
			b.TitleI18n = input.TitleI18n
		}
		b.Subtitle = input.Subtitle
		if input.SubtitleI18n != nil {
			b.SubtitleI18n = input.SubtitleI18n
		}
		b.BadgeText = input.BadgeText
		b.ImageURL = input.ImageURL
		b.ImageMobileURL = input.ImageMobileURL
		b.BgColorStart = input.BgColorStart
		b.BgColorEnd = input.BgColorEnd
		b.BgOverlayOpacity = input.BgOverlayOpacity
		b.PrimaryBtnText = input.PrimaryBtnText
		b.PrimaryBtnURL = input.PrimaryBtnURL
		b.SecondaryBtnText = input.SecondaryBtnText
		b.SecondaryBtnURL = input.SecondaryBtnURL
		if input.Sequence != nil {
			b.Sequence = *input.Sequence
		}
		if input.IsActive != nil {
			b.IsActive = *input.IsActive
		}
		if input.StartAt != nil {
			b.StartAt = *input.StartAt
		}
		if input.EndAt != nil {
			b.EndAt = *input.EndAt
		}
		if input.AutoSlideInterval != nil {
			b.AutoSlideInterval = *input.AutoSlideInterval
		}

		updated, err := h.uc.UpdateBanner(ctx, b)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *PortalHandler) toggleBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		updated, err := h.uc.ToggleBanner(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *PortalHandler) getBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		banner, err := h.uc.GetBanner(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "banner not found")
		}
		return server.OKCtx(ctx, banner)
	}
}

func (h *PortalHandler) deleteBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		if err := h.uc.DeleteBanner(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

// ==================== Admin CustomPage Handlers ====================

func (h *PortalHandler) listCustomPages() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}

		items, total, err := h.uc.ListCustomPages(ctx, page, pageSize)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, map[string]any{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PortalHandler) createCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Title          string `json:"title" binding:"required"`
			Slug           string `json:"slug" binding:"required"`
			Type           string `json:"type"`
			ContentFormat  string `json:"content_format"`
			Content        string `json:"content"`
			Layout         string `json:"layout"`
			IsPublished    *bool  `json:"is_published"`
			SeoTitle       string `json:"seo_title"`
			SeoDescription string `json:"seo_description"`
			FeaturedImage  string `json:"featured_image"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		isPublished := false
		if input.IsPublished != nil {
			isPublished = *input.IsPublished
		}

		p := &dto.PortalCustomPageDTO{
			Title:          input.Title,
			Slug:           input.Slug,
			Type:           input.Type,
			ContentFormat:  input.ContentFormat,
			Content:        input.Content,
			Layout:         input.Layout,
			IsPublished:    isPublished,
			SeoTitle:       input.SeoTitle,
			SeoDescription: input.SeoDescription,
			FeaturedImage:  input.FeaturedImage,
		}
		if isPublished {
			p.PublishedAt = time.Now()
		}

		created, err := h.uc.CreateCustomPage(ctx, p)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, created)
	}
}

func (h *PortalHandler) getCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		page, err := h.uc.GetCustomPageByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "page not found")
		}
		return server.OKCtx(ctx, page)
	}
}

// ==================== Public Portal Config ====================

func (h *PortalHandler) getPortalConfig() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		modules := biz.PortalModules{
			Articles: getPortalBool(h.settingUC, ctx, "module_articles", true),
			Videos:   getPortalBool(h.settingUC, ctx, "module_videos", true),
			Music:    getPortalBool(h.settingUC, ctx, "module_music", false),
		}

		configuredLayout := h.settingUC.Get(ctx, "homepage_layout")
		if configuredLayout == "" {
			configuredLayout = "auto"
		}
		layout := resolvePortalLayout(modules, configuredLayout)

		site := biz.PortalSite{
			SiteName:          h.settingUC.GetNoCache(ctx, "site_name"),
			SiteDescription:   h.settingUC.GetNoCache(ctx, "site_description"),
			AllowRegistration: getPortalBool(h.settingUC, ctx, "allow_registration", true),
			AllowUpload:       getPortalBool(h.settingUC, ctx, "allow_upload", true),
			PrimaryURL:        h.settingUC.GetNoCache(ctx, "primary_url"),
			SiteLogoURL:       h.settingUC.GetNoCache(ctx, "site_logo_url"),
		}
		if urls := h.settingUC.GetNoCache(ctx, "base_urls"); urls != "" {
			_ = json.Unmarshal([]byte(urls), &site.AllowedURLs)
		}

		navItems, _ := h.uc.ListNavItems(ctx)
		banners, _ := h.uc.ListActiveBanners(ctx)
		categories, _ := h.catUC.ListActiveCategories(ctx)
		pages, _ := h.uc.ListPublishedCustomPages(ctx)

		features := h.featureUC.GetAll(ctx)

		resp := map[string]any{
			"modules":     modules,
			"layout":      layout,
			"site":        site,
			"navigation":  navItems,
			"banners":     banners,
			"categories":  categories,
			"pages":       pages,
			"features":    features,
		}

		return server.OKCtx(ctx, resp)
	}
}

func getPortalBool(uc *systembiz.SettingUseCase, ctx context.Context, key string, defaultValue bool) bool {
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

func resolvePortalLayout(modules biz.PortalModules, configuredLayout string) string {
	if configuredLayout != "auto" {
		if configuredLayout == "video" && !modules.Videos {
			return "welcome"
		}
		if configuredLayout == "article" && !modules.Articles {
			return "welcome"
		}
		return configuredLayout
	}
	if modules.Videos && modules.Articles {
		return "mixed"
	}
	if modules.Videos {
		return "video"
	}
	if modules.Articles {
		return "article"
	}
	return "welcome"
}

func (h *PortalHandler) updateCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var input struct {
			Title          string `json:"title"`
			Slug           string `json:"slug"`
			Type           string `json:"type"`
			ContentFormat  string `json:"content_format"`
			Content        string `json:"content"`
			Layout         string `json:"layout"`
			IsPublished    *bool  `json:"is_published"`
			SeoTitle       string `json:"seo_title"`
			SeoDescription string `json:"seo_description"`
			FeaturedImage  string `json:"featured_image"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existing, err := h.uc.GetCustomPageByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "page not found")
		}

		p := &dto.PortalCustomPageDTO{ID: id}
		if input.Title != "" {
			p.Title = input.Title
		} else {
			p.Title = existing.Title
		}
		if input.Slug != "" {
			p.Slug = input.Slug
		} else {
			p.Slug = existing.Slug
		}
		if input.Type != "" {
			p.Type = input.Type
		} else {
			p.Type = existing.Type
		}
		if input.ContentFormat != "" {
			p.ContentFormat = input.ContentFormat
		} else {
			p.ContentFormat = existing.ContentFormat
		}
		p.Content = input.Content
		if input.Layout != "" {
			p.Layout = input.Layout
		} else {
			p.Layout = existing.Layout
		}
		if input.IsPublished != nil {
			p.IsPublished = *input.IsPublished
			if *input.IsPublished && existing.PublishedAt.IsZero() {
				p.PublishedAt = time.Now()
			}
		} else {
			p.IsPublished = existing.IsPublished
			p.PublishedAt = existing.PublishedAt
		}
		p.SeoTitle = input.SeoTitle
		p.SeoDescription = input.SeoDescription
		p.FeaturedImage = input.FeaturedImage

		updated, err := h.uc.UpdateCustomPage(ctx, p)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *PortalHandler) deleteCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		if err := h.uc.DeleteCustomPage(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

// ==================== Public Page Handler ====================

func (h *PortalHandler) getPublicPageBySlug() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		slug := ctx.Var("slug")
		if slug == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "slug is required")
		}

		page, err := h.uc.GetCustomPageBySlug(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "page not found")
		}

		if !page.IsPublished {
			return server.FailCtx(ctx, server.ErrNotFound, "page not found")
		}

		_ = h.uc.IncrementPageViewCount(ctx, page.ID)

		return server.OKCtx(ctx, page)
	}
}
