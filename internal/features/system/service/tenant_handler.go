package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"
)

type TenantHandler struct {
	tenantUC *systembiz.TenantUseCase
}

func NewTenantHandler(tenantUC *systembiz.TenantUseCase) *TenantHandler {
	return &TenantHandler{tenantUC: tenantUC}
}

func (h *TenantHandler) RegisterRoutes(r http2.Router) {
	tenants := r.Group("/admin/tenants")
	{
		tenants.GET("", server.HTTPToHandlerFunc(h.list))
		tenants.POST("", server.HTTPToHandlerFunc(h.create))
		tenants.GET("/:id", server.HTTPToHandlerFunc(h.getByID))
		tenants.PUT("/:id", server.HTTPToHandlerFunc(h.update))
		tenants.DELETE("/:id", server.HTTPToHandlerFunc(h.delete))
	}

	current := r.Group("/tenant")
	{
		current.GET("/current", server.HTTPToHandlerFunc(h.current))
	}
}

func (h *TenantHandler) list(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if pageSize <= 0 {
		pageSize = 20
	}

	tenants, total, err := h.tenantUC.List(r.Context(), int(page), int(pageSize))
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenants":   tenants,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *TenantHandler) create(w http.ResponseWriter, r *http.Request) {
	var dto systembiz.TenantDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeTenantJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.tenantUC.Create(r.Context(), &dto)
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusCreated, map[string]interface{}{
		"tenant": result,
	})
}

func (h *TenantHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeTenantJSONError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	dto, err := h.tenantUC.Get(r.Context(), id)
	if err != nil {
		writeTenantJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenant": dto,
	})
}

func (h *TenantHandler) update(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeTenantJSONError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	var dto systembiz.TenantDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeTenantJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.tenantUC.Update(r.Context(), id, &dto)
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenant": result,
	})
}

func (h *TenantHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeTenantJSONError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	if err := h.tenantUC.Delete(r.Context(), id); err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": true,
	})
}

func (h *TenantHandler) current(w http.ResponseWriter, r *http.Request) {
	dto, err := h.tenantUC.ResolveFromContext(r.Context())
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if dto == nil {
		writeTenantJSON(w, http.StatusOK, map[string]interface{}{
			"tenant":    nil,
			"is_system": true,
		})
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenant":    dto,
		"is_system": false,
	})
}

func writeTenantJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeTenantJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
