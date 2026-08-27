package service

import (
	"strconv"
	"time"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/content/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

type AdHandler struct {
	uc  *biz.AdUseCase
	jwt *auth.Manager
}

func NewAdHandler(uc *biz.AdUseCase, jwt *auth.Manager) *AdHandler {
	return &AdHandler{uc: uc, jwt: jwt}
}

func (h *AdHandler) RegisterRoutes(r http2.Router) {
	adminPlacements := r.Group("/admin/ad-placements")
	adminPlacements.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminPlacements.GET("", h.listPlacements())
		adminPlacements.POST("", h.createPlacement())
		adminPlacements.PUT("/:id", h.updatePlacement())
		adminPlacements.POST("/:id/toggle", h.togglePlacement())
		adminPlacements.DELETE("/:id", h.deletePlacement())
		adminPlacements.GET("/ads-count", h.countAdsByPlacement())
	}

	adminAds := r.Group("/admin/ads")
	adminAds.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminAds.GET("", h.listAds())
		adminAds.POST("", h.createAd())
		adminAds.PUT("/:id", h.updateAd())
		adminAds.POST("/:id/toggle", h.toggleAd())
		adminAds.DELETE("/:id", h.deleteAd())
		adminAds.GET("/:id/click-logs", h.listClickLogs())
	}

	publicAds := r.Group("/ads")
	{
		publicAds.GET("/placements", h.listActivePlacementsWithAds())
		publicAds.GET("/placement/:slug", h.listActiveAdsByPlacement())
		publicAds.POST("/:id/impression", h.recordImpression())
		publicAds.POST("/:id/click", h.recordClick())
	}
}

// ==================== Admin Placement Handlers ====================

func (h *AdHandler) listPlacements() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListPlacements(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, items)
	}
}

func (h *AdHandler) createPlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Name        string `json:"name" binding:"required"`
			Slug        string `json:"slug" binding:"required"`
			Type        string `json:"type" binding:"required"`
			Description string `json:"description"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			MaxAds      int    `json:"max_ads"`
			IsActive    *bool  `json:"is_active"`
			Sequence    int    `json:"sequence"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		maxAds := 1
		if input.MaxAds > 0 {
			maxAds = input.MaxAds
		}

		p := &dto.AdPlacementDTO{
			Name:        input.Name,
			Slug:        input.Slug,
			Type:        input.Type,
			Description: input.Description,
			Width:       input.Width,
			Height:      input.Height,
			MaxAds:      maxAds,
			IsActive:    isActive,
			Sequence:    input.Sequence,
		}

		created, err := h.uc.CreatePlacement(ctx, p)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, created)
	}
}

func (h *AdHandler) updatePlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var input struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Type        string `json:"type"`
			Description string `json:"description"`
			Width       *int   `json:"width"`
			Height      *int   `json:"height"`
			MaxAds      *int   `json:"max_ads"`
			IsActive    *bool  `json:"is_active"`
			Sequence    *int   `json:"sequence"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existing, err := h.uc.GetPlacementByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "placement not found")
		}

		p := &dto.AdPlacementDTO{ID: id}
		if input.Name != "" {
			p.Name = input.Name
		} else {
			p.Name = existing.Name
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
		p.Description = input.Description
		if input.Width != nil {
			p.Width = *input.Width
		} else {
			p.Width = existing.Width
		}
		if input.Height != nil {
			p.Height = *input.Height
		} else {
			p.Height = existing.Height
		}
		if input.MaxAds != nil {
			p.MaxAds = *input.MaxAds
		} else {
			p.MaxAds = existing.MaxAds
		}
		if input.IsActive != nil {
			p.IsActive = *input.IsActive
		} else {
			p.IsActive = existing.IsActive
		}
		if input.Sequence != nil {
			p.Sequence = *input.Sequence
		} else {
			p.Sequence = existing.Sequence
		}

		updated, err := h.uc.UpdatePlacement(ctx, p)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *AdHandler) togglePlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		updated, err := h.uc.TogglePlacement(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *AdHandler) deletePlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		// 级联删除：DeletePlacement内部会先删除关联广告再删除广告位
		if err := h.uc.DeletePlacement(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

// countAdsByPlacement 返回指定广告位下的广告数量（用于删除前提示）
func (h *AdHandler) countAdsByPlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placementID := ctx.QueryVar("placement_id")
		if placementID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "placement_id is required")
		}

		count, err := h.uc.CountAdsByPlacement(ctx, placementID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, map[string]interface{}{
			"count": count,
		})
	}
}

// ==================== Admin Ad Handlers ====================

func (h *AdHandler) listAds() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placementID := ctx.QueryVar("placement_id")
		if placementID == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "placement_id is required")
		}

		items, total, err := h.uc.ListAds(ctx, placementID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, map[string]any{
			"items": items,
			"total": total,
		})
	}
}

func (h *AdHandler) createAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			PlacementID    string            `json:"placement_id" binding:"required"`
			Title          string            `json:"title" binding:"required"`
			TitleI18n      map[string]string `json:"title_i18n"`
			ImageURL       string            `json:"image_url"`
			ImageMobileURL string            `json:"image_mobile_url"`
			LinkURL        string            `json:"link_url"`
			LinkTarget     string            `json:"link_target"`
			BadgeText      string            `json:"badge_text"`
			Priority       int               `json:"priority"`
			IsActive       *bool             `json:"is_active"`
			StartAt        *time.Time        `json:"start_at"`
			EndAt          *time.Time        `json:"end_at"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		linkTarget := "_blank"
		if input.LinkTarget != "" {
			linkTarget = input.LinkTarget
		}

		a := &dto.AdDTO{
			PlacementID:    input.PlacementID,
			Title:          input.Title,
			TitleI18n:      input.TitleI18n,
			ImageURL:       input.ImageURL,
			ImageMobileURL: input.ImageMobileURL,
			LinkURL:        input.LinkURL,
			LinkTarget:     linkTarget,
			BadgeText:      input.BadgeText,
			Priority:       input.Priority,
			IsActive:       isActive,
		}
		if input.StartAt != nil {
			a.StartAt = *input.StartAt
		}
		if input.EndAt != nil {
			a.EndAt = *input.EndAt
		}

		created, err := h.uc.CreateAd(ctx, a)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, created)
	}
}

func (h *AdHandler) updateAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		var input struct {
			PlacementID    string            `json:"placement_id"`
			Title          string            `json:"title"`
			TitleI18n      map[string]string `json:"title_i18n"`
			ImageURL       string            `json:"image_url"`
			ImageMobileURL string            `json:"image_mobile_url"`
			LinkURL        string            `json:"link_url"`
			LinkTarget     string            `json:"link_target"`
			BadgeText      string            `json:"badge_text"`
			Priority       *int              `json:"priority"`
			IsActive       *bool             `json:"is_active"`
			StartAt        *time.Time        `json:"start_at"`
			EndAt          *time.Time        `json:"end_at"`
		}

		if err := ctx.BindJSON(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		existing, err := h.uc.GetAdByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "ad not found")
		}

		a := &dto.AdDTO{ID: id}
		if input.PlacementID != "" {
			a.PlacementID = input.PlacementID
		} else {
			a.PlacementID = existing.PlacementID
		}
		if input.Title != "" {
			a.Title = input.Title
		} else {
			a.Title = existing.Title
		}
		if input.TitleI18n != nil {
			a.TitleI18n = input.TitleI18n
		} else {
			a.TitleI18n = existing.TitleI18n
		}
		a.ImageURL = input.ImageURL
		a.ImageMobileURL = input.ImageMobileURL
		a.LinkURL = input.LinkURL
		if input.LinkTarget != "" {
			a.LinkTarget = input.LinkTarget
		} else {
			a.LinkTarget = existing.LinkTarget
		}
		a.BadgeText = input.BadgeText
		if input.Priority != nil {
			a.Priority = *input.Priority
		} else {
			a.Priority = existing.Priority
		}
		if input.IsActive != nil {
			a.IsActive = *input.IsActive
		} else {
			a.IsActive = existing.IsActive
		}
		if input.StartAt != nil {
			a.StartAt = *input.StartAt
		} else {
			a.StartAt = existing.StartAt
		}
		if input.EndAt != nil {
			a.EndAt = *input.EndAt
		} else {
			a.EndAt = existing.EndAt
		}

		updated, err := h.uc.UpdateAd(ctx, a)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *AdHandler) toggleAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		updated, err := h.uc.ToggleAd(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, updated)
	}
}

func (h *AdHandler) deleteAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		if err := h.uc.DeleteAd(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

func (h *AdHandler) listClickLogs() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}

		items, total, err := h.uc.ListClickLogs(ctx, id, page, pageSize)
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

// ==================== Public Ad Handlers ====================

func (h *AdHandler) listActiveAdsByPlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		slug := ctx.Var("slug")
		if slug == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "slug is required")
		}

		items, err := h.uc.ListActiveAdsByPlacement(ctx, slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, items)
	}
}

func (h *AdHandler) recordImpression() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		if err := h.uc.RecordImpression(ctx, id); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, nil)
	}
}

func (h *AdHandler) recordClick() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "id is required")
		}

		ad, err := h.uc.GetAdByID(ctx, id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "ad not found")
		}

		clickLog := &dto.AdClickLogDTO{
			AdID:        id,
			PlacementID: ad.PlacementID,
			IP:          ctx.ClientIP(),
			UserAgent:   ctx.Request().UserAgent(),
			Referer:     ctx.Request().Referer(),
		}

		if userID, exists := ctx.Get("user_id"); exists {
			if uid, ok := userID.(string); ok {
				clickLog.UserID = uid
			}
		}

		if err := h.uc.RecordClick(ctx, id, clickLog); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, map[string]any{"url": ad.LinkURL})
	}
}

func (h *AdHandler) listActivePlacementsWithAds() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListActivePlacementsWithAds(ctx)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		if items == nil {
			items = []*dto.AdPlacementWithAdsDTO{}
		}
		return server.OKCtx(ctx, items)
	}
}
