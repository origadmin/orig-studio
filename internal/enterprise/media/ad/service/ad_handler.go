package service

import (
	"strconv"
	"time"


	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/enterprise/media/ad/biz"
	"origadmin/application/origstudio/internal/enterprise/media/ad/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

type Handler struct {
	uc  *biz.UseCase
	jwt *auth.Manager
}

func NewHandler(uc *biz.UseCase, jwt *auth.Manager) *Handler {
	return &Handler{uc: uc, jwt: jwt}
}

func (h *Handler) RegisterRoutes(r http2.Router) {
	adminPlacements := r.Group("/admin/ad-placements")
	adminPlacements.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminPlacements.GET("", h.listPlacements())
		adminPlacements.POST("", h.createPlacement())
		adminPlacements.PUT("/:id", h.updatePlacement())
		adminPlacements.POST("/:id/toggle", h.togglePlacement())
		adminPlacements.DELETE("/:id", h.deletePlacement())
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
		publicAds.GET("", h.listActiveAdsByPlacement())
		publicAds.GET("/placements", h.listActivePlacementsWithAds())
	publicAds.POST("/:id/impression", h.recordImpression())
	publicAds.POST("/:id/click", h.recordClick())
	// Creative telemetry (BUG-173): public endpoints so the portal player can
	// report impressions/clicks for AdCreative items (no auth, mirrors /ads).
	publicCreatives := r.Group("/creatives")
	{
		publicCreatives.POST("/:id/impression", h.recordCreativeImpression())
		publicCreatives.POST("/:id/click", h.recordCreativeClick())
	}
}

	adminCreatives := r.Group("/admin/creatives")
	adminCreatives.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminCreatives.GET("", h.listCreatives())
		adminCreatives.POST("", h.createCreative())
		adminCreatives.GET("/:id", h.getCreative())
		adminCreatives.PUT("/:id", h.updateCreative())
		adminCreatives.DELETE("/:id", h.deleteCreative())
	}

	adminPlacementCreatives := r.Group("/admin/ad-placements/:id/creatives")
	adminPlacementCreatives.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminPlacementCreatives.GET("", h.listPlacementCreativeIDs())
		adminPlacementCreatives.POST("", h.assignCreative())
		adminPlacementCreatives.DELETE("/:creativeId", h.unassignCreative())
	}
}

func (h *Handler) listPlacements() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListPlacements(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) createPlacement() http2.HandlerFunc {
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
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
			Name: input.Name, Slug: input.Slug, Type: input.Type,
			Description: input.Description, Width: input.Width, Height: input.Height,
			MaxAds: maxAds, IsActive: isActive, Sequence: input.Sequence,
		}
		created, err := h.uc.CreatePlacement(ctx, p)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updatePlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		existing, err := h.uc.GetPlacementByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "placement not found")
			return nil
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) togglePlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		updated, err := h.uc.TogglePlacement(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deletePlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeletePlacement(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listAds() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placementID := ctx.QueryVar("placement_id")
		if placementID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "placement_id is required")
			return nil
		}
		items, total, err := h.uc.ListAds(ctx, placementID)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total})
		return nil
	}
}

func (h *Handler) createAd() http2.HandlerFunc {
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
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
			PlacementID: input.PlacementID, Title: input.Title, TitleI18n: input.TitleI18n,
			ImageURL: input.ImageURL, ImageMobileURL: input.ImageMobileURL,
			LinkURL: input.LinkURL, LinkTarget: linkTarget, BadgeText: input.BadgeText,
			Priority: input.Priority, IsActive: isActive,
		}
		if input.StartAt != nil {
			a.StartAt = *input.StartAt
		}
		if input.EndAt != nil {
			a.EndAt = *input.EndAt
		}
		created, err := h.uc.CreateAd(ctx, a)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updateAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
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
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		existing, err := h.uc.GetAdByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "ad not found")
			return nil
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
	if input.ImageURL != "" {
		a.ImageURL = input.ImageURL
	} else {
		a.ImageURL = existing.ImageURL
	}
	if input.ImageMobileURL != "" {
		a.ImageMobileURL = input.ImageMobileURL
	} else {
		a.ImageMobileURL = existing.ImageMobileURL
	}
	if input.LinkURL != "" {
		a.LinkURL = input.LinkURL
	} else {
		a.LinkURL = existing.LinkURL
	}
	if input.LinkTarget != "" {
		a.LinkTarget = input.LinkTarget
	} else {
		a.LinkTarget = existing.LinkTarget
	}
	if input.BadgeText != "" {
		a.BadgeText = input.BadgeText
	} else {
		a.BadgeText = existing.BadgeText
	}
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) toggleAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		updated, err := h.uc.ToggleAd(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deleteAd() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeleteAd(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listClickLogs() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
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
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total, "page": page, "page_size": pageSize})
		return nil
	}
}

func (h *Handler) listActiveAdsByPlacement() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placement := ctx.QueryVar("placement")
		if placement == "" {
			http2.Fail(ctx, server.ErrBadRequest, "placement query parameter is required")
			return nil
		}
		items, err := h.uc.ListActiveAdsByPlacement(ctx, placement)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) listActivePlacementsWithAds() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListActivePlacementsWithAds(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		if items == nil {
			items = []*dto.AdPlacementWithAdsDTO{}
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) recordImpression() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.RecordImpression(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) recordCreativeImpression() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.RecordCreativeImpression(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) recordCreativeClick() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		creative, err := h.uc.GetCreative(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "creative not found")
			return nil
		}
		if err := h.uc.RecordCreativeClick(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"url": creative.LinkURL})
		return nil
	}
}

func (h *Handler) recordClick() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		ad, err := h.uc.GetAdByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "ad not found")
			return nil
		}
		clickLog := &dto.AdClickLogDTO{
			AdID: id, PlacementID: ad.PlacementID,
			IP: ctx.Request().RemoteAddr, UserAgent: ctx.GetHeader("User-Agent"), Referer: ctx.GetHeader("Referer"),
		}
		if userID, exists := ctx.Get("user_id"); exists {
			if uid, ok := userID.(string); ok {
				clickLog.UserID = uid
			}
		}
		if err := h.uc.RecordClick(ctx, id, clickLog); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"url": ad.LinkURL})
		return nil
	}
}

// ==================== Admin Creative Library Handlers (G6-3) ====================

func (h *Handler) listCreatives() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListCreatives(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) getCreative() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		item, err := h.uc.GetCreative(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "creative not found")
			return nil
		}
		http2.OK(ctx, item)
		return nil
	}
}

func (h *Handler) createCreative() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Title          string            `json:"title" binding:"required"`
			TitleI18n      map[string]string `json:"title_i18n"`
			ImageURL       string            `json:"image_url"`
			ImageMobileURL string            `json:"image_mobile_url"`
			LinkURL        string            `json:"link_url"`
			LinkTarget     string            `json:"link_target"`
			BadgeText      string            `json:"badge_text"`
			Priority       int               `json:"priority"`
			IsActive       *bool             `json:"is_active"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		linkTarget := "_blank"
		if input.LinkTarget != "" {
			linkTarget = input.LinkTarget
		}
		c := &dto.AdCreativeDTO{
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
		created, err := h.uc.CreateCreative(ctx, c)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updateCreative() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		var input struct {
			Title          string            `json:"title"`
			TitleI18n      map[string]string `json:"title_i18n"`
			ImageURL       string            `json:"image_url"`
			ImageMobileURL string            `json:"image_mobile_url"`
			LinkURL        string            `json:"link_url"`
			LinkTarget     string            `json:"link_target"`
			BadgeText      string            `json:"badge_text"`
			Priority       *int              `json:"priority"`
			IsActive       *bool             `json:"is_active"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		existing, err := h.uc.GetCreative(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "creative not found")
			return nil
		}
		c := &dto.AdCreativeDTO{ID: id}
		c.Title = input.Title
		if input.TitleI18n != nil {
			c.TitleI18n = input.TitleI18n
		} else {
			c.TitleI18n = existing.TitleI18n
		}
		c.ImageURL = input.ImageURL
		c.ImageMobileURL = input.ImageMobileURL
		c.LinkURL = input.LinkURL
		if input.LinkTarget != "" {
			c.LinkTarget = input.LinkTarget
		} else {
			c.LinkTarget = existing.LinkTarget
		}
		c.BadgeText = input.BadgeText
		if input.Priority != nil {
			c.Priority = *input.Priority
		} else {
			c.Priority = existing.Priority
		}
		if input.IsActive != nil {
			c.IsActive = *input.IsActive
		} else {
			c.IsActive = existing.IsActive
		}
		updated, err := h.uc.UpdateCreative(ctx, c)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deleteCreative() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeleteCreative(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listPlacementCreativeIDs() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placementID := ctx.Var("id")
		if placementID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "placement id is required")
			return nil
		}
		ids, err := h.uc.ListCreativeIDsByPlacement(ctx, placementID)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, ids)
		return nil
	}
}

func (h *Handler) assignCreative() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placementID := ctx.Var("id")
		if placementID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "placement id is required")
			return nil
		}
		var input struct {
			CreativeID string `json:"creative_id" binding:"required"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		if err := h.uc.AssignCreative(ctx, placementID, input.CreativeID); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) unassignCreative() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		placementID := ctx.Var("id")
		creativeID := ctx.Var("creativeId")
		if placementID == "" || creativeID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "placement id and creative id are required")
			return nil
		}
		if err := h.uc.UnassignCreative(ctx, placementID, creativeID); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}
