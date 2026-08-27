package service

import (
	"strconv"
	"time"


	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/enterprise/media/drm/biz"
	"origadmin/application/origstudio/internal/enterprise/media/drm/dto"
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
	adminPolicies := r.Group("/admin/drm-policies")
	adminPolicies.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminPolicies.GET("", h.listPolicies())
		adminPolicies.POST("", h.createPolicy())
		adminPolicies.PUT("/:id", h.updatePolicy())
		adminPolicies.DELETE("/:id", h.deletePolicy())
		adminPolicies.GET("/:id/keys", h.listKeys())
		adminPolicies.POST("/:id/keys", h.generateKey())
	}

	adminKeys := r.Group("/admin/drm-keys")
	adminKeys.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminKeys.DELETE("/:id", h.deleteKey())
	}

	adminLicenses := r.Group("/admin/drm-licenses")
	adminLicenses.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminLicenses.GET("", h.listLicenses())
	}
}

func (h *Handler) listPolicies() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListPolicies(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) createPolicy() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Name            string `json:"name" binding:"required"`
			Type            string `json:"type" binding:"required"`
			HlsKeyURL       string `json:"hls_key_url"`
			WidevinePssh    string `json:"widevine_pssh"`
			FairplayCertURL string `json:"fairplay_cert_url"`
			IsDefault       *bool  `json:"is_default"`
			Description     string `json:"description"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		isDefault := false
		if input.IsDefault != nil {
			isDefault = *input.IsDefault
		}
		p := &dto.DrmPolicyDTO{
			Name:            input.Name,
			Type:            input.Type,
			HlsKeyURL:       input.HlsKeyURL,
			WidevinePssh:    input.WidevinePssh,
			FairplayCertURL: input.FairplayCertURL,
			IsDefault:       isDefault,
			Description:     input.Description,
		}
		created, err := h.uc.CreatePolicy(ctx, p)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updatePolicy() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		var input struct {
			Name            string `json:"name"`
			Type            string `json:"type"`
			HlsKeyURL       string `json:"hls_key_url"`
			WidevinePssh    string `json:"widevine_pssh"`
			FairplayCertURL string `json:"fairplay_cert_url"`
			IsDefault       *bool  `json:"is_default"`
			Description     string `json:"description"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		existing, err := h.uc.GetPolicyByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "policy not found")
			return nil
		}
		p := &dto.DrmPolicyDTO{ID: id}
		if input.Name != "" {
			p.Name = input.Name
		} else {
			p.Name = existing.Name
		}
		if input.Type != "" {
			p.Type = input.Type
		} else {
			p.Type = existing.Type
		}
		p.HlsKeyURL = input.HlsKeyURL
		p.WidevinePssh = input.WidevinePssh
		p.FairplayCertURL = input.FairplayCertURL
		p.Description = input.Description
		if input.IsDefault != nil {
			p.IsDefault = *input.IsDefault
		} else {
			p.IsDefault = existing.IsDefault
		}
		updated, err := h.uc.UpdatePolicy(ctx, p)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deletePolicy() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeletePolicy(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listKeys() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		policyID := ctx.Var("id")
		if policyID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "policy id is required")
			return nil
		}
		items, err := h.uc.ListKeysByPolicy(ctx, policyID)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) generateKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		policyID := ctx.Var("id")
		if policyID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "policy id is required")
			return nil
		}
		var input struct {
			ContentID string `json:"content_id" binding:"required"`
			ExpiresAt string `json:"expires_at"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		expiresAt := time.Now().Add(365 * 24 * time.Hour)
		if input.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339, input.ExpiresAt); err == nil {
				expiresAt = t
			}
		}
		created, err := h.uc.GenerateKey(ctx, policyID, input.ContentID, expiresAt)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) deleteKey() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeleteKey(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listLicenses() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}
		items, total, err := h.uc.ListLicenses(ctx, page, pageSize)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total, "page": page, "page_size": pageSize})
		return nil
	}
}