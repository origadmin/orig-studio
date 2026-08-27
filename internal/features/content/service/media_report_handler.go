package service

import (
	"strings"

	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/server"
)

type MediaReportHandler struct {
	mediaReportUC *contentbiz.MediaReportUseCase
	jwtMgr        *auth.Manager
}

func NewMediaReportHandler(mediaReportUC *contentbiz.MediaReportUseCase, jwtMgr *auth.Manager) *MediaReportHandler {
	return &MediaReportHandler{
		mediaReportUC: mediaReportUC,
		jwtMgr:        jwtMgr,
	}
}

func (h *MediaReportHandler) RegisterRoutes(r http2.Router) {
	r.POST("/medias/:token/report", server.WithJWTCtx(h.jwtMgr, h.reportMedia()))
}

type MediaReportRequest struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

type MediaReportResultDTO struct {
	Message     string `json:"message"`
	ReportCount int    `json:"report_count"`
	Status      string `json:"status"`
}

func (h *MediaReportHandler) reportMedia() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("token")
		if id == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "media ID is required")
		}

		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
		}
		userID := claims.GetUserID()

		var req MediaReportRequest
		if err := ctx.BindJSON(&req); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		if req.Reason == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "reason is required")
		}

		reportCount, status, err := h.mediaReportUC.ReportMedia(ctx, id, userID, req.Reason, req.Description)
		if err != nil {
			if strings.Contains(err.Error(), "already reported") {
				return server.FailCtx(ctx, server.ErrConflict, err.Error())
			}
			if strings.Contains(err.Error(), "cannot report your own") {
				return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			}
			if strings.Contains(err.Error(), "failed to get media") {
				return server.FailCtx(ctx, server.ErrBadRequest, "media not found")
			}
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, MediaReportResultDTO{
			Message:     "report submitted",
			ReportCount: reportCount,
			Status:      status,
		})
	}
}
