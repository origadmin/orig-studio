# F001: Upload Page Refactor — SPEC

> beads: orig-cms-02q | type: feature | priority: P2

## Summary

Three changes to the upload flow:

1. Remove metadata form from upload page — move info entry to media edit
2. Extract full media info via ffprobe at upload completion
3. Promote file from temp to originals immediately after upload merge

## Background

原上传页面设计假设上传耗时较长，在等待期间让用户填写基础信息（标题、描述、分类、标签）。但实际上传很快，用户来不及操作表单。因此将基础信息填写移到媒体编辑页面。

上传完毕后当前只提取了 duration，缺少 width/height/codec/bitrate 等完整数据。应在合并完成时通过 ffprobe 一次性提取全部媒体信息。

文件路径存在临时性问题：上传完成后 media.url 指向 `temp/...`，直到转码成功才 Promote 到 `originals/...`。导致转码期间的源文件 URL 不可用。

## Goals

- Upload page: select files → auto-upload → done (zero user interaction)
- Media info: width, height, duration, size, extension, mime_type populated at upload
- File path: media.url always points to `originals/...` from creation moment

## Current vs Target Flow

### Current
```
Upload → Merge to temp/ → media.url = temp/... → (gap: file inaccessible)
→ TranscodeHandler reads temp/ → transcodes → Promote to originals/ → media.url = originals/...
```

### Target
```
Upload → Merge to temp/ → ffprobe extract all info → Promote to originals/
→ media.url = originals/... (file immediately accessible)
→ TranscodeHandler reads originals/ → transcodes (no promote needed)
```

## Changes

| # | Change | Files |
|---|--------|-------|
| 1 | Remove metadata form from upload page | web/src/components/upload/UploadComponent.tsx |
| 2 | Add GetMediaInfo() ffprobe parser | internal/helpers/ffmpeg/ffmpeg.go |
| 3 | ffprobe extraction + early Promote + Extension | internal/features/media/biz/upload.go |
| 4 | Remove stale promote logic | internal/features/media/biz/transcode_handler.go |
| 5 | Add tech info section to edit form | web/src/components/common/MediaEditForm.tsx |

## Implementation Status

- [x] ffmpeg.GetMediaInfo() added
- [x] CompleteMultipartUpload: ffprobe + early Promote + Extension/Width/Height
- [x] transcode_handler.go: promote logic removed
- [x] UploadComponent.tsx: metadata form removed, single-column layout
- [x] MediaEditForm.tsx: Technical Info section added
- [x] go build passes (ffmpeg + media/biz packages)
- [x] bun run build passes