# F001: Acceptance Criteria

## AC1: Upload page simplified

- [ ] Upload page shows only drag-drop zone + file list + progress
- [ ] No title/description/category/tags/cover form visible
- [ ] Files auto-upload immediately on selection
- [ ] Cancel button available at bottom

## AC2: Full media info populated

- [ ] Video upload: width, height from video stream filled in media record
- [ ] Audio upload: audio codec, channels, sample_rate extracted (logged)
- [ ] Image upload: ffprobe gracefully handles no video stream
- [ ] Extension field populated from file extension
- [ ] Duration extracted from format.duration (ffprobe)

## AC3: File path correct from creation

- [ ] media.url starts with "originals/" immediately after upload completion
- [ ] media.url NEVER contains "temp/" after media record creation
- [ ] Source file accessible via /uploads/ URL immediately
- [ ] Temp parts directory cleaned up after promotion

## AC4: Transcoding still works

- [ ] TranscodeHandler reads file from originals/ path
- [ ] HLS output generated correctly
- [ ] Master playlist generated
- [ ] Post-transcode sprite/gif processing still works

## AC5: Edit page shows tech info

- [ ] MediaEditForm shows Resolution, Duration, MIME Type, File Size, Extension
- [ ] Fields display "N/A" when value is zero/empty
- [ ] Tech info section is read-only