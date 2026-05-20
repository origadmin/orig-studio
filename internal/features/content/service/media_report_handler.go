package service

import (
	"net/http"
	"strings"

	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
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
	r.POST("/medias/:id/report", server.WithJWTCtx(h.jwtMgr, server.HTTPToHandlerFunc(h.reportMedia())))
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

func (h *MediaReportHandler) reportMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		ctx := r.Context()

		id := gc.Param("id")
		if id == "" {
			server.Fail(gc, server.ErrBadRequest, "media ID is required")
			return
		}

		claims, ok := server.GetClaims(gc)
		if !ok {
			server.Fail(gc, server.ErrUnauthorized, "unauthorized")
			return
		}
		userID := claims.GetUserID()

		var req MediaReportRequest
		if err := gc.ShouldBindJSON(&req); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		if req.Reason == "" {
			server.Fail(gc, server.ErrBadRequest, "reason is required")
			return
		}

		reportCount, status, err := h.mediaReportUC.ReportMedia(ctx, id, userID, req.Reason, req.Description)
		if err != nil {
			if strings.Contains(err.Error(), "already reported") {
				server.Fail(gc, server.ErrConflict, err.Error())
				return
			}
			if strings.Contains(err.Error(), "cannot report your own") {
				server.Fail(gc, server.ErrBadRequest, err.Error())
				return
			}
			if strings.Contains(err.Error(), "failed to get media") {
				server.Fail(gc, server.ErrBadRequest, "media not found")
				return
			}
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, MediaReportResultDTO{
			Message:     "report submitted",
			ReportCount: reportCount,
			Status:      status,
		})
	}
}
