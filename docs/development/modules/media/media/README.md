# Media Module

> Media base module: upload, list, detail, playback.

## Module Owner

- Owner: **TBD**
- Status: **Partially Complete**

## Documents

- [DESIGN.md](./DESIGN.md) — Media base system design
- [TESTING.md](./TESTING.md) — Acceptance criteria and test plan

## Current Implementation

| Feature | Status | Notes |
|---------|--------|-------|
| Media CRUD | ✅ Done | `internal/server/media.go` + `svc-media/biz/media.go` |
| File Upload | ✅ Done | Chunked upload via `svc-media/biz/upload.go` |
| Thumbnail | ✅ Done | ffmpeg frame extraction |
| List & Detail API | ✅ Done | Paginated list, static file serving |
| Frontend Upload | ✅ Done | `components/MediaUpload.tsx` |
| Watch Page | ⚠️ Partial | Bug: i18n keys not resolving (`watch.subscribe`) |

---

*Last updated: 2026-04-13*
