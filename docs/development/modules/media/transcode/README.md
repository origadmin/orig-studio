# Media Transcode Module

> Video transcoding: HLS generation, encoding profiles, progress tracking.

## Module Owner

- Owner: **TBD**
- Status: **Complete**

## Documents

- [DESIGN.md](./DESIGN.md) — Transcoding architecture (see project docs)
- [TESTING.md](./TESTING.md) — Acceptance criteria

## Current Implementation

| Feature | Status | Notes |
|---------|--------|-------|
| FFmpeg wrapper | ✅ Done | `internal/helpers/ffmpeg/` |
| TranscodeWorker | ✅ Done | `svc-media/biz/transcode_worker.go` |
| TranscodeHandler | ✅ Done | `svc-media/biz/transcode_handler.go` |
| HLS generation | ✅ Done | Bento4 master playlist |
| SSE progress | ✅ Done | `server/media.go:transcodingEvents()` |
| Frontend HLS.js | ✅ Done | `pages/home/Watch.tsx` |
| Quality switcher | ✅ Done | Settings button + dropdown |
| Retry button | ✅ Done | On failed encoding |

---

*Last updated: 2026-04-13*
