# Implementation Plan M3: Media Processing & Adaptive Streaming

> See [MILESTONES.md](./MILESTONES.md#m3-视频转码与-hls) for the full task breakdown.

## Goal

Transform the system into a professional video platform by implementing automatic HLS transcoding (adaptive streaming) and enhancing the playback experience with status monitoring.

## Approach

> **HLS Transcoding** is resource-intensive. For now, it will be executed as a background goroutine on the same server using standard transcoding parameters via FFmpeg.

## Backend Changes

### HLS Adaptive Streaming

- Create `internal/helpers/ffmpeg/hls.go`:
  - `TranscodeToHLS` function: generates `.ts` segments and a master `.m3u8` playlist
  - Optional multi-resolution support (720p, 480p)
- Update `internal/svc-media/biz/upload.go`:
  - Background worker: refactor to include both `generateThumbnail` and `startTranscoding`
  - Status updates: `encoding_status` transitions from `processing` to `success`/`failed`
  - Store HLS master playlist path in `hls_file` field
- Update `internal/svc-media/data/media_repo.go`:
  - Add `hls_file` and `encoding_status` to the `Update` method

## Frontend Changes

### Enhanced Player & Status UI

- Update `VideoCard` component:
  - Check for `hls_file` availability, use `hls.js` for adaptive streaming
  - Show "Processing" badge when `encoding_status` is not `success`
- Add `hls.js` dependency

## Open Questions

- Should we implement a task queue (Redis/Asynq) now, or keep goroutines for the initial M3 phase?
- Which resolutions to prioritize? (single 720p stream or multi-resolution ladder?)

## Verification

1. Upload a new video
2. Observe "Processing" status on the video card
3. Once status changes to "Success", verify via network tab that `.ts` segments are being downloaded

---

*Last updated: 2026-04-03*
