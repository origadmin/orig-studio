# F001: R2 — State Machine

## Upload Flow States

### Before (current)

```
[Files Selected]
    │
    ▼
[Uploading] ──metadata sync──► [CompleteMultipartUpload]
    │                               │
    │  temp/ parts                   ▼ merge to temp/
    │                            [media.url = temp/...]
    │                               │
    │                               ▼ publish encode request (MediaPath=temp/...)
    │                          [TranscodeHandler reads temp/]
    │                               │
    │                               ▼ transcode...
    │                          [Promote to originals/]
    │                               │
    │                               ▼
    │                          [media.url = originals/...] ✅
```

Problem: media.url = temp/... from upload completion to transcode finish. File inaccessible via /uploads/ during this window.

### After (target)

```
[Files Selected]
    │
    ▼
[Uploading]
    │
    ▼ temp/ parts
[CompleteMultipartUpload]
    │
    ├──► ffprobe GetMediaInfo() → fill width/height/duration/extension
    │
    ├──► Merge parts to temp/
    │
    ├──► Create media record (url = temp/...)
    │
    ├──► PromoteToOriginal() → move temp/ → originals/   ← NEW: immediate
    │
    ├──► Update media.url = originals/...                  ← NEW: correct from start
    │
    ├──► Delete temp parts
    │
    └──► publish encode request (MediaPath = originals/...) ← NEW: originals path
                    │
                    ▼
               [TranscodeHandler reads originals/] ← no promote needed
```

No more `temp/` exposure in media.url. File immediately accessible via /uploads/.

### TranscodeHandler: removed states

The following state transition is **removed**:

```
TranscodeHandler (old): [transcode success] → [check temp/ prefix] → [PromoteToOriginal] → [update media.url]
TranscodeHandler (new): [transcode success] → [skip, already in originals/]
```

## UploadSession status flow (unchanged)

```
pending → uploading → completed
                   → aborted
```