# F001: R3 — API Contract

## New Function: ffmpeg.GetMediaInfo

```go
// Package: origadmin/application/origcms/internal/helpers/ffmpeg

func GetMediaInfo(ctx context.Context, inputPath string) (*MediaInfo, error)
```

**Command**: `ffprobe -v error -print_format json -show_format -show_streams {inputPath}`

**Behavior**: Parses JSON output into MediaInfo struct. Missing streams handled gracefully (zero values). Returns error only if ffprobe fails or JSON is unparseable.

**Example output structure**:
```json
{
  "streams": [
    {"codec_type": "video", "codec_name": "h264", "width": 1920, "height": 1080, "r_frame_rate": "30/1", "bit_rate": "3000000"},
    {"codec_type": "audio", "codec_name": "aac", "channels": 2, "sample_rate": "48000", "bit_rate": "128000"}
  ],
  "format": {"duration": "120.5", "size": "52428800", "bit_rate": "3480000", "format_name": "mov,mp4,m4a,3gp,3g2,mj2"}
}
```

## Modified: CompleteMultipartUpload

```go
// Package: origadmin/application/origcms/internal/features/media/biz

func (uc *UploadUseCase) CompleteMultipartUpload(
    ctx context.Context, uploadID string, sha256 string,
    title, description string, categoryID *int64, tags []string, thumbnail string,
) (*Media, error)
```

**New steps**:
1. After merge: call `ffmpeg.GetMediaInfo(ctx, fullPath)` → populate width/height/duration/extension
2. After CreateWithEntity: call `uc.paths.PromoteToOriginal(userID, filename)` → update media.Url
3. Encode request: `MediaPath: media.Url` (originals path)

**Error handling on promote failure**:
- Delete temp file
- Delete media record
- Return error

## Removed: TranscodeHandler promote block

```go
// Package: origadmin/application/origcms/internal/features/media/biz

// REMOVED: L477-L506 in transcode_handler.go
// if strings.HasPrefix(media.Url, "temp/") { ... PromoteToOriginal ... }
```

## Frontend: UploadComponent API

```tsx
// No API change — upload still uses same endpoints
// mediaApi.upload(file, metadata) with minimal metadata
// startMultipartUpload(task, callbacks) with empty metadata fields
```

**Simplified UploadTask**:
```tsx
interface UploadTask {
    id: string
    file: File
    progress: number
    status: UploadStatus
    parts: PartInfo[]
    uploadId?: string
    // REMOVED: title, description, categoryId, tags
}
```

## Frontend: MediaEditForm

**New displayed fields** (read-only):
- Resolution: `{width} x {height}` | "N/A"
- Duration: `formatDuration(seconds)` | "N/A"
- MIME Type: `{mime_type}` | "N/A"
- File Size: `formatFileSize(bytes)` | "N/A"
- Extension: `{extension}` | "N/A"

Uses existing `Media` interface fields: width, height, duration, size, mime_type, extension.