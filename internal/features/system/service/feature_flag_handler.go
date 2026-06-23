package service

import (
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type FeatureFlagHandler struct {
	featureUC *systembiz.FeatureFlagUseCase
}

func NewFeatureFlagHandler(featureUC *systembiz.FeatureFlagUseCase) *FeatureFlagHandler {
	return &FeatureFlagHandler{featureUC: featureUC}
}

func (h *FeatureFlagHandler) RegisterRoutes(r http2.Router) {
	features := r.Group("/features")
	{
		features.GET("", h.getFeatureFlags())
		features.PUT("", h.updateFeatureFlag())
	}
}

func (h *FeatureFlagHandler) getFeatureFlags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		flags := h.featureUC.GetAll(ctx.Request().Context())
		http2.OK(ctx, map[string]interface{}{"features": flags})
		return nil
	}
}

func (h *FeatureFlagHandler) updateFeatureFlag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		var req struct {
			Flag    string `json:"flag"`
			Enabled bool   `json:"enabled"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			http2.Fail(ctx, http2.ErrBadRequest, "invalid request body")
			return nil
		}

		if req.Flag == "" {
			http2.Fail(ctx, http2.ErrBadRequest, "flag is required")
			return nil
		}

		if err := h.featureUC.SetFlag(ctx.Request().Context(), req.Flag, req.Enabled); err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}

		flags := h.featureUC.GetAll(ctx.Request().Context())
		http2.OK(ctx, map[string]interface{}{"features": flags})
		return nil
	}
}
