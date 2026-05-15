package service

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	http2 "origadmin/application/origstudio/internal/helpers/http"
	ginadapter "origadmin/application/origstudio/internal/helpers/http/gin"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type PortalHandler struct {
	uc        *biz.PortalUseCase
	jwt       *auth.Manager
	settingUC *systembiz.SettingUseCase
}

func NewPortalHandler(uc *biz.PortalUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase) *PortalHandler {
	return &PortalHandler{uc: uc, jwt: jwt, settingUC: settingUC}
}

func (h *PortalHandler) RegisterRoutes(r http2.Router) {
	adminNavItems := r.Group("/admin/nav-items")
	adminNavItems.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminNavItems.GET("", server.HTTPToHandlerFunc(h.listNavItems()))
		adminNavItems.POST("", server.HTTPToHandlerFunc(h.createNavItem()))
		adminNavItems.PUT("/:id", server.HTTPToHandlerFunc(h.updateNavItem()))
		adminNavItems.DELETE("/:id", server.HTTPToHandlerFunc(h.deleteNavItem()))
		adminNavItems.PUT("/reorder", server.HTTPToHandlerFunc(h.reorderNavItems()))
	}

	adminBanners := r.Group("/admin/banners")
	adminBanners.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminBanners.GET("", server.HTTPToHandlerFunc(h.listBanners()))
		adminBanners.POST("", server.HTTPToHandlerFunc(h.createBanner()))
		adminBanners.PUT("/:id", server.HTTPToHandlerFunc(h.updateBanner()))
		adminBanners.POST("/:id/toggle", server.HTTPToHandlerFunc(h.toggleBanner()))
	}

	adminPages := r.Group("/admin/pages")
	adminPages.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminPages.GET("", server.HTTPToHandlerFunc(h.listCustomPages()))
		adminPages.POST("", server.HTTPToHandlerFunc(h.createCustomPage()))
		adminPages.GET("/:id", server.HTTPToHandlerFunc(h.getCustomPage()))
		adminPages.PUT("/:id", server.HTTPToHandlerFunc(h.updateCustomPage()))
		adminPages.DELETE("/:id", server.HTTPToHandlerFunc(h.deleteCustomPage()))
	}

	pages := r.Group("/p")
	{
		pages.GET("/:slug", server.HTTPToHandlerFunc(h.getPublicPageBySlug()))
	}
}

// ==================== Admin NavItem Handlers ====================

func (h *PortalHandler) listNavItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		items, err := h.uc.ListNavItems(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, items)
	}
}

func (h *PortalHandler) createNavItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)

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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		isVisible := true
		if input.IsVisible != nil {
			isVisible = *input.IsVisible
		}
		openNewTab := false
		if input.OpenNewTab != nil {
			openNewTab = *input.OpenNewTab
		}

		item := &entity.PortalNavItem{
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

		created, err := h.uc.CreateNavItem(r.Context(), item)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, created)
	}
}

func (h *PortalHandler) updateNavItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		existing, err := h.uc.GetNavItemByID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "nav item not found")
			return
		}
		_ = existing

		item := &entity.PortalNavItem{
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

		updated, err := h.uc.UpdateNavItem(r.Context(), item)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, updated)
	}
}

func (h *PortalHandler) deleteNavItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		if err := h.uc.DeleteNavItem(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, nil)
	}
}

func (h *PortalHandler) reorderNavItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)

		var input struct {
			IDs []string `json:"ids" binding:"required"`
		}

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if err := h.uc.ReorderNavItems(r.Context(), input.IDs); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, nil)
	}
}

// ==================== Admin Banner Handlers ====================

func (h *PortalHandler) listBanners() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		items, err := h.uc.ListBanners(r.Context())
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, items)
	}
}

func (h *PortalHandler) createBanner() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)

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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}

		b := &entity.PortalBanner{
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

		created, err := h.uc.CreateBanner(r.Context(), b)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, created)
	}
}

func (h *PortalHandler) updateBanner() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		b := &entity.PortalBanner{ID: id}
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

		updated, err := h.uc.UpdateBanner(r.Context(), b)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, updated)
	}
}

func (h *PortalHandler) toggleBanner() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		updated, err := h.uc.ToggleBanner(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, updated)
	}
}

// ==================== Admin CustomPage Handlers ====================

func (h *PortalHandler) listCustomPages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		page, _ := strconv.Atoi(gc.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(gc.DefaultQuery("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}

		items, total, err := h.uc.ListCustomPages(r.Context(), page, pageSize)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func (h *PortalHandler) createCustomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)

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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		isPublished := false
		if input.IsPublished != nil {
			isPublished = *input.IsPublished
		}

		p := &entity.PortalCustomPage{
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

		created, err := h.uc.CreateCustomPage(r.Context(), p)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, created)
	}
}

func (h *PortalHandler) getCustomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		page, err := h.uc.GetCustomPageByID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "page not found")
			return
		}
		server.OK(gc, page)
	}
}

func (h *PortalHandler) updateCustomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
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

		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		existing, err := h.uc.GetCustomPageByID(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "page not found")
			return
		}

		p := &entity.PortalCustomPage{ID: id}
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

		updated, err := h.uc.UpdateCustomPage(r.Context(), p)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, updated)
	}
}

func (h *PortalHandler) deleteCustomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "id is required")
			return
		}

		if err := h.uc.DeleteCustomPage(r.Context(), id); err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, nil)
	}
}

// ==================== Public Page Handler ====================

func (h *PortalHandler) getPublicPageBySlug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		slug := gc.Param("slug")
		if slug == "" {
			server.Fail(gc, server.ErrBadRequest, "slug is required")
			return
		}

		page, err := h.uc.GetCustomPageBySlug(r.Context(), slug)
		if err != nil {
			server.Fail(gc, server.ErrNotFound, "page not found")
			return
		}

		if !page.IsPublished {
			server.Fail(gc, server.ErrNotFound, "page not found")
			return
		}

		_ = h.uc.IncrementPageViewCount(r.Context(), page.ID)

		server.OK(gc, page)
	}
}


