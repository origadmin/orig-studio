package service

import (
	"encoding/json"
	"net/http"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	"origadmin/application/origstudio/internal/server"
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
		features.GET("", server.HTTPToHandlerFunc(h.getFeatureFlags()))
		features.PUT("", server.HTTPToHandlerFunc(h.updateFeatureFlag()))
	}
}

func (h *FeatureFlagHandler) getFeatureFlags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flags := h.featureUC.GetAll(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"features": flags,
		})
	}
}

func (h *FeatureFlagHandler) updateFeatureFlag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Flag    string `json:"flag"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		if req.Flag == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "flag is required"})
			return
		}

		if err := h.featureUC.SetFlag(r.Context(), req.Flag, req.Enabled); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		flags := h.featureUC.GetAll(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"features": flags,
		})
	}
}
