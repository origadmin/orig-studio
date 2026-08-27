package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/enterprise/user/tenant/biz"
	"origadmin/application/origstudio/internal/enterprise/user/tenant/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

type Handler struct {
	tenantUC *biz.UseCase
	jwtMgr   *auth.Manager
}

func NewHandler(tenantUC *biz.UseCase, jwtMgr *auth.Manager) *Handler {
	return &Handler{tenantUC: tenantUC, jwtMgr: jwtMgr}
}

// RegisterRoutes registers tenant routes with JWT + Admin middleware protection.
// Security fix: previously these admin routes had NO authentication.
func (h *Handler) RegisterRoutes(r http2.Router) {
	tenants := r.Group("/admin/tenants")
	tenants.Use(server.JWTMiddlewareCtx(h.jwtMgr), server.AdminMiddlewareCtx(h.jwtMgr))
	{
		tenants.GET("", server.HTTPToHandlerFunc(h.list))
		tenants.POST("", server.HTTPToHandlerFunc(h.create))
		tenants.GET("/:id", server.HTTPToHandlerFunc(h.getByID))
		tenants.PUT("/:id", server.HTTPToHandlerFunc(h.update))
		tenants.DELETE("/:id", server.HTTPToHandlerFunc(h.delete))
	}

	current := r.Group("/tenant")
	current.Use(server.JWTMiddlewareCtx(h.jwtMgr))
	{
		current.GET("/current", server.HTTPToHandlerFunc(h.current))
	}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var t dto.TenantDTO
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeTenantJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.tenantUC.Create(r.Context(), &t)
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusCreated, map[string]interface{}{
		"tenant": result,
	})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeTenantJSONError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	result, err := h.tenantUC.Get(r.Context(), id)
	if err != nil {
		writeTenantJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenant": result,
	})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeTenantJSONError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	var t dto.TenantDTO
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeTenantJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.tenantUC.Update(r.Context(), id, &t)
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenant": result,
	})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) current(w http.ResponseWriter, r *http.Request) {
	result, err := h.tenantUC.ResolveFromContext(r.Context())
	if err != nil {
		writeTenantJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result == nil {
		writeTenantJSON(w, http.StatusOK, map[string]interface{}{
			"tenant":    nil,
			"is_system": true,
		})
		return
	}

	writeTenantJSON(w, http.StatusOK, map[string]interface{}{
		"tenant":    result,
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