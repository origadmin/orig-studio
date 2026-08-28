package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"origadmin/application/origstudio/internal/conf"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	contentdal "origadmin/application/origstudio/internal/features/content/dal"
	"origadmin/application/origstudio/internal/data/entity"
	mediadal "origadmin/application/origstudio/internal/features/media/dal"
	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"

	"github.com/origadmin/runtime/log"
)

// subtitleLanguagesKey is the settings-table KV key holding the configurable
// subtitle language list (G5: 手动可加 + 列表可选 + 可配置).
const subtitleLanguagesKey = "subtitle_languages"

// defaultSubtitleLanguages is the built-in fallback list, used when the
// settings row is absent or unparsable.
var defaultSubtitleLanguages = []map[string]string{
	{"code": "zh", "label": "中文"},
	{"code": "en", "label": "English"},
	{"code": "ja", "label": "日本語"},
	{"code": "ko", "label": "한국어"},
	{"code": "fr", "label": "Français"},
	{"code": "de", "label": "Deutsch"},
	{"code": "es", "label": "Español"},
	{"code": "ru", "label": "Русский"},
}

// SubtitleHandler — BUG-186: real subtitle endpoints replacing the stubs.
// G5: owner + admin can upload/delete; public users only consume existing tracks.
type SubtitleHandler struct {
	jwt     *auth.Manager
	data    *contentdal.Data
	repo    contentdal.SubtitleRepo
	cfg     contentbiz.SystemConfigRepo
	storage mediadal.Storage
	paths   *conf.StoragePaths
	logger  *log.Helper
}

func NewSubtitleHandler(jwt *auth.Manager, data *contentdal.Data, paths *conf.StoragePaths, logger log.Logger) *SubtitleHandler {
	return &SubtitleHandler{
		jwt:     jwt,
		data:    data,
		repo:    contentdal.NewSubtitleRepo(data),
		cfg:     contentdal.NewSystemConfigRepo(data, logger),
		storage: mediadal.NewLocalStorage(paths),
		paths:   paths,
		logger:  log.NewHelper(log.With(logger, "module", "service.subtitle")),
	}
}

// RegisterRoutes registers the subtitle endpoints (BUG-186).
// Registered BEFORE gRPC-Gateway so the UUID-aware /media/:token paths win.
func (h *SubtitleHandler) RegisterRoutes(r http2.Router) {
	medias := r.Group("/medias")
	medias.GET("/:token/subtitles", h.handleSubtitleList)
	medias.POST("/:token/subtitles", server.WithJWTCtx(h.jwt, h.handleSubtitleCreate))

	subtitles := r.Group("/subtitles")
	subtitles.DELETE("/:id", server.WithJWTCtx(h.jwt, h.handleSubtitleDelete))
	subtitles.GET("/languages", h.handleSubtitleLanguages)

	// Admin language configuration (G5: 可配置). Both admin-only.
	adminSubs := r.Group("/admin/subtitle-languages")
	adminSubs.GET("", server.WithJWTCtx(h.jwt, h.handleAdminLanguagesGet))
	adminSubs.POST("", server.WithJWTCtx(h.jwt, h.handleAdminLanguagesSet))
}

// languages returns the configurable language list (G5: 手动可加 + 列表可选).
// Reads the settings-table KV `subtitle_languages` (JSON [{code,label},...]);
// falls back to the built-in default list when absent/unparsable.
func (h *SubtitleHandler) languages(ctx context.Context) []map[string]string {
	raw, err := h.cfg.Get(ctx, subtitleLanguagesKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return defaultSubtitleLanguages
	}
	var list []map[string]string
	if err := json.Unmarshal([]byte(raw), &list); err != nil || len(list) == 0 {
		return defaultSubtitleLanguages
	}
	return list
}

func (h *SubtitleHandler) handleSubtitleLanguages(ctx http2.Context) error {
	return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": h.languages(ctx.Request().Context())})
}

// requireAdmin returns true when the JWT claims carry admin role.
func (h *SubtitleHandler) requireAdmin(ctx http2.Context) bool {
	claims, ok := server.GetClaimsCtx(ctx)
	return ok && claims != nil && claims.IsAdmin()
}

// handleAdminLanguagesGet: GET /admin/subtitle-languages — current list.
func (h *SubtitleHandler) handleAdminLanguagesGet(ctx http2.Context) error {
	if !h.requireAdmin(ctx) {
		return server.FailCtx(ctx, 403, "PERMISSION_DENIED: admin only")
	}
	return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": h.languages(ctx.Request().Context())})
}

// handleAdminLanguagesSet: POST /admin/subtitle-languages — full replace.
// Body: {"languages":[{"code":"zh","label":"中文"},...]} (admin only).
func (h *SubtitleHandler) handleAdminLanguagesSet(ctx http2.Context) error {
	if !h.requireAdmin(ctx) {
		return server.FailCtx(ctx, 403, "PERMISSION_DENIED: admin only")
	}
	var req struct {
		Languages []map[string]string `json:"languages"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return server.FailCtx(ctx, 400, "BAD_REQUEST: "+err.Error())
	}
	if len(req.Languages) == 0 {
		return server.FailCtx(ctx, 400, "BAD_REQUEST: languages must not be empty")
	}
	// Validate: code + label required, code unique.
	seen := make(map[string]bool, len(req.Languages))
	for _, l := range req.Languages {
		code := strings.TrimSpace(l["code"])
		label := strings.TrimSpace(l["label"])
		if code == "" || label == "" {
			return server.FailCtx(ctx, 400, "BAD_REQUEST: each language needs non-empty code and label")
		}
		if seen[code] {
			return server.FailCtx(ctx, 400, "BAD_REQUEST: duplicate language code "+code)
		}
		seen[code] = true
	}
	b, _ := json.Marshal(req.Languages)
	if err := h.cfg.Set(ctx.Request().Context(), subtitleLanguagesKey, string(b)); err != nil {
		return server.FailCtx(ctx, 500, "LANGUAGE_SAVE_FAILED: "+err.Error())
	}
	return server.OKCtx(ctx, map[string]any{"code": 0, "message": "saved", "data": req.Languages})
}

// handleSubtitleList: GET /medias/:token/subtitles — list tracks for a media.
func (h *SubtitleHandler) handleSubtitleList(ctx http2.Context) error {
	media, err := h.mediaByToken(ctx, ctx.Var("token"))
	if err != nil {
		return server.FailCtx(ctx, 404, "MEDIA_NOT_FOUND: media not found")
	}
	items, err := h.repo.ListByMediaID(ctx.Request().Context(), media.ID)
	if err != nil {
		return server.FailCtx(ctx, 500, "SUBTITLE_LIST_FAILED: "+err.Error())
	}
	return server.OKCtx(ctx, map[string]any{"code": 0, "message": "ok", "data": items})
}

// handleSubtitleCreate: POST /medias/:token/subtitles (multipart file + language).
// Owner or admin only. Converts srt/vtt -> vtt, stores file, writes row with
// status=active; on conversion failure writes status=failed + error_message.
func (h *SubtitleHandler) handleSubtitleCreate(ctx http2.Context) error {
	media, err := h.mediaByToken(ctx, ctx.Var("token"))
	if err != nil {
		return server.FailCtx(ctx, 404, "MEDIA_NOT_FOUND: media not found")
	}
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok || claims == nil || !h.canManage(claims, media) {
		return server.FailCtx(ctx, 403, "PERMISSION_DENIED: only the media owner or an admin can manage subtitles")
	}

	file, header, err := ctx.FormFile("file")
	if err != nil {
		return server.FailCtx(ctx, 400, "BAD_REQUEST: missing 'file' field")
	}
	defer file.Close()

	language := ctx.FormVar("language")
	if language == "" {
		return server.FailCtx(ctx, 400, "BAD_REQUEST: missing 'language' field")
	}
	if !h.languageSupported(ctx.Request().Context(), language) {
		return server.FailCtx(ctx, 400, "BAD_REQUEST: unsupported language '"+language+"'")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	isVTT := ext == ".vtt"
	if ext != ".srt" && !isVTT {
		msg := "仅支持 .srt / .vtt，收到 " + ext
		_ = h.recordFailed(ctx, media.ID, language, msg)
		return server.FailCtx(ctx, 400, "UNSUPPORTED_FORMAT: "+msg)
	}

	content, rerr := io.ReadAll(io.LimitReader(file, 5<<20+1))
	if rerr != nil {
		return server.FailCtx(ctx, 400, "BAD_REQUEST: failed to read file")
	}
	if len(content) > 5<<20 {
		msg := "文件超过 5MB 限制"
		_ = h.recordFailed(ctx, media.ID, language, msg)
		return server.FailCtx(ctx, 400, "FILE_TOO_LARGE: "+msg)
	}

	normalized, convErr := contentbiz.NormalizeSubtitle(content, isVTT)
	if convErr != nil {
		// G5: report line-level failure so the user knows what to fix.
		_ = h.recordFailed(ctx, media.ID, language, convErr.Error())
		return server.FailCtx(ctx, 400, "SUBTITLE_PARSE_FAILED: "+convErr.Error())
	}

	path := fmt.Sprintf("subtitles/%s/%s.vtt", media.ID, language)
	if err := h.storage.PutFile(ctx.Request().Context(), path, normalized); err != nil {
		return server.FailCtx(ctx, 500, "STORAGE_FAILED: "+err.Error())
	}

	item, err := h.repo.Create(ctx.Request().Context(), &contentdal.SubtitleItem{
		MediaID:  media.ID,
		Language: language,
		Label:    languageLabel(h.languages(ctx.Request().Context()), language),
		FileURL:  path,
		Status:   "active",
	})
	if err != nil {
		return server.FailCtx(ctx, 500, "SUBTITLE_CREATE_FAILED: "+err.Error())
	}
	return server.OKCtx(ctx, map[string]any{"code": 0, "message": "created", "data": item})
}

// handleSubtitleDelete: DELETE /subtitles/:id — owner or admin only.
func (h *SubtitleHandler) handleSubtitleDelete(ctx http2.Context) error {
	id := ctx.Var("id")
	item, err := h.repo.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return server.FailCtx(ctx, 404, "SUBTITLE_NOT_FOUND: subtitle not found")
	}
	media, err := h.data.MediaByID(ctx.Request().Context(), item.MediaID)
	if err != nil {
		return server.FailCtx(ctx, 404, "MEDIA_NOT_FOUND: media not found")
	}
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok || claims == nil || !h.canManage(claims, media) {
		return server.FailCtx(ctx, 403, "PERMISSION_DENIED: only the media owner or an admin can delete subtitles")
	}

	if err := h.repo.Delete(ctx.Request().Context(), id); err != nil {
		return server.FailCtx(ctx, 500, "SUBTITLE_DELETE_FAILED: "+err.Error())
	}
	if item.FileURL != "" {
		_ = h.storage.DeleteFile(ctx.Request().Context(), item.FileURL)
	}
	return server.OKCtx(ctx, map[string]any{"code": 0, "message": "deleted"})
}

// mediaByToken resolves a media row by short_token (id + owner checks).
func (h *SubtitleHandler) mediaByToken(ctx http2.Context, token string) (*entity.Media, error) {
	return h.data.MediaByToken(ctx.Request().Context(), token)
}

func (h *SubtitleHandler) canManage(claims *auth.Claims, media *entity.Media) bool {
	if claims.IsAdmin() {
		return true
	}
	return media.UserID != "" && claims.GetUserID() == media.UserID
}

func (h *SubtitleHandler) languageSupported(ctx context.Context, code string) bool {
	for _, l := range h.languages(ctx) {
		if l["code"] == code {
			return true
		}
	}
	return false
}

func (h *SubtitleHandler) recordFailed(ctx http2.Context, mediaID, language, msg string) error {
	_, err := h.repo.Create(ctx.Request().Context(), &contentdal.SubtitleItem{
		MediaID:      mediaID,
		Language:     language,
		Status:       "failed",
		ErrorMessage: msg,
	})
	return err
}

func languageLabel(list []map[string]string, code string) string {
	for _, l := range list {
		if l["code"] == code {
			return l["label"]
		}
	}
	return code
}
