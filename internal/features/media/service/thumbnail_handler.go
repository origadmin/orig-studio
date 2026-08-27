package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"origadmin/application/origstudio/internal/server"
)

// ---------------------------------------------------------------------------
// Thumbnail management endpoints (gin-free, net/http)
//
// These replace the gin SpriteHandler that only existed in the CE monolith
// (cmd/server/main.go) and never ran in the EE microservice. They are mounted
// on the EE media HTTP server at distinct paths (NOT the proto
// `regenerate-thumbnail` literal) because Kratos/httprouter panics on
// duplicate route registration - re-registering the same path as the
// proto-generated gRPC-Gateway route would crash the server at startup.
//
// All handlers run FFmpeg / storage ops under a background-derived context so
// the work is not cancelled when the HTTP request context is cancelled.
// ---------------------------------------------------------------------------

const maxThumbnailUploadSize = 10 << 20 // 10 MB

type regenThumbnailRequest struct {
	ThumbnailTime  float64 `json:"thumbnail_time"`
	UseSpriteSheet bool    `json:"use_sprite_sheet"`
}

type thumbnailResponse struct {
	Thumbnail     string  `json:"thumbnail"`
	ThumbnailTime float64 `json:"thumbnail_time"`
	Message       string  `json:"message"`
	Success       bool    `json:"success"`
}

type errorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeThumbnailJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeThumbnailError(w http.ResponseWriter, status int, msg string) {
	writeThumbnailJSON(w, status, errorBody{Error: http.StatusText(status), Message: msg})
}

// seg extracts the path segment at idx after trimming leading/trailing slashes.
// e.g. /api/v1/admin/medias/{id}/regen-thumbnail -> seg(r, 4) == "{id}".
func seg(r *http.Request, idx int) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if idx < 0 || idx >= len(parts) {
		return ""
	}
	return parts[idx]
}

// uploadExt maps an uploaded file's name / content-type to a file extension
// that reflects its real format, so downstream FFmpeg decoding picks the
// correct demuxer. Defaulting to ".jpg" broke PNG/WebP uploads (HTTP 500).
func uploadExt(filename, contentType string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	}
	// Fall back to the filename extension when the content-type is missing
	// or unrecognised.
	if i := strings.LastIndex(filename, "."); i >= 0 {
		ext := strings.ToLower(filename[i:])
		if ext == ".png" || ext == ".webp" || ext == ".jpg" || ext == ".jpeg" {
			return ext
		}
	}
	return ".jpg"
}

// saveUploadedFile reads the multipart "file" field into a temp file and
// returns its path plus a cleanup func. The caller must invoke cleanup.
func saveUploadedFile(r *http.Request) (string, func(), error) {
	empty := func() {}
	if err := r.ParseMultipartForm(maxThumbnailUploadSize); err != nil {
		return "", empty, fmt.Errorf("invalid multipart form: %w", err)
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return "", empty, fmt.Errorf("missing 'file' field: %w", err)
	}
	defer file.Close()
	if header.Size > maxThumbnailUploadSize {
		return "", empty, fmt.Errorf("file too large (max 10MB)")
	}
	// Derive a temp-file extension that matches the ACTUAL uploaded content
	// type. Naming it ".jpg" unconditionally made FFmpeg pick the MJPEG
	// demuxer for PNG/WebP uploads, which then failed to decode the file and
	// returned HTTP 500 ("chunk too big" / "Invalid data found"). The
	// extension must reflect the real format so FFmpeg auto-detects correctly.
	ext := uploadExt(header.Filename, header.Header.Get("Content-Type"))
	tmp, err := os.CreateTemp("", "thumb-*"+ext)
	if err != nil {
		return "", empty, fmt.Errorf("cannot create temp file: %w", err)
	}
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", empty, fmt.Errorf("cannot save uploaded file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", empty, fmt.Errorf("cannot close temp file: %w", err)
	}
	return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
}

// resolveThumbnailURL re-fetches the media to return its (fresh) thumbnail path.
func (s *AdminMediaService) resolveThumbnailURL(bgCtx context.Context, id string) string {
	m, err := s.uc.GetMedia(bgCtx, id)
	if err != nil || m == nil {
		return ""
	}
	return m.Thumbnail
}

// AdminRegenThumbnailHTTPHandler triggers thumbnail regeneration at a chosen
// timestamp for an admin. POST /api/v1/admin/medias/{id}/regen-thumbnail
func (s *AdminMediaService) AdminRegenThumbnailHTTPHandler(w http.ResponseWriter, r *http.Request) {
	id := seg(r, 4)
	if id == "" {
		writeThumbnailError(w, http.StatusBadRequest, "missing media id")
		return
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()

	if _, err := s.uc.GetMedia(bgCtx, id); err != nil {
		writeThumbnailError(w, http.StatusNotFound, "media not found")
		return
	}

	var body regenThumbnailRequest
	if r.Body != nil && r.ContentLength != 0 {
		// Best-effort parse; missing body defaults thumbnail_time to 0 (backend
		// then uses the configured default position).
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	if body.UseSpriteSheet {
		// Cover = whole sprite sheet image (not a single sampled frame).
		if err := s.uc.SetSpriteSheetThumbnail(bgCtx, id); err != nil {
			writeThumbnailError(w, http.StatusInternalServerError, "failed to set sprite-sheet cover: "+err.Error())
			return
		}
	} else if err := s.uc.RegenerateThumbnail(bgCtx, id, body.ThumbnailTime); err != nil {
		writeThumbnailError(w, http.StatusInternalServerError, "failed to regenerate thumbnail: "+err.Error())
		return
	}

	writeThumbnailJSON(w, http.StatusOK, thumbnailResponse{
		Thumbnail:     s.resolveThumbnailURL(bgCtx, id),
		ThumbnailTime: body.ThumbnailTime,
		Message:       "thumbnail regenerated",
		Success:       true,
	})
}

// AdminUploadThumbnailHTTPHandler sets a custom cover image for an admin.
// POST /api/v1/admin/medias/{id}/set-thumbnail  (multipart/form-data, field "file")
func (s *AdminMediaService) AdminUploadThumbnailHTTPHandler(w http.ResponseWriter, r *http.Request) {
	id := seg(r, 4)
	if id == "" {
		writeThumbnailError(w, http.StatusBadRequest, "missing media id")
		return
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()

	if _, err := s.uc.GetMedia(bgCtx, id); err != nil {
		writeThumbnailError(w, http.StatusNotFound, "media not found")
		return
	}

	localPath, cleanup, err := saveUploadedFile(r)
	if err != nil {
		writeThumbnailError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cleanup()

	if err := s.uc.SetCustomThumbnail(bgCtx, id, localPath); err != nil {
		writeThumbnailError(w, http.StatusInternalServerError, "failed to set custom thumbnail: "+err.Error())
		return
	}

	writeThumbnailJSON(w, http.StatusOK, thumbnailResponse{
		Thumbnail: s.resolveThumbnailURL(bgCtx, id),
		Message:   "custom thumbnail set",
		Success:   true,
	})
}

// OwnerRegenThumbnailHTTPHandler triggers thumbnail regeneration for the owner
// of a media item identified by its short token.
// POST /api/v1/me/medias/{token}/regen-thumbnail
func (s *AdminMediaService) OwnerRegenThumbnailHTTPHandler(w http.ResponseWriter, r *http.Request) {
	token := seg(r, 4)
	if token == "" {
		writeThumbnailError(w, http.StatusBadRequest, "missing media token")
		return
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()

	m, err := s.uc.GetByShortToken(bgCtx, token)
	if err != nil || m == nil {
		writeThumbnailError(w, http.StatusNotFound, "media not found")
		return
	}
	if !ownerAuthorized(r, m.UserId) {
		writeThumbnailError(w, http.StatusForbidden, "not allowed to modify this media")
		return
	}
	id := m.Id

	var body regenThumbnailRequest
	if r.Body != nil && r.ContentLength != 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	if body.UseSpriteSheet {
		// Cover = whole sprite sheet image (not a single sampled frame).
		if err := s.uc.SetSpriteSheetThumbnail(bgCtx, id); err != nil {
			writeThumbnailError(w, http.StatusInternalServerError, "failed to set sprite-sheet cover: "+err.Error())
			return
		}
	} else if err := s.uc.RegenerateThumbnail(bgCtx, id, body.ThumbnailTime); err != nil {
		writeThumbnailError(w, http.StatusInternalServerError, "failed to regenerate thumbnail: "+err.Error())
		return
	}

	writeThumbnailJSON(w, http.StatusOK, thumbnailResponse{
		Thumbnail:     s.resolveThumbnailURL(bgCtx, id),
		ThumbnailTime: body.ThumbnailTime,
		Message:       "thumbnail regenerated",
		Success:       true,
	})
}

// AdminGetThumbnailHTTPHandler returns the current cover thumbnail for an
// admin. GET /api/v1/admin/medias/{id}/thumbnail
//
// The non-admin equivalent (GetMediaThumbnail) is registered by the proto
// gRPC-Gateway at /api/v1/medias/{id}/thumbnail. This admin-variant exists so
// the admin UI and tooling can fetch the cover under the /admin/ path without
// 404. It returns the stored relative thumbnail path; the caller applies its
// own getFullUrl() prefix.
func (s *AdminMediaService) AdminGetThumbnailHTTPHandler(w http.ResponseWriter, r *http.Request) {
	id := seg(r, 4)
	if id == "" {
		writeThumbnailError(w, http.StatusBadRequest, "missing media id")
		return
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m, err := s.uc.GetMedia(bgCtx, id)
	if err != nil || m == nil {
		writeThumbnailError(w, http.StatusNotFound, "media not found")
		return
	}

	writeThumbnailJSON(w, http.StatusOK, thumbnailResponse{
		Thumbnail: m.Thumbnail,
		Message:   "ok",
		Success:   true,
	})
}

// OwnerUploadThumbnailHTTPHandler sets a custom cover image for the owner of a
// media item identified by its short token.
// POST /api/v1/me/medias/{token}/set-thumbnail  (multipart/form-data, field "file")
func (s *AdminMediaService) OwnerUploadThumbnailHTTPHandler(w http.ResponseWriter, r *http.Request) {
	token := seg(r, 4)
	if token == "" {
		writeThumbnailError(w, http.StatusBadRequest, "missing media token")
		return
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()

	m, err := s.uc.GetByShortToken(bgCtx, token)
	if err != nil || m == nil {
		writeThumbnailError(w, http.StatusNotFound, "media not found")
		return
	}
	if !ownerAuthorized(r, m.UserId) {
		writeThumbnailError(w, http.StatusForbidden, "not allowed to modify this media")
		return
	}
	id := m.Id

	localPath, cleanup, err := saveUploadedFile(r)
	if err != nil {
		writeThumbnailError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cleanup()

	if err := s.uc.SetCustomThumbnail(bgCtx, id, localPath); err != nil {
		writeThumbnailError(w, http.StatusInternalServerError, "failed to set custom thumbnail: "+err.Error())
		return
	}

	writeThumbnailJSON(w, http.StatusOK, thumbnailResponse{
		Thumbnail: s.resolveThumbnailURL(bgCtx, id),
		Message:   "custom thumbnail set",
		Success:   true,
	})
}

// ownerAuthorized verifies the authenticated caller owns the media. The JWT
// subject is compared against the media's owner id. When either side is
// unknown we fail open to forbidden only if the media has a known owner but
// the caller is unauthenticated.
func ownerAuthorized(r *http.Request, mediaUserID string) bool {
	claims, ok := server.GetClaimsFromStdCtx(r.Context())
	if !ok {
		return false
	}
	if mediaUserID == "" {
		// No owner recorded - allow any authenticated caller (defensive).
		return true
	}
	return claims.GetUserID() == mediaUserID
}
