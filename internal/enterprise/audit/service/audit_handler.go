package service

import (
	"strconv"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/enterprise/audit/dal"
	"origadmin/application/origstudio/internal/enterprise/audit/dto"
)

type Handler struct {
	svc *dal.AuditService
	jwt *auth.Manager
}

func NewHandler(svc *dal.AuditService, jwt *auth.Manager) *Handler {
	return &Handler{svc: svc, jwt: jwt}
}

func (h *Handler) RegisterRoutes(r http2.Router) {
	adminAudit := r.Group("/admin/audit-logs")
	adminAudit.Use(server.JWTMiddlewareCtx(h.jwt), server.AdminMiddlewareCtx(h.jwt))
	{
		adminAudit.GET("", h.ListAuditLogs())
	}
}

func (h *Handler) ListAuditLogs() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}

		filter := &dto.QueryFilter{
			UserID:    ctx.QueryVar("user_id"),
			Action:    ctx.QueryVar("action"),
			Resource:  ctx.QueryVar("resource"),
			Result:    ctx.QueryVar("result"),
			StartTime: ctx.QueryVar("start_time"),
			EndTime:   ctx.QueryVar("end_time"),
			Page:      page,
			PageSize:  pageSize,
		}

		result, err := h.svc.Query(ctx, filter)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, result)
		return nil
	}
}