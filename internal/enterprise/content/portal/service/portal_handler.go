package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"


	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/enterprise/content/portal/biz"
	adbiz "origadmin/application/origstudio/internal/enterprise/media/ad/biz"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/enterprise/content/portal/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type Handler struct {
	uc         *biz.UseCase
	adUC       *adbiz.UseCase
	catUC      *contentbiz.CategoryTagUseCase
	jwt        *auth.Manager
	settingUC  *systembiz.SettingUseCase
	featureUC  *systembiz.FeatureFlagUseCase
}

func NewHandler(uc *biz.UseCase, adUC *adbiz.UseCase, catUC *contentbiz.CategoryTagUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase, featureUC *systembiz.FeatureFlagUseCase) *Handler {
	return &Handler{uc: uc, adUC: adUC, catUC: catUC, jwt: jwt, settingUC: settingUC, featureUC: featureUC}
}

// validateBannerInput returns a human-readable error message when the banner
// payload is invalid, or an empty string when it is acceptable.
func validateBannerInput(b *dto.PortalBannerDTO) string {
	switch b.Type {
	case "", "custom", "video", "hot_videos", "new_videos", "ad":
		// allowed
	default:
		return "type must be one of: custom, video, hot_videos, new_videos, ad"
	}
	if (b.Type == "hot_videos" || b.Type == "new_videos") && b.Count <= 0 {
		return "count must be > 0 for video banners"
	}
	if b.StartAt != nil && !b.StartAt.IsZero() && b.EndAt != nil && !b.EndAt.IsZero() && b.StartAt.After(*b.EndAt) {
		return "start_at must be before end_at"
	}
	if b.ImageURL != "" {
		if err := validateMediaPath(b.ImageURL); err != "" {
			return err
		}
	}
	if b.ImageMobileURL != "" {
		if err := validateMediaPath(b.ImageMobileURL); err != "" {
			return err
		}
	}
	if b.VideoURL != "" {
		if err := validateMediaPath(b.VideoURL); err != "" {
			return err
		}
	}
	if b.AutoSlideInterval < 0 {
		return "auto_slide_interval must be >= 0"
	}
	return ""
}

// validateMediaPath accepts: http(s) URLs, /absolute paths, relative storage paths, data URIs
func validateMediaPath(path string) string {
	if path == "" {
		return ""
	}
	// Allow data URIs (base64 images)
	if strings.HasPrefix(path, "data:") {
		return ""
	}
	// Allow http(s) URLs
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if u, err := url.Parse(path); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return "must be a valid http(s) URL"
		}
		return ""
	}
	// Allow absolute server paths (e.g. /files/xxx, /media/xxx)
	if strings.HasPrefix(path, "/") {
		return ""
	}
	// Allow relative storage paths (e.g. "bucket/key.jpg", "xxx.mp4")
	return ""
}

func (h *Handler) RegisterRoutes(r http2.Router) {
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
		adminBanners.POST("/suggest", h.suggestBanners())
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

func (h *Handler) listNavItems() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListNavItems(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": len(items)})
		return nil
	}
}

func (h *Handler) createNavItem() http2.HandlerFunc {
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updateNavItem() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		existing, err := h.uc.GetNavItemByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "nav item not found")
			return nil
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deleteNavItem() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}

		if err := h.uc.DeleteNavItem(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) reorderNavItems() http2.HandlerFunc {
	return func(ctx http2.Context) error {

		var input struct {
			IDs []string `json:"ids" binding:"required"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if err := h.uc.ReorderNavItems(ctx, input.IDs); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listBanners() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListBanners(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": len(items)})
		return nil
	}
}

func (h *Handler) createBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {

		var input struct {
			Title             string            `json:"title"`
			TitleI18n         map[string]string `json:"title_i18n"`
			Subtitle          string            `json:"subtitle"`
			SubtitleI18n      map[string]string `json:"subtitle_i18n"`
			BadgeText         string            `json:"badge_text"`
			ImageURL          string            `json:"image_url"`
			ImageMobileURL    string            `json:"image_mobile_url"`
			VideoURL          string            `json:"video_url"`
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
			Type              string            `json:"type"`
			Count             int               `json:"count"`
			CategoryID        string            `json:"category_id"`
			DisplayMode       string            `json:"display_mode"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}

		bannerType := input.Type
		if bannerType == "" {
			bannerType = "custom"
		}
		count := input.Count
		if count <= 0 {
			count = 5
		}

		displayMode := input.DisplayMode
		if displayMode == "" {
			displayMode = "wide"
		}

		title := input.Title
		if title == "" {
			switch bannerType {
			case "hot_videos":
				title = "最火视频"
			case "new_videos":
				title = "最新上线"
			case "ad":
				title = "广告位"
			}
		}

		subtitle := input.Subtitle
		if subtitle == "" {
			switch bannerType {
			case "hot_videos":
				subtitle = "大家都在看"
			case "new_videos":
				subtitle = "抢先看新鲜内容"
			}
		}

		badgeText := input.BadgeText
		if badgeText == "" {
			switch bannerType {
			case "hot_videos":
				badgeText = "HOT"
			case "new_videos":
				badgeText = "NEW"
			}
		}

		primaryBtnText := input.PrimaryBtnText
		primaryBtnURL := input.PrimaryBtnURL
		if (bannerType == "hot_videos" || bannerType == "new_videos") && primaryBtnURL == "" {
			primaryBtnURL = "/videos"
			if primaryBtnText == "" {
				primaryBtnText = "立即观看"
			}
		}

		autoSlide := input.AutoSlideInterval
		if autoSlide <= 0 {
			autoSlide = 5000
		}

		b := &dto.PortalBannerDTO{
			Title:             title,
			TitleI18n:         input.TitleI18n,
			Subtitle:          subtitle,
			SubtitleI18n:      input.SubtitleI18n,
			BadgeText:         badgeText,
			ImageURL:          input.ImageURL,
			ImageMobileURL:    input.ImageMobileURL,
			VideoURL:          input.VideoURL,
			BgColorStart:      input.BgColorStart,
			BgColorEnd:        input.BgColorEnd,
			BgOverlayOpacity:  input.BgOverlayOpacity,
			PrimaryBtnText:    primaryBtnText,
			PrimaryBtnURL:     primaryBtnURL,
			SecondaryBtnText:  input.SecondaryBtnText,
			SecondaryBtnURL:   input.SecondaryBtnURL,
			Sequence:          input.Sequence,
			IsActive:          isActive,
			AutoSlideInterval: autoSlide,
			Type:              bannerType,
			Count:             count,
			CategoryID:        input.CategoryID,
			DisplayMode:       displayMode,
		}
		if input.StartAt != nil {
			b.StartAt = input.StartAt
		}
		if input.EndAt != nil {
			b.EndAt = input.EndAt
		}

		if msg := validateBannerInput(b); msg != "" {
			http2.Fail(ctx, server.ErrBadRequest, msg)
			return nil
		}

		created, err := h.uc.CreateBanner(ctx, b)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updateBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}

		existing, err := h.uc.GetBanner(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "banner not found")
			return nil
		}

		var input struct {
			Title             *string           `json:"title"`
			TitleI18n         *map[string]string `json:"title_i18n"`
			Subtitle          *string           `json:"subtitle"`
			SubtitleI18n      *map[string]string `json:"subtitle_i18n"`
			BadgeText         *string           `json:"badge_text"`
			ImageURL          *string           `json:"image_url"`
			ImageMobileURL    *string           `json:"image_mobile_url"`
			VideoURL          *string           `json:"video_url"`
			BgColorStart      *string           `json:"bg_color_start"`
			BgColorEnd        *string           `json:"bg_color_end"`
			BgOverlayOpacity  *float64          `json:"bg_overlay_opacity"`
			PrimaryBtnText    *string           `json:"primary_btn_text"`
			PrimaryBtnURL     *string           `json:"primary_btn_url"`
			SecondaryBtnText  *string           `json:"secondary_btn_text"`
			SecondaryBtnURL   *string           `json:"secondary_btn_url"`
			Sequence          *int              `json:"sequence"`
			IsActive          *bool             `json:"is_active"`
			StartAt           *time.Time        `json:"start_at"`
			EndAt             *time.Time        `json:"end_at"`
			ClearEndAt        bool              `json:"clear_end_at"`
			ClearStartAt      bool              `json:"clear_start_at"`
			AutoSlideInterval *int              `json:"auto_slide_interval"`
			Type              *string           `json:"type"`
			Count             *int              `json:"count"`
			CategoryID        *string           `json:"category_id"`
			DisplayMode       *string           `json:"display_mode"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		// Start from the existing values so the update is non-destructive:
		// only fields explicitly provided in the request override stored ones.
		b := &dto.PortalBannerDTO{
			ID:                id,
			Title:             existing.Title,
			TitleI18n:         existing.TitleI18n,
			Subtitle:          existing.Subtitle,
			SubtitleI18n:      existing.SubtitleI18n,
			BadgeText:         existing.BadgeText,
			ImageURL:          existing.ImageURL,
			ImageMobileURL:    existing.ImageMobileURL,
			VideoURL:          existing.VideoURL,
			BgColorStart:      existing.BgColorStart,
			BgColorEnd:        existing.BgColorEnd,
			BgOverlayOpacity:  existing.BgOverlayOpacity,
			PrimaryBtnText:    existing.PrimaryBtnText,
			PrimaryBtnURL:     existing.PrimaryBtnURL,
			SecondaryBtnText:  existing.SecondaryBtnText,
			SecondaryBtnURL:   existing.SecondaryBtnURL,
			Sequence:          existing.Sequence,
			IsActive:          existing.IsActive,
			StartAt:           existing.StartAt,
			EndAt:             existing.EndAt,
			AutoSlideInterval: existing.AutoSlideInterval,
			Type:              existing.Type,
			Count:             existing.Count,
			CategoryID:        existing.CategoryID,
			DisplayMode:       existing.DisplayMode,
		}
		if input.Title != nil {
			b.Title = *input.Title
		}
		if input.TitleI18n != nil {
			b.TitleI18n = *input.TitleI18n
		}
		if input.Subtitle != nil {
			b.Subtitle = *input.Subtitle
		}
		if input.SubtitleI18n != nil {
			b.SubtitleI18n = *input.SubtitleI18n
		}
		if input.BadgeText != nil {
			b.BadgeText = *input.BadgeText
		}
		if input.ImageURL != nil {
			b.ImageURL = *input.ImageURL
		}
		if input.ImageMobileURL != nil {
			b.ImageMobileURL = *input.ImageMobileURL
		}
		if input.VideoURL != nil {
			b.VideoURL = *input.VideoURL
		}
		if input.BgColorStart != nil {
			b.BgColorStart = *input.BgColorStart
		}
		if input.BgColorEnd != nil {
			b.BgColorEnd = *input.BgColorEnd
		}
		if input.BgOverlayOpacity != nil {
			b.BgOverlayOpacity = *input.BgOverlayOpacity
		}
		if input.PrimaryBtnText != nil {
			b.PrimaryBtnText = *input.PrimaryBtnText
		}
		if input.PrimaryBtnURL != nil {
			b.PrimaryBtnURL = *input.PrimaryBtnURL
		}
		if input.SecondaryBtnText != nil {
			b.SecondaryBtnText = *input.SecondaryBtnText
		}
		if input.SecondaryBtnURL != nil {
			b.SecondaryBtnURL = *input.SecondaryBtnURL
		}
		if input.Sequence != nil {
			b.Sequence = *input.Sequence
		}
		if input.IsActive != nil {
			b.IsActive = *input.IsActive
		}
		if input.StartAt != nil {
			b.StartAt = input.StartAt
		}
		if input.EndAt != nil {
			b.EndAt = input.EndAt
		}
		if input.ClearEndAt {
			b.ClearEndAt = true
			b.EndAt = nil
		}
		if input.ClearStartAt {
			b.ClearStartAt = true
			b.StartAt = nil
		}
		if input.AutoSlideInterval != nil {
			b.AutoSlideInterval = *input.AutoSlideInterval
		}
		if input.Type != nil {
			b.Type = *input.Type
		}
		if input.Count != nil {
			b.Count = *input.Count
		}
		if input.CategoryID != nil {
			b.CategoryID = *input.CategoryID
		}
		if input.DisplayMode != nil {
			b.DisplayMode = *input.DisplayMode
		}

		if msg := validateBannerInput(b); msg != "" {
			http2.Fail(ctx, server.ErrBadRequest, msg)
			return nil
		}

		updated, err := h.uc.UpdateBanner(ctx, b)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) toggleBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}

		updated, err := h.uc.ToggleBanner(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) getBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		b, err := h.uc.GetBanner(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "banner not found")
			return nil
		}
		http2.OK(ctx, b)
		return nil
	}
}

func (h *Handler) deleteBanner() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeleteBanner(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrNotFound, "banner not found")
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"success": true, "id": id})
		return nil
	}
}

// suggestBanners seeds a sensible default portal banner set: a hero custom
// banner plus auto-aggregated "hot videos" and "new videos" rails. It is safe
// to call repeatedly — duplicates can be removed from the admin UI.
func (h *Handler) suggestBanners() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		suggestions := []*dto.PortalBannerDTO{
			{
				Type:              "custom",
				Title:             "欢迎来到我们的视频门户",
				Subtitle:          "发现最热门与最新上线的精彩内容",
				Sequence:          0,
				IsActive:          true,
				AutoSlideInterval: 5000,
				BgColorStart:      "#4f46e5",
				BgColorEnd:        "#7c3aed",
				PrimaryBtnText:    "立即浏览",
				PrimaryBtnURL:     "/videos",
			},
			{
				Type:              "hot_videos",
				Title:             "最火视频",
				Subtitle:          "大家都在看",
				Sequence:          1,
				IsActive:          false,
				Count:             5,
				AutoSlideInterval: 5000,
			},
			{
				Type:              "new_videos",
				Title:             "最新上线",
				Subtitle:          "抢先看新鲜内容",
				Sequence:          2,
				IsActive:          false,
				Count:             5,
				AutoSlideInterval: 5000,
			},
		}

		created := make([]*dto.PortalBannerDTO, 0, len(suggestions))
		for _, s := range suggestions {
			c, err := h.uc.CreateBanner(ctx, s)
			if err != nil {
				http2.Fail(ctx, server.ErrInternal, err.Error())
				return nil
			}
			created = append(created, c)
		}
		http2.OK(ctx, map[string]interface{}{"items": created, "total": len(created)})
		return nil
	}
}

func (h *Handler) listCustomPages() http2.HandlerFunc {
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, map[string]interface{}{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
		return nil
	}
}

func (h *Handler) createCustomPage() http2.HandlerFunc {
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) getCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}

		page, err := h.uc.GetCustomPageByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "page not found")
			return nil
		}
		http2.OK(ctx, page)
		return nil
	}
}

func (h *Handler) updateCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		existing, err := h.uc.GetCustomPageByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "page not found")
			return nil
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deleteCustomPage() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}

		if err := h.uc.DeleteCustomPage(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) getPublicPageBySlug() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		slug := ctx.Var("slug")
		if slug == "" {
			http2.Fail(ctx, server.ErrBadRequest, "slug is required")
			return nil
		}

		page, err := h.uc.GetCustomPageBySlug(ctx, slug)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "page not found")
			return nil
		}

		if !page.IsPublished {
			http2.Fail(ctx, server.ErrNotFound, "page not found")
			return nil
		}

		_ = h.uc.IncrementPageViewCount(ctx, page.ID)

		http2.OK(ctx, page)
		return nil
	}
}

func (h *Handler) getPortalConfig() http2.HandlerFunc {
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
		allCategories, _ := h.catUC.ListCategories(ctx)
		var categories []*contentbiz.Category
		for _, cat := range allCategories {
			if cat.Status == 1 {
				categories = append(categories, cat)
			}
		}
		pages, _ := h.uc.ListPublishedCustomPages(ctx)

		features := h.featureUC.GetAll(ctx)

		resp := map[string]interface{}{
			"modules":     modules,
			"layout":      layout,
			"site":        site,
			"navigation":  navItems,
			"banners":     banners,
			"categories":  categories,
			"pages":       pages,
			"features":    features,
			"share":       parseSharePlatforms(h.settingUC, ctx),
		}

		http2.OK(ctx, resp)
		return nil
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

// sharePlatformKeys is the canonical set of link-based share platforms.
var sharePlatformKeys = []string{"twitter", "facebook", "whatsapp", "telegram", "linkedin", "weibo"}

// parseSharePlatforms reads the share_platforms setting (JSON map of
// platform->enabled) and returns a map containing every known platform key,
// defaulting missing/unparseable entries to enabled so the share dialog never
// silently drops a platform. Mirrors the CE implementation in
// internal/features/system/service/system_handler.go so the EE acceptance
// stack (microservices) returns the same share contract as the monolith.
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

func resolvePortalLayout(modules biz.PortalModules, configuredLayout string) string {
	if configuredLayout != "auto" {
		switch configuredLayout {
		case "mixed":
			// mixed = 视频+文章混合首页。文章模块未实现/未启用时不得进入 mixed，
			// 否则前端会渲染未实现的"文章"Tab（回归根因之一）。降级到可用的布局。
			switch {
			case modules.Videos && modules.Articles:
				return "mixed"
			case modules.Videos:
				return "video"
			case modules.Articles:
				return "article"
			default:
				return "welcome"
			}
		case "video":
			if !modules.Videos {
				return "welcome"
			}
		case "article":
			if !modules.Articles {
				return "welcome"
			}
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
