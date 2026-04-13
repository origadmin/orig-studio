# Media Base System Design

> Core media management: upload -> store -> metadata extraction -> thumbnail generation -> playback.

## 1. Overview

### 1.1 Module Scope

- Media upload (chunked, supports resume)
- Media listing (paginated, filterable by type/status/category)
- Media detail (metadata, variants, subtitles)
- Media playback (HLS stream URL / MP4 fallback)
- Media deletion (cascade delete files)

### 1.2 Out of Scope

- Media encoding/transcoding — handled in `media/transcode` module
- Media sharing — future feature
- Media playlist/collection — handled in `content` module

---

## 2. Architecture

### 2.1 Directory Structure

```
internal/svc-media/
├── biz/
│   ├── media.go           # MediaUseCase, Media entity
│   ├── upload.go          # Upload logic, chunk assembly
│   ├── storage.go         # StorageBackend interface
│   ├── interfaces.go      # MediaUseCaseInterface (for svc-content dependency)
│   ├── media_list_test.go
│   ├── upload_test.go
│   ├── transcode_handler.go
│   └── transcode_worker.go
├── data/
│   ├── media_repo.go      # ent-based MediaRepo implementation
│   ├── upload_repo.go     # UploadSession CRUD
│   ├── encoding_task_repo.go
│   ├── encode_profile_repo.go
│   └── seed.go            # Default encode profiles
├── dto/
│   └── media.go           # DTOs for API responses
├── server/
│   └── server.go          # Gin handlers
└── service/
    ├── media.go           # HTTP-to-biz conversion
    └── upload.go
```

### 2.2 Media Entity (current state)

```
Media (entity/internal/data/entity/media)
├── id              int64
├── uid             uuid
├── title           string
├── description     string
├── type            string     (video/audio/image)
├── state           string     (public/private/unlisted)
├── status          string     (pending/processing/success/failed)
├── file_path       string     (original file path)
├── size            int64
├── duration        int        (seconds, for video/audio)
├── width           int
├── height          int
├── thumbnail_path  string
├── user_id         int64      (FK)
├── category_id     *int64     (FK)
├── view_count      int64
├── like_count      int64
├── dislike_count   int64
├── favorite_count  int64
├── comment_count   int64
├── add_date        time
├── edit_date       time
└── status_reason   string     (error message if failed)

Edges:
- media.user        -> User (M2O)
- media.category     -> Category (M2O)
- media.tags        -> Tag[] (M2M)
```

### 2.3 Upload Flow

```
[Client]
  |  POST /api/v1/medias/upload/chunk (multipart/form-data, chunk + index + upload_id)
  v
[UploadHandler]
  |  1. Validate chunk (size, type)
  |  2. Store chunk to temp path: data/uploads/.chunks/{upload_id}/{index}
  |  3. Update/Insert UploadSession record
  v
[UploadRepo.CreateChunk() / UpdateChunk()]

[Client] <- 200 {chunk_index: 3, total_chunks: 10}

[Client] --- POST /api/v1/medias/upload/complete ---
[UploadHandler]
  |  1. Read all chunks from temp dir, concatenate
  |  2. Move assembled file to final path: data/uploads/{uid}.{ext}
  |  3. Extract metadata (ffprobe): duration, width, height
  |  4. Generate thumbnail (ffmpeg -i {file} -ss 00:00:01 -frames:v 1)
  |  5. Create Media record in DB (status=pending)
  |  6. Delete temp chunks dir
  v
[TranscodeHandler.Handle()]: Publish "media.encode.request" event
  |  -> TranscodeWorker.Submit() per profile
  |  -> ffmpeg TranscodeToHLS
  v
[Media status updated: pending -> processing -> success/failed]
```

### 2.4 API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/medias` | Optional | List medias |
| GET | `/api/v1/medias/{id}` | Optional | Get media detail |
| POST | `/api/v1/medias` | Required | Create media (metadata only) |
| PUT | `/api/v1/medias/{id}` | Required (owner/admin) | Update media |
| DELETE | `/api/v1/medias/{id}` | Required (owner/admin) | Delete media |
| POST | `/api/v1/medias/upload/init` | Required | Initialize chunked upload |
| POST | `/api/v1/medias/upload/chunk` | Required | Upload a chunk |
| POST | `/api/v1/medias/upload/complete` | Required | Complete upload, trigger encode |
| GET | `/api/v1/medias/{id}/stream` | Optional | Get HLS stream URL |
| GET | `/api/v1/medias/{id}/download` | Optional | Get download URL |
| GET | `/api/v1/medias/{id}/thumbnail` | Optional | Get thumbnail URL |
| POST | `/api/v1/medias/{id}/view` | Optional | Increment view count |

---

## 3. Known Issues

### 3.1 Bug: i18n Keys Not Resolving

**Location**: `web/src/pages/home/Watch.tsx`

**Problem**: Page displays raw i18n keys like `watch.subscribe`, `watch.like`, `watch.dislike` instead of translated strings.

**Root Cause**: i18n configuration not properly initialized, or i18n namespace files missing keys.

**Fix**:
1. Check `web/src/i18n/` for missing namespace files
2. Ensure `watch.*` namespace is loaded
3. Add missing keys to locale files

### 3.2 Missing Admin Actions on Watch Page

**Problem**: User's own media page lacks Edit Media / Edit Subtitles / Delete Media buttons.

**Fix**: Add action buttons in Watch.tsx for media owner and admin.

---

## 4. Acceptance Criteria

### 4.1 Upload

- [ ] Chunked upload works (resume after network failure)
- [ ] File type validation (video: mp4/mkv/webm, audio: mp3/flac, image: jpg/png/gif)
- [ ] File size limit enforced (configurable, default 2GB)
- [ ] Metadata extracted correctly (ffprobe)
- [ ] Thumbnail generated for video/audio

### 4.2 Playback

- [ ] HLS playback for transcoded video (hls.js)
- [ ] MP4 fallback for pending/failed transcoding
- [ ] Quality switcher works (360p/720p/1080p/AUTO)
- [ ] Safari native HLS fallback works

### 4.3 Frontend

- [ ] Watch page i18n resolves correctly (no raw keys)
- [ ] Own media page shows Edit/Delete actions
- [ ] Upload progress shown accurately
- [ ] Thumbnail displayed on media cards

---

*Last updated: 2026-04-13*
